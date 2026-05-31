package article

import "rwa/internal/model"

type Repository interface {
	GetUser(email string) (model.Article, error)
	InsertUser(email string, article model.Article) (int, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) *service {
	return &service{repo: repo}
}

func (s *service) RegisterUser() {

}
