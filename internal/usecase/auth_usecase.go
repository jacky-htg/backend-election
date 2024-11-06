package usecase

import (
	"backend-election/internal/dto"
	"backend-election/internal/model"
	"backend-election/internal/pkg/jwttoken"
	"backend-election/internal/pkg/logger"
	"backend-election/internal/repository"
	"context"
	"database/sql"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type AuthUC struct {
	Log *logger.Logger
	DB  *sql.DB
}

func (uc AuthUC) Login(ctx context.Context, loginRequest dto.LoginRequest) (string, int, error) {
	switch ctx.Err() {
	case context.Canceled:
		return "", http.StatusInternalServerError, uc.Log.Error(context.Canceled)
	case context.DeadlineExceeded:
		return "", http.StatusInternalServerError, uc.Log.Error(context.DeadlineExceeded)
	default:
	}

	userRepo := repository.UserRepository{Log: uc.Log, Db: uc.DB, UserEntity: model.User{Email: loginRequest.Email}}
	if err := userRepo.GetByEmail(ctx); err != nil {
		return "", http.StatusInternalServerError, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(strings.TrimSpace(userRepo.UserEntity.Password)), []byte(loginRequest.Password)); err != nil {
		return "", http.StatusUnauthorized, uc.Log.Error(err)
	}

	token, err := jwttoken.ClaimToken(userRepo.UserEntity.Email)
	if err != nil {
		return "", http.StatusInternalServerError, uc.Log.Error(err)
	}

	return token, http.StatusOK, nil
}
