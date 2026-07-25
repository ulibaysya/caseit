package key

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ulibaysya/caseit/internal/user"
)

type storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *storage {
	return &storage{
		pool: pool,
	}
}

// TODO test
func (s storage) Save(ctx context.Context, authKey string, userID user.ID) error {
	const sql = `INSERT INTO keys (key, users_id) VALUES ($1, $2);`

	_, err := s.pool.Exec(ctx, sql, authKey, userID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == pgerrcode.ForeignKeyViolation {
				return fmt.Errorf("user %v: %w", userID, user.ErrNotFound)
			} else {
				return fmt.Errorf("postgres error: %w", pgErr)
			}
		} else {
			return err
		}
	}

	return nil
}
