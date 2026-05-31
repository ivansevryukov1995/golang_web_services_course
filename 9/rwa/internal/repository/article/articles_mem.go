package article

import (
	"errors"
	"rwa/internal/model"
	"sync"
)

// TODO:

type memRepo struct {
	mu   sync.RWMutex
	data map[string]model.Article
}

func NewMemRepo() *memRepo {
	return &memRepo{
		data: make(map[string]model.Article),
	}
}

func (r *memRepo) GetUser(email string) (model.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exist := r.data[email]
	if !exist {
		return model.Article{}, errors.New("user not found")
	}
	return user, nil
}

func (r *memRepo) InsertUser(email string, article model.Article) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exist := r.data[email]
	if exist {
		return 0, errors.New("looks like user exists")
	}

	// user.User.ID = strconv.Itoa(len(r.data) + 1)
	// r.data[email] = user

	return len(r.data), nil
}
