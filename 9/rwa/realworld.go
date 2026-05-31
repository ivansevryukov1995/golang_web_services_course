package main

import (
	"net/http"
	"rwa/internal/handler"
	articleHandler "rwa/internal/handler/article"
	sessionHandler "rwa/internal/handler/session"
	userHandler "rwa/internal/handler/user"
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
	userService := userService.NewService(userMemRepo)
	articleService := articleService.NewService(articleMemRepo)

	// Delivery
	userHandler := userHandler.NewHandler(userService, sessionService)
	articleHandler := articleHandler.NewHandler(articleService)

	// Router
	mux := mux.NewRouter()

	api := mux.PathPrefix("/api").Subrouter()

	handler.RegisterRoutes(api, userHandler, articleHandler)

	api.Use(sessionHandler.AuthMiddleware(sessionService))

	return mux
}
