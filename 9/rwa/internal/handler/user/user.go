package user

import (
	"context"
	"encoding/json"
	"net/http"
	"rwa/internal/handler/user/dto"
	auth "rwa/internal/middleware"
	"rwa/internal/model"
	userService "rwa/internal/service/user"
)

type ServiceUser interface {
	RegisterUser(ctx context.Context, in userService.UserInput) (model.User, string, error)
	LoginUser(ctx context.Context, in userService.UserInput) (model.User, string, error)
	CurrentUser(ctx context.Context, email string) (model.User, error)
	UpdateUser(ctx context.Context, id string, in userService.UserInput) (model.User, string, error)
}

type handler struct {
	user ServiceUser
}

func NewHandler(service ServiceUser) *handler {
	return &handler{
		user: service,
	}
}

func (h *handler) Reg(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	req := dto.UserRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, token, err := h.user.RegisterUser(r.Context(), dto.ToUserInput(req))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	user, token, err := h.user.LoginUser(r.Context(), dto.ToUserInput(req))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(dto.ToUserLoginResponse(user, token))
}

func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {

	sess, err := auth.SessionFromContext(r.Context())

	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.user.CurrentUser(r.Context(), sess.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(dto.ToUserCurrentResponse(user, sess.Token))
}

func (h *handler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	sess, err := auth.SessionFromContext(r.Context())

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

	user, token, err := h.user.UpdateUser(r.Context(), sess.UserID, dto.ToUserInput(req))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(dto.ToUserUpdateResponse(user, token))
}
