package handler

import (
	"backend-election/internal/dto"
	"backend-election/internal/model"
	"backend-election/internal/pkg/fingerprinting/extraction"
	"backend-election/internal/pkg/fingerprinting/helpers"
	"backend-election/internal/pkg/fingerprinting/matching"
	"backend-election/internal/pkg/fingerprinting/types"
	"backend-election/internal/pkg/httpresponse"
	"backend-election/internal/pkg/logger"
	"backend-election/internal/pkg/redis"
	"backend-election/internal/repository"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// Users handler
type Pemilihs struct {
	Log   *logger.Logger
	DB    *sql.DB
	Cache *redis.Cache
}

// @Security Bearer
// @Summary Get Pemilih by PemilihId
// @Description Get Pemilih by PemilihId
// @Tags Pemilih
// @Accept  json
// @Produce  json
// @Param Authorization header string true "
// @Param file formData file true "File to be uploaded"
// @Success 200 {object} dto.PemilihResponse
// @Router /pemilih [post]
func (h *Pemilihs) GetByPemilihId(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var ctx = r.Context()

	switch ctx.Err() {
	case context.Canceled:
		h.Log.Error(context.Canceled)
		http.Error(w, "Request is canceled", http.StatusExpectationFailed)
		return
	case context.DeadlineExceeded:
		h.Log.Error(context.DeadlineExceeded)
		http.Error(w, "Deadline is exceeded", http.StatusExpectationFailed)
		return
	default:
	}

	///////////////////////////////////////////

	if err := r.ParseMultipartForm(1024); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uploadedFile, handler, err := r.FormFile("file")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer uploadedFile.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer uploadedFile.Close()

	dir, err := os.Getwd()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := uuid.New().String() + filepath.Ext(handler.Filename)

	fileLocation := filepath.Join(dir, "files", filename)
	targetFile, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, uploadedFile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := os.Stat(fileLocation); os.IsNotExist(err) {
		log.Fatalf("File tidak ditemukan: %s, err: %v\n", fileLocation, err)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	_, m := helpers.LoadImage(fileLocation)
	result := extraction.DetectionResult(m)
	d, _ := json.Marshal(result)

	var signature []model.MatchingSignature
	var pemilihRepo = repository.PemilihRepository{Log: h.Log, Db: h.DB}
	pemilihList, err := pemilihRepo.FindAll(ctx)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Process each pemilih record to generate matching signatures.
	for _, pemilih := range pemilihList {

		log.Printf("Processing pemilih: %s", pemilih.PemilihID)

		var r1, r2 types.DetectionResult
		json.Unmarshal([]byte(d), &r1)
		json.Unmarshal([]byte(pemilih.SidikJari), &r2)
		max := len(r1.Minutia)
		if len(r1.Minutia) > len(r1.Minutia) {
			max = len(r2.Minutia)
		}
		matches := matching.Match(r1, r2)
		d3, _ := json.Marshal(matches)

		resultMatch := len(matches)

		log.Printf("matched minutiaes: %d/%d", len(matches), max)
		log.Printf("matches: %s", d3)

		if resultMatch != 0 {
			signature = append(signature, model.MatchingSignature{
				CodeID: pemilih.PemilihID,
				Nilai:  resultMatch,
			})
		}
	}

	if len(signature) == 0 {
		http.Error(w, "No matching signature found", http.StatusNotFound)
		return
	}

	// Mencari nilai terbesar dari signature
	maxSignature := signature[0]
	for _, sig := range signature {
		if sig.Nilai > maxSignature.Nilai {
			maxSignature = sig
		}
	}

	// Melakukan pencarian di database (placeholder)
	// Gantilah dengan logika pencarian database yang sesuai
	var pemilihRepoOne = repository.PemilihRepository{Log: h.Log, Db: h.DB}
	pemilihRepoOne.PemilihEntity = model.Pemilih{PemilihID: string(maxSignature.CodeID)}
	err = pemilihRepoOne.Find(ctx)

	log.Printf("Searching for pemilih with PemilihID: %s", maxSignature.CodeID)

	httpres := httpresponse.Response{Cache: h.Cache}

	key := fmt.Sprintf("users.%d")
	if cacheValue, isExist := h.Cache.Get(ctx, key); isExist {
		httpres.Set(w, http.StatusOK, cacheValue)
		return
	}

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		fmt.Printf("Found matching signature with CodeId: %s and Nilai: %v\n", maxSignature.CodeID, pemilihRepoOne) // Assuming Nilai is a field in the Pemilih struct
	}

	var response dto.PemilihResponse
	response.FromEntity(pemilihRepoOne.PemilihEntity)
	httpres.SetMarshal(ctx, w, http.StatusOK, response, key)
}
