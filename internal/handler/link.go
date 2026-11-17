package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/kirillat6/urlShorter/internal/errs"
	"github.com/kirillat6/urlShorter/internal/service"
)

type LinkReq struct {
	URL string `json:"url"`
}

type LinkResp struct {
	ShortURL string `json:"short_url"`
}

func CreateLink(
	w http.ResponseWriter,
	r *http.Request,
	linkService *service.LinkService,
) {
	var req LinkReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	normalizedURL, err := NormalizeURL(req.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf("ссылка не прошла валидацию: %v", err), http.StatusBadRequest)
		return
	}

	linkCode, err := linkService.CreateLink(r.Context(), normalizedURL)
	if err != nil {
		log.Printf("Ошибка создания короткой ссылки: %v", err)
		http.Error(w, "Не удалось сохранить ссылку", http.StatusInternalServerError)
		return
	}

	resp := LinkResp{
		ShortURL: "http://localhost:8080/" + linkCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Ошибка кодирования ответа: %v", err)
		return
	}
}

func RedirectLink(
	w http.ResponseWriter,
	r *http.Request,
	linkService *service.LinkService,
) {
	code := r.PathValue("code")
	url, err := linkService.GetOriginalURL(r.Context(), code)
	if errors.Is(err, errs.ErrLinkNotFound) {
		http.NotFound(w, r)
		return
	}

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func NormalizeURL(rawURL string) (string, error) {
	normalizedURL := strings.TrimSpace(rawURL)

	if normalizedURL == "" {
		return "", fmt.Errorf("URL не может быть пустым!")
	}
	if !strings.Contains(normalizedURL, "://") {
		normalizedURL = "https://" + normalizedURL

	}

	parsedURL, err := url.Parse(normalizedURL)
	if err != nil {
		return "", fmt.Errorf("некорректный URL: %w", err)
	}

	if parsedURL.Hostname() == "" ||
		(parsedURL.Scheme != "http" &&
			parsedURL.Scheme != "https") {
		return "", fmt.Errorf("у URL должно присутствовать имя хоста и схема")
	}
	return normalizedURL, nil
}
