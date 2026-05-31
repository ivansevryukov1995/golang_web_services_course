package user

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	sessionHandler "rwa/internal/handler/session"
	"rwa/internal/handler/user/dto"
	"rwa/internal/model"
	userService "rwa/internal/service/user"
)

type ServiceUser interface {
	RegisterUser(ctx context.Context, in userService.UserInput) (model.User, error)
	LoginUser(ctx context.Context, in userService.UserInput) (model.User, error)
	CurrentUser(ctx context.Context, email string) (model.User, error)
	UpdateUser(ctx context.Context, email string, in userService.UserInput) (model.User, error)
}

type ServiceSession interface {
	Check(ctx context.Context, token string) (model.Session, error)
	Create(ctx context.Context, userID string, email string) (string, error)
}

type handler struct {
	user    ServiceUser
	session ServiceSession
}

func NewHandler(service ServiceUser, session ServiceSession) *handler {
	return &handler{
		user:    service,
		session: session,
	}
}

func (h *handler) Reg(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	req := dto.UserRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.user.RegisterUser(r.Context(), dto.ToUserInput(req))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, _ := h.session.Create(r.Context(), user.ID, user.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(dto.ToUserRegResponse(user, token))
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	req := dto.UserRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.user.LoginUser(r.Context(), dto.ToUserInput(req))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, _ := h.session.Create(r.Context(), user.ID, user.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(dto.ToUserLoginResponse(user, token))
}

func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {

	sess, err := sessionHandler.SessionFromContext(r.Context())

	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("sess %s", sess)

	user, err := h.user.CurrentUser(r.Context(), sess.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("user %s", user)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(dto.ToUserCurrentResponse(user, sess.Token))
}

func (h *handler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	sess, err := sessionHandler.SessionFromContext(r.Context())

	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()

	req := dto.UserRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.user.UpdateUser(r.Context(), sess.Email, dto.ToUserInput(req))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// log.Printf("bio %s", user.Bio)

	token, _ := h.session.Create(r.Context(), user.ID, user.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(dto.ToUserUpdateResponse(user, token))
}
