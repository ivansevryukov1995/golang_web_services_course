package user

import (
	"context"
	"errors"
	"rwa/internal/model"
	"time"
)

type Repository interface {
	GetByEmail(ctx context.Context, email string) (model.User, error)
	InsertUser(ctx context.Context, user model.User) (model.User, error)
	UpdateUser(ctx context.Context, id string, user model.User) (model.User, error)
}

type ServiceSession interface {
	Check(ctx context.Context, token string) (model.Session, error)
	Create(ctx context.Context, userID string, email string) (string, error)
}

type UserInput struct {
	Email    string
	Password string
	Username string
	Bio      string
}

type service struct {
	repo    Repository
	session ServiceSession
}

func NewService(repo Repository, session ServiceSession) *service {
	return &service{repo: repo,
		session: session,
	}
}

func (s *service) RegisterUser(ctx context.Context, in UserInput) (model.User, string, error) {
	user := model.User{
		Email:    in.Email,
		Password: in.Password,
		Username: in.Username,
	}

	newUser, err := s.repo.InsertUser(ctx, user)
	if err != nil && errors.Is(err, errors.New("looks like user exists")) {
		return model.User{}, "", err
	}

	token, _ := s.session.Create(ctx, newUser.ID, newUser.Email)

	return newUser, token, nil
}

func (s *service) LoginUser(ctx context.Context, in UserInput) (model.User, string, error) {

	user, err := s.repo.GetByEmail(ctx, in.Email)
	if err != nil && errors.Is(err, errors.New("user not found")) {
		return model.User{}, "", err
	}

	if in.Password != user.Password {
		return model.User{}, "", nil
	}

	token, _ := s.session.Create(ctx, user.ID, user.Email)

	return user, token, nil
}

func (s *service) CurrentUser(ctx context.Context, email string) (model.User, error) {

	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil && errors.Is(err, errors.New("user not found")) {
		return model.User{}, err
	}

	u.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	return u, nil
}

func (s *service) UpdateUser(ctx context.Context, id string, in UserInput) (model.User, string, error) {
	u := model.User{
		Email: in.Email,
		Bio:   in.Bio,
	}

	user, err := s.repo.UpdateUser(ctx, id, u)
	if err != nil && errors.Is(err, errors.New("user not found")) {
		return model.User{}, "", err
	}

	token, _ := s.session.Create(ctx, id, user.Email)

	return user, token, nil
}
