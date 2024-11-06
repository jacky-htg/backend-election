package tests

import (
	"backend-election/internal/handler"
	"backend-election/internal/middleware"
	"backend-election/internal/pkg/config"
	"backend-election/internal/pkg/logger"
	"backend-election/internal/pkg/redis"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/julienschmidt/httprouter"
)

var (
	db                 *sql.DB
	cache              *redis.Cache
	done               func()
	log                *logger.Logger
	token              string
	publicMiddlewares  []func(httprouter.Handle) httprouter.Handle
	privateMiddlewares []func(httprouter.Handle) httprouter.Handle
	mid                middleware.Middleware
)

func TestMain(m *testing.M) {
	var err error

	log = logger.New()
	if _, ok := os.LookupEnv("APP_NAME"); !ok {
		if err := config.Setup("../.env"); err != nil {
			fmt.Println("failed to setup config", err)
			return
		}
	}

	rootPath, err := filepath.Abs("../") // Path relatif ke folder root proyek
	if err != nil {
		fmt.Println("failed to find root directory", err)
		return
	}

	// Ubah working directory ke root proyek
	err = os.Chdir(rootPath)
	if err != nil {
		fmt.Println("failed to change working directory", err)
		return
	}

	db, cache, done = NewUnit()
	defer done()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:")
		}
		done()
	}()

	mid = middleware.Middleware{Log: log, DB: db, Cache: cache}
	publicMiddlewares = []func(httprouter.Handle) httprouter.Handle{
		mid.CORS,
		mid.PanicRecovery,
		mid.Semaphore,
		mid.RateLimit,
		mid.Idempotency,
	}
	privateMiddlewares = append(publicMiddlewares, mid.Authentication, mid.Authorization)

	err = login(log, db)
	if err != nil {
		fmt.Println("failed to login", err)
		return
	}

	code := m.Run()
	if code != 0 {
		fmt.Println("Some tests failed. Check the output above for details.")
	}
	done()

	os.Exit(code)
}

func login(log *logger.Logger, db *sql.DB) error {
	loginData := map[string]string{
		"email":    "rijal.asep.nugroho@gmail.com",
		"password": "qwertyuiop!1Q",
	}

	dataJSON, err := json.Marshal(loginData)
	if err != nil {
		return fmt.Errorf("could not marshal login data: %v", err)
	}

	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(dataJSON))
	if err != nil {
		return fmt.Errorf("could not create login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "6b34d72b-2b1d-42ab-ad43-89ecd9312441")

	rr := httptest.NewRecorder()
	router := httprouter.New()
	authHandler := handler.Auths{DB: db, Log: log}
	router.POST("/login", mid.WrapMiddleware(publicMiddlewares, authHandler.Login))

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		return fmt.Errorf("login handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		return fmt.Errorf("could not unmarshal login response: %v", err)
	}

	token = response["token"].(string)
	return nil
}
