package main

import (
	"net/http"
	"rwa/internal/handler"
	articleHandler "rwa/internal/handler/article"
	userHandler "rwa/internal/handler/user"
	auth "rwa/internal/middleware"
	articleRepository "rwa/internal/repository/article"
	sessionRepository "rwa/internal/repository/session"
	userRepository "rwa/internal/repository/user"
	articleService "rwa/internal/service/article"
	sessionService "rwa/internal/service/session"
	userService "rwa/internal/service/user"

	"github.com/gorilla/mux"
)

func GetApp() http.Handler {

	// Sessions
	sessionMemRepo := sessionRepository.NewMemRepo()
	sessionService := sessionService.NewService(sessionMemRepo)
	// sessionHandler := sessionHandler.NewHandler(sessionService)

	// Repository - слой db на слайсах
	userMemRepo := userRepository.NewMemRepo()
	articleMemRepo := articleRepository.NewMemRepo()

	// Service
	userService := userService.NewService(userMemRepo, sessionService)
	articleService := articleService.NewService(articleMemRepo)

	// Delivery
	userHandler := userHandler.NewHandler(userService)
	articleHandler := articleHandler.NewHandler(articleService)

	// Router
	mux := mux.NewRouter()

	api := mux.PathPrefix("/api").Subrouter()

	handler.RegisterRoutes(api, userHandler, articleHandler)

	api.Use(auth.AuthMiddleware(sessionService))

	return mux
}
