package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kirillat6/urlShorter/internal/errs"
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

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return errs.ErrCodeAlreadyExists
	}

	if err != nil {
		return fmt.Errorf("Произошла ошибка при добавлении ссылки: %w", err)
	}
	return nil
}

func (r *LinkRepository) GetLink(ctx context.Context, code string) (string, error) {
	sql := `
		SELECT original_url
			FROM links
		WHERE code = $1;
	`
	var originalURL string

	err := r.pool.QueryRow(ctx, sql, code).Scan(&originalURL)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", errs.ErrLinkNotFound
	}

	if err != nil {
		return "", fmt.Errorf("ошибка получения ссылки: %w", err)
	}

	return originalURL, nil
}

func (r *LinkRepository) IncrementClickCount(
	ctx context.Context,
	code string,
) error {
	query := `
	UPDATE links
	SET click_count = click_count + 1
	WHERE code = $1
	`

	tag, err := r.pool.Exec(ctx, query, code)
	if err != nil {
		return fmt.Errorf("не удалось прибавить клик: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("ссылка с кодом: %q не найдена: %w", code, errs.ErrLinkNotFound)
	}
	return nil
}
