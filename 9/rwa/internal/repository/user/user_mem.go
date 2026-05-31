package user

import (
	"context"
	"errors"
	"rwa/internal/model"
	"strconv"
	"sync"
)

// TODO:

type memRepo struct {
	mu     sync.RWMutex
	data   map[string]model.User
	nextID int
}

func NewMemRepo() *memRepo {
	return &memRepo{
		data:   make(map[string]model.User),
		nextID: 1,
	}
}

func (r *memRepo) GetByEmail(ctx context.Context, email string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exist := r.data[email]
	if !exist {
		return model.User{}, errors.New("user not found")
	}
	return user, nil
}

func (r *memRepo) InsertUser(ctx context.Context, user model.User) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exist := r.data[user.Email]
	if exist {
		return model.User{}, errors.New("looks like user exists")
	}

	user.ID = strconv.Itoa(r.nextID)
	r.nextID++
	r.data[user.Email] = user

	return user, nil
}

func (r *memRepo) UpdateUser(ctx context.Context, email string, input model.User) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	curUser, _ := r.data[email]

	curUser.Bio = input.Bio
	curUser.Email = input.Email

	r.data[curUser.Email] = curUser

	if curUser.Email != email {
		delete(r.data, email)
	}

	return curUser, nil
}
