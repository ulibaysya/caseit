package app

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type UserHandler interface {
	Create(http.ResponseWriter, *http.Request)
}

type AuthKeyHandler interface {
	Create(http.ResponseWriter, *http.Request)
	// Exchange(http.ResponseWriter, *http.Request)
}

func NewHandler(logger *slog.Logger, userHandler UserHandler, authKeyHandler AuthKeyHandler) http.Handler {
	mux := chi.NewMux()

	mux.HandleFunc("POST /users", userHandler.Create)

	mux.HandleFunc("POST /auth/keys", authKeyHandler.Create)
	// mux.HandleFunc("POST /auth/keys", authKeyHandler.Exchange)

	return mux
}
