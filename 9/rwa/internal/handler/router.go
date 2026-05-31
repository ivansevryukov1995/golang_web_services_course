package handler

import (
	"net/http"

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
	Get(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
}

func RegisterRoutes(mux *mux.Router, userHandler UserHandler, articleHandler ArticleHandler) {
	mux.HandleFunc("/users", userHandler.Reg).
		Methods("POST")

	mux.HandleFunc("/users/login", userHandler.Login).
		Methods("POST")

	mux.HandleFunc("/user", userHandler.GetCurrentUser).
		Methods("GET")
	mux.HandleFunc("/user", userHandler.UpdateCurrentUser).
		Methods("PUT")

	mux.HandleFunc("/user/logout", userHandler.Logout).
		Methods("POST")

	mux.HandleFunc("/articles", articleHandler.Create).
		Methods("POST")
	mux.HandleFunc("/articles", articleHandler.Get).
		Methods("GET")
}
