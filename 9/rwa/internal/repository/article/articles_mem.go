package article

import (
	"context"
	"rwa/internal/model"
	"slices"
	"sync"
	"time"
)

// TODO:

type memRepo struct {
	mu sync.RWMutex
	db []model.Article
}

func NewMemRepo() *memRepo {
	return &memRepo{
		db: []model.Article{},
	}
}

func (r *memRepo) GetArticle(ctx context.Context, username string) (model.Article, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, article := range r.db {
		if article.Author.Username == username {
			return article, true
		}
	}

	return model.Article{}, false

}

func (r *memRepo) InsertArticle(ctx context.Context, article model.Article) model.Article {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	article.CreatedAt = now
	article.UpdatedAt = now

	r.db = append(r.db, article)

	return article
}

func (r *memRepo) GetAllArticles(ctx context.Context) []model.Article {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.db

}

func (r *memRepo) GetByAuthor(ctx context.Context, username string) []model.Article {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var byAuthor []model.Article
	for _, article := range r.db {
		if article.Author.Username == username {
			byAuthor = append(byAuthor, article)
		}
	}

	return byAuthor

}

func (r *memRepo) GetByTag(ctx context.Context, tag string) []model.Article {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var byAuthor []model.Article

	for _, article := range r.db {

		if slices.Contains(article.TagList, tag) {
			byAuthor = append(byAuthor, article)
		}
	}

	return byAuthor

}
