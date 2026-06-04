package handler

import (
	"context"
	"net/http"
	auth "rwa/internal/middleware"
	"rwa/internal/model"

	"github.com/gorilla/mux"
)

type UserHandler interface {
	Reg(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	GetCurrentUser(w http.ResponseWriter, r *http.Request)
	UpdateCurrentUser(w http.ResponseWriter, r *http.Request)
}

type ArticleHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	GetByAuthor(w http.ResponseWriter, r *http.Request)
	GetByTag(w http.ResponseWriter, r *http.Request)
}

type SessionService interface {
	Check(ctx context.Context, token string) (model.Session, error)
	Create(ctx context.Context, userID string, email string) (string, error)
}

func RegisterRoutes(mux *mux.Router, userHandler UserHandler, articleHandler ArticleHandler, sessionService SessionService) {

	private := mux.PathPrefix("/api").Subrouter()

	private.HandleFunc("/user", userHandler.GetCurrentUser).
		Methods("GET")

	private.HandleFunc("/user", userHandler.UpdateCurrentUser).
		Methods("PUT")

	private.HandleFunc("/user/logout", userHandler.Logout).
		Methods("POST")

	private.HandleFunc("/articles", articleHandler.Create).
		Methods("POST")

	private.HandleFunc("/articles", articleHandler.GetByAuthor).
		Methods("GET").
		Queries("author", "{author}")

	private.HandleFunc("/articles", articleHandler.GetByTag).
		Methods("GET").
		Queries("tag", "{tag}")

	private.Use(auth.AuthMiddleware(sessionService))

	public := mux.PathPrefix("/api").Subrouter()

	public.HandleFunc("/users", userHandler.Reg).
		Methods("POST")

	public.HandleFunc("/users/login", userHandler.Login).
		Methods("POST")

	public.HandleFunc("/articles", articleHandler.GetAll).
		Methods("GET")

}
