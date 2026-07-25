package key

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/ulibaysya/caseit/internal/user"
	pkghttp "github.com/ulibaysya/caseit/pkg/http"
	pkgjson "github.com/ulibaysya/caseit/pkg/json"
)

type Service interface {
	Create(ctx context.Context, userID user.ID) (authKey string, err error)
	// Exchange(ctx context.Context, key string) (tokens auth.Tokens, err error)
}

type handler struct {
	logger  *slog.Logger
	service Service
}

func NewHandler(logger *slog.Logger, service Service) handler {
	return handler{
		logger:  logger,
		service: service,
	}
}

// TODO fix case when received "{}". json just sets id to 0 and tells nothing that field is omitted
func (h handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO abstract decoding json request error checking
	request, err := pkgjson.Read[creationRequest](r.Body, true)
	if err != nil {
		if errors.Is(err, io.EOF) {
			h.logger.LogAttrs(ctx, slog.LevelWarn, "got empty body")
			pkghttp.ErrorJSON(w, "empty body", http.StatusBadRequest)
		} else if _, ok := errors.AsType[*json.SyntaxError](err); ok {
			h.logger.LogAttrs(ctx, slog.LevelWarn, "got invalid json", slog.String("error", err.Error()))
			pkghttp.ErrorJSON(w, "invalid json", http.StatusBadRequest)
		} else if _, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
			h.logger.LogAttrs(ctx, slog.LevelWarn, "inappropriate request", slog.String("error", err.Error()))
			pkghttp.ErrorJSON(w, "inappropriate request", http.StatusBadRequest)
		} else if fieldErr, ok := errors.AsType[*pkgjson.UnknownFieldErr](err); ok {
			h.logger.LogAttrs(ctx, slog.LevelWarn, "unknown json field", slog.String("field", fieldErr.Field))
			pkghttp.ErrorJSON(w, "inappropriate request", http.StatusBadRequest)
		} else {
			h.logger.LogAttrs(ctx, slog.LevelError, "decoding request body", slog.String("error", err.Error()))
			pkghttp.ErrorJSON500(w)
		}
		return
	}

	if err = request.Validate(); err != nil {
		errStr := err.Error()
		h.logger.LogAttrs(ctx, slog.LevelWarn, "validating request parameters", slog.String("error", errStr))
		pkghttp.ErrorJSON(w, errStr, http.StatusBadRequest)
		return
	}

	authKey, err := h.service.Create(ctx, *request.UserID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			h.logger.LogAttrs(ctx, slog.LevelWarn, "non-existent user is triggered", slog.Int64("id", int64(*request.UserID)))
			pkghttp.ErrorJSON(w, err.Error(), http.StatusNotFound)
		} else {
			h.logger.LogAttrs(ctx, slog.LevelError, "creating auth key", slog.String("error", err.Error()))
			pkghttp.ErrorJSON500(w)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err = pkgjson.Write(w, authKey); err != nil {
		h.logger.LogAttrs(ctx, slog.LevelError, "writing response", slog.String("error", err.Error()))
		return
	}
}

type creationRequest struct {
	UserID *user.ID `json:"user_id"`
}

func (r creationRequest) Validate() error {
	if r.UserID == nil {
		return fmt.Errorf("user_id is omitted")
	}
	err := r.UserID.Validate()
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}
	return nil
}
