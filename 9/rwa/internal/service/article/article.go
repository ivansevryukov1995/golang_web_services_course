package article

import (
	"context"
	"errors"
	"rwa/internal/model"
)

type ArticleRepository interface {
	GetArticle(ctx context.Context, username string) (model.Article, bool)
	InsertArticle(ctx context.Context, article model.Article) model.Article
	GetAllArticles(ctx context.Context) []model.Article
	GetByAuthor(ctx context.Context, username string) []model.Article
	GetByTag(ctx context.Context, tag string) []model.Article
}

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (model.User, error)
	InsertUser(ctx context.Context, user model.User) (model.User, error)
	UpdateUser(ctx context.Context, id string, user model.User) (model.User, error)
}

type service struct {
	articleRepo ArticleRepository
	userRepo    UserRepository
}

func NewService(articleRepo ArticleRepository, userRepo UserRepository) *service {
	return &service{
		articleRepo: articleRepo,
		userRepo:    userRepo,
	}
}

func (s *service) CreateArticle(ctx context.Context, email string, article model.Article) (model.Article, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil && errors.Is(err, errors.New("user not found")) {
		return model.Article{}, err
	}

	article.Author = user

	articleNew := s.articleRepo.InsertArticle(ctx, article)
	return articleNew, nil
}

func (s *service) GetAllArticle(ctx context.Context) []model.Article {
	return s.articleRepo.GetAllArticles(ctx)
}

func (s *service) GetByAuthor(ctx context.Context, username string) []model.Article {
	articles := s.articleRepo.GetByAuthor(ctx, username)
	return articles
}

func (s *service) GetByTag(ctx context.Context, tag string) []model.Article {
	articles := s.articleRepo.GetByTag(ctx, tag)
	return articles
}
