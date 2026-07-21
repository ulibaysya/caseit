package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	pkghttp "github.com/ulibaysya/caseit/pkg/http"
	pkgjson "github.com/ulibaysya/caseit/pkg/json"
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

	parameters, err := pkgjson.Read[createParameters](r.Body)
	if err != nil {
		if errors.Is(err, io.EOF) {
			h.logger.LogAttrs(ctx, slog.LevelWarn, "got empty body")
			pkghttp.ErrorJSON(w, "empty body", http.StatusBadRequest)
		} else if errors.Is(err, pkgjson.ErrInvalid){
			h.logger.LogAttrs(ctx, slog.LevelWarn, "got invalid json", slog.String("error", err.Error()))
			pkghttp.ErrorJSON(w, "invalid json", http.StatusBadRequest)
		} else {
			h.logger.LogAttrs(ctx, slog.LevelError, "decoding request body", slog.String("error", err.Error()))
			pkghttp.ErrorJSON500(w)
		}
		return
	}

	if err = parameters.Validate(); err != nil {
		errStr := err.Error()
		h.logger.LogAttrs(ctx, slog.LevelWarn, "validating request parameters", slog.String("error", errStr))
		pkghttp.ErrorJSON(w, errStr, http.StatusBadRequest)
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

	if err = pkgjson.Write(w, id); err != nil {
		h.logger.LogAttrs(ctx, slog.LevelError, "writing response", slog.String("error", err.Error()))
		return
	}
}

type createParameters struct {
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

func (p createParameters) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("empty name")
	}
	if p.ImageURL != "" {
		imageURL, err := url.ParseRequestURI(p.ImageURL)
		if err != nil {
			return fmt.Errorf("image_url is not an url")
		}
		if imageURL.Scheme != "http" && imageURL.Scheme != "https" {
			return fmt.Errorf("image_url's scheme is not an http/https")
		}
		if imageURL.OmitHost || imageURL.Host == "" {
			return fmt.Errorf("image_url's host is empty")
		}
	}
	return nil
}
