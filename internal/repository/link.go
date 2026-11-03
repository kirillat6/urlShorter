package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LinkRepository struct {
	pool *pgxpool.Pool
}

func NewLinkRepository(pool *pgxpool.Pool) *LinkRepository {
	return &LinkRepository{
		pool: pool,
	}
}

func (r *LinkRepository) CreateLink(
	ctx context.Context,
	code string,
	originalURL string,
) error {
	sql := `
		INSERT INTO links (code, original_url)
		VALUES ($1, $2);
	`
	_, err := r.pool.Exec(ctx, sql, code, originalURL)
	if err != nil {
		return fmt.Errorf("Произошла ошибка при добавлении ссылки: %w", err)
	}
	return nil
}

func (r *LinkRepository) GetLink(ctx context.Context, code string) (*LinkReq, error) {
	sql := `
		SELECT original_url
			FROM links
		WHERE code = $1;
	`
	var originalURL string

	err := r.pool.QueryRow(ctx, sql, code).Scan(&originalURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("не было найдено ни одной строки: %w", err)
	}
	if err != nil {
		return "", fmt.Errorf("ошибка получения оригинальной ссылки: %w", err)
	}

	return originalURL, nil
}
