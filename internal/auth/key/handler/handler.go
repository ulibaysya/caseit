package handler

import (
	"context"
	"log/slog"

	// "github.com/ulibaysya/caseit/internal/auth"
	"github.com/ulibaysya/caseit/internal/user"
)

type Service interface {
	Create(ctx context.Context, userID user.ID) (authKey string, err error)
	// Exchange(ctx context.Context, key string) (tokens auth.Tokens, err error)
}

type Handler struct {
	logger  *slog.Logger
	service Service
}

func New(logger *slog.Logger, service Service) Handler {
	return Handler{
		logger:  logger,
		service: service,
	}
}
