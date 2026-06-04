package session

import (
	"context"
	"rwa/internal/model"
)

type Repository interface {
	Get(ctx context.Context, token string) (model.Session, error)
	Create(ctx context.Context, userID string, email string) (string, error)
	DestroyCurrent(ctx context.Context, token string)
	// DestroyAll(http.ResponseWriter, *User) error
}

type SessionInput struct {
	Email string
	Token string
}

type service struct {
	repo Repository
}

func NewService(repo Repository) *service {
	return &service{repo: repo}
}

func (s *service) Check(ctx context.Context, token string) (model.Session, error) {
	sess, err := s.repo.Get(ctx, token)
	if err != nil {
		return model.Session{}, err
	}

	return sess, nil
}

func (s *service) Create(ctx context.Context, userID string, email string) (string, error) {
	token, _ := s.repo.Create(ctx, userID, email)

	return token, nil
}

func (s *service) DestroyCurrent(ctx context.Context, token string) {

	s.repo.DestroyCurrent(ctx, token)

}
