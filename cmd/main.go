package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

type LinkReq struct {
	URL string `json:"url"`
}

type LinkResp struct {
	ShortURL string `json:"short_url"`
}

type LinkStorage struct {
	mu    sync.RWMutex
	links map[string]string
}

func NewLinkStorage() *LinkStorage {
	return &LinkStorage{
		links: make(map[string]string),
	}
}

func main() {
	storage := NewLinkStorage()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		CreateLink(w, r, storage)
	})
	mux.HandleFunc("GET /{code}", func(w http.ResponseWriter, r *http.Request) {
		RedirectLink(w, r, storage)
	})

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal("Проблема с сервером!")
	}

}

func CreateLink(w http.ResponseWriter, r *http.Request, storage *LinkStorage) {
	var req LinkReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.URL) == "" {
		http.Error(w, "URL не может быть пустым!", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(req.URL, "https://") &&
		!strings.HasPrefix(req.URL, "http://") {
		http.Error(w, "URL должен начинаться с http или https!", http.StatusBadRequest)
		return
	}

	code, err := GenerateCode()
	if err != nil {
		http.Error(w, "Проблемы с генерацией кода", http.StatusInternalServerError)
		return
	}

	resp := LinkResp{
		ShortURL: "http://localhost:8080/" + code,
	}

	res := storage.SaveLink(code, req.URL)
	if !res {
		http.Error(w, "Не удалось сохранить ссылку", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Ошибка кодирования ответа: %v", err)
		return
	}
}

func GenerateCode() (string, error) {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *LinkStorage) FindLink(code string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	link, ok := s.links[code]
	return link, ok
}

func (s *LinkStorage) SaveLink(code string, originalURL string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.links[code]
	if ok {
		return false
	}

	s.links[code] = originalURL
	return true
}

func RedirectLink(
	w http.ResponseWriter,
	r *http.Request,
	storage *LinkStorage,
) {
	code := r.PathValue("code")
	url, exist := storage.FindLink(code)
	if !exist {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}
