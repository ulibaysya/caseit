package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	pkghttp "github.com/ulibaysya/caseit/pkg/http"
)

type UserCreatorService interface {
	Create(ctx context.Context, name string, imageURL string) (id int64, err error)
}

type CreateHandler struct {
	logger      *slog.Logger
	userService UserCreatorService
}

func NewCreateHandler(logger *slog.Logger, userService UserCreatorService) CreateHandler {
	return CreateHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	parameters := struct {
		Name     string `json:"name"`
		ImageURL string `json:"image_url"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&parameters)
	if err != nil {
		if errors.Is(err, io.EOF) {
			h.logger.LogAttrs(ctx, slog.LevelWarn, "got empty body")
			pkghttp.ErrorJSON(w, "empty body", http.StatusBadRequest)
		} else if _, ok := errors.AsType[*json.SyntaxError](err); ok {
			h.logger.LogAttrs(ctx, slog.LevelWarn, "got invalid json", slog.String("error", err.Error()))
			pkghttp.ErrorJSON(w, "invalid json", http.StatusBadRequest)
		} else {
			h.logger.LogAttrs(ctx, slog.LevelError, "decoding request body", slog.String("error", err.Error()))
			pkghttp.ErrorJSON500(w)
		}
		return
	}

	if parameters.Name == "" {
		h.logger.LogAttrs(ctx, slog.LevelWarn, "request with empty name")
		pkghttp.ErrorJSON(w, "empty name", http.StatusBadRequest)
		return
	}
	if parameters.ImageURL == "" {
		h.logger.LogAttrs(ctx, slog.LevelWarn, "request with empty image_url")
		pkghttp.ErrorJSON(w, "empty image_url", http.StatusBadRequest)
		return
	}

	id, err := h.userService.Create(ctx, parameters.Name, parameters.ImageURL)
	if err != nil {
		h.logger.LogAttrs(ctx, slog.LevelError, "creating user", slog.String("error", err.Error()))
		pkghttp.ErrorJSON500(w)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/users/%d", id))

	w.WriteHeader(http.StatusCreated)

	if err = pkghttp.WriteJSON(w, id); err != nil {
		h.logger.LogAttrs(ctx, slog.LevelError, "writing response", slog.String("error", err.Error()))
		pkghttp.ErrorJSON500(w)
		return
	}
}
