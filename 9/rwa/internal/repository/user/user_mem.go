package user

import (
	"context"
	"errors"
	"rwa/internal/model"
	"strconv"
	"sync"
	"time"
)

// TODO:

type memRepo struct {
	mu      sync.RWMutex
	db      map[string]model.User
	byEmail map[string]string
	nextID  int
}

func NewMemRepo() *memRepo {
	return &memRepo{
		db:      make(map[string]model.User),
		byEmail: make(map[string]string),
		nextID:  1,
	}
}

func (r *memRepo) GetByEmail(ctx context.Context, email string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id := r.byEmail[email]

	user, exist := r.db[id]
	if !exist {
		return model.User{}, errors.New("user not found")
	}

	user.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	return user, nil
}

func (r *memRepo) InsertUser(ctx context.Context, user model.User) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exist := r.byEmail[user.Email]
	if exist {
		return model.User{}, errors.New("looks like user exists")
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)

	user.CreatedAt = now
	user.UpdatedAt = now
	user.ID = strconv.Itoa(r.nextID)

	r.nextID++

	r.db[user.ID] = user
	r.byEmail[user.Email] = user.ID

	return user, nil
}

func (r *memRepo) UpdateUser(ctx context.Context, id string, input model.User) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	curUser, ok := r.db[id]
	if !ok {
		return model.User{}, errors.New("not found")
	}

	if input.Bio != "" {
		curUser.Bio = input.Bio
	}

	if input.Email != "" {
		oldEmail := curUser.Email
		curUser.Email = input.Email
		delete(r.byEmail, oldEmail)
		r.byEmail[curUser.Email] = id
	}

	curUser.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	r.db[id] = curUser

	return curUser, nil
}
