package key

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ulibaysya/caseit/internal/user"
)

type Storage interface {
	Save(ctx context.Context, authKey string, userID user.ID) error
}

type KeyGenerator interface {
	Generate() string
}

type service struct {
	logger  *slog.Logger
	storage Storage
	kg      KeyGenerator
}

func NewService(logger *slog.Logger, storage Storage, keyGenerator KeyGenerator) service {
	return service{
		logger:  logger,
		storage: storage,
		kg:      keyGenerator,
	}
}

func (s service) Create(ctx context.Context, userID user.ID) (string, error) {
	authKey := s.kg.Generate()

	err := s.storage.Save(ctx, authKey, userID)
	if err != nil {
		return "", fmt.Errorf("saving authentication key: %w", err)
	}

	s.logger.LogAttrs(ctx, slog.LevelInfo, "created authentication key", slog.Int64("user_id", int64(userID)))

	return authKey, nil
}
