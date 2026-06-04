package article

import (
	"context"
	"encoding/json"
	"net/http"
	"rwa/internal/handler/article/dto"
	auth "rwa/internal/middleware"
	"rwa/internal/model"
)

type Service interface {
	CreateArticle(ctx context.Context, email string, article model.Article) (model.Article, error)
	GetAllArticle(ctx context.Context) []model.Article
	GetByAuthor(ctx context.Context, username string) []model.Article
	GetByTag(ctx context.Context, tag string) []model.Article
}

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
}

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	sess, err := auth.SessionFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()

	req := dto.ArticleRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	article, err := h.service.CreateArticle(r.Context(), sess.Email, dto.ToArticleInput(req))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(dto.ToArticleResponse(article))
}

func (h *handler) GetAll(w http.ResponseWriter, r *http.Request) {

	articles := h.service.GetAllArticle(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(dto.ToArticlesResponse(articles))
}

func (h *handler) GetByAuthor(w http.ResponseWriter, r *http.Request) {
	_, err := auth.SessionFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	author := r.URL.Query().Get("author")

	articles := h.service.GetByAuthor(r.Context(), author)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(dto.ToArticlesResponse(articles))
}

func (h *handler) GetByTag(w http.ResponseWriter, r *http.Request) {
	_, err := auth.SessionFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tag := r.URL.Query().Get("tag")

	articles := h.service.GetByTag(r.Context(), tag)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(dto.ToArticlesResponse(articles))
}
