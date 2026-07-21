package postgres

import (
	"context"
	"fmt"
)

func (s Storage) Create(ctx context.Context, name string, imageURL string) (int64, error) {
	const sql = "INSERT INTO users (name, image_url) VALUES ($1, $2) RETURNING id;"

	var id int64
	err := s.pool.QueryRow(ctx, sql, name, imageURL).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("inserting into users: %w", err)
	}

	return id, nil
}
