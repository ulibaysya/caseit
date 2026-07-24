package key

import (
	"context"
	"fmt"

	"github.com/ulibaysya/caseit/internal/user"
)

type Storage interface {
	Save(ctx context.Context, authKey string, userID user.ID) error
}

type KeyGenerator interface {
	Generate() string
}

type service struct {
	storage Storage
	kg      KeyGenerator
}

func NewService(storage Storage, keyGenerator KeyGenerator) service {
	return service{
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

	return authKey, nil
}
