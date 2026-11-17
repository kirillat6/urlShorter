package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"

	"github.com/kirillat6/urlShorter/internal/errs"
	repo "github.com/kirillat6/urlShorter/internal/repository"
)

type LinkService struct {
	repo *repo.LinkRepository
}

func NewLinkService(linkRepo *repo.LinkRepository) *LinkService {
	return &LinkService{
		repo: linkRepo,
	}
}

func (s *LinkService) CreateLink(
	ctx context.Context,
	originalURL string,
) (string, error) {

	for i := 0; i < 5; i++ {
		var code string
		var err error
		code, err = GenerateCode()

		if err != nil {
			return "", fmt.Errorf("Проблемы с генерацией кода на уровне сервиса: %w", err)
		}
		err = s.repo.CreateLink(ctx, code, originalURL)
		if errors.Is(err, errs.ErrCodeAlreadyExists) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("Проблемы с созданием ссылки на уровне сервиса: %w", err)
		}
		return code, nil
	}
	return "", fmt.Errorf("не удалось создать уникальный код за 5 попыток: %w", errs.ErrCodeAlreadyExists)
}

func GenerateCode() (string, error) {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *LinkService) GetOriginalURL(
	ctx context.Context,
	code string,
) (string, error) {
	originalURL, err := s.repo.GetLink(ctx, code)
	if err != nil {
		return "", fmt.Errorf("не получилось вернуть оригинальную ссылку с кодом: %q: %w", code, err)
	}
	if err = s.repo.IncrementClickCount(ctx, code); err != nil {
		log.Printf("не получилось увеличить счётчик переходов для кода: %q: %v", code, err)
	}

	return originalURL, nil
}
