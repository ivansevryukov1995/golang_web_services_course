package user

import (
	"context"
	"rwa/internal/model"
	"time"
)

type Repository interface {
	GetByEmail(ctx context.Context, email string) (model.User, error)
	InsertUser(ctx context.Context, user model.User) (model.User, error)
	UpdateUser(ctx context.Context, email string, user model.User) (model.User, error)
}

type UserInput struct {
	Email    string
	Password string
	Username string
	Bio      string
}

type service struct {
	repo Repository
}

func NewService(repo Repository) *service {
	return &service{repo: repo}
}

func (s *service) RegisterUser(ctx context.Context, in UserInput) (model.User, error) {

	now := time.Now().UTC().Format(time.RFC3339Nano)
	user := model.User{
		Email:     in.Email,
		Password:  in.Password,
		Username:  in.Username,
		CreatedAt: now,
		UpdatedAt: now,
	}

	u, err := s.repo.InsertUser(ctx, user)
	if err != nil && err.Error() == "looks like user exists" {
		return model.User{}, err
	}
	return u, nil

}

func (s *service) LoginUser(ctx context.Context, in UserInput) (model.User, error) {

	u, err := s.repo.GetByEmail(ctx, in.Email)
	if err != nil && err.Error() == "user not found" {
		return model.User{}, err
	}

	if in.Password != u.Password {
		return model.User{}, nil
	}

	u.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	return u, nil
}

func (s *service) CurrentUser(ctx context.Context, email string) (model.User, error) {

	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil && err.Error() == "user not found" {
		return model.User{}, err
	}

	u.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	return u, nil
}

func (s *service) UpdateUser(ctx context.Context, email string, in UserInput) (model.User, error) {

	now := time.Now().UTC().Format(time.RFC3339Nano)
	user := model.User{
		Email:     in.Email,
		Username:  in.Username,
		Bio:       in.Bio,
		UpdatedAt: now,
	}

	u, err := s.repo.UpdateUser(ctx, email, user)
	if err != nil && err.Error() == "user not found" {
		return model.User{}, err
	}

	return u, nil
}
