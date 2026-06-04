package session

import (
	"context"
	"errors"
	"rwa/internal/model"
	"sync"
)

type memRepo struct {
	mu sync.RWMutex
	db map[string]model.Session
}

func NewMemRepo() *memRepo {
	return &memRepo{
		db: make(map[string]model.Session),
	}
}

func (sm *memRepo) Get(ctx context.Context, token string) (model.Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sess, exist := sm.db[token]
	if !exist {
		return model.Session{}, errors.New("no session found")
	}
	return sess, nil
}

func (sm *memRepo) Create(ctx context.Context, userID string, email string) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	token := userID

	sm.db[token] = model.Session{
		UserID: userID,
		Email:  email,
		Token:  token,
	}

	return token, nil
}

func (sm *memRepo) DestroyCurrent(ctx context.Context, token string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.db[token]; ok {
		delete(sm.db, token)
	}
}

// func (sm *memRepo) DestroyAll(w http.ResponseWriter, user *User) error {
// 	result, err := sm.DB.Exec("DELETE FROM sessions WHERE user_id = ?",
// 		user.ID)
// 	if err != nil {
// 		return err
// 	}

// 	affected, _ := result.RowsAffected()
// 	log.Println("destroyed sessions", affected, "for user", user.ID)

// 	return nil
// }
