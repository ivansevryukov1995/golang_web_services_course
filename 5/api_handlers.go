package main

import "net/http"

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		h.handlerProfile(w, r)
	case "/user/create":
		h.handlerCreate(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		h.handlerCreate(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func (h *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {
	res, err := h.Profile(ctx, params)
}

func (h *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	res, err := h.Create(ctx, params)
}

func (h *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	res, err := h.Create(ctx, params)
}

