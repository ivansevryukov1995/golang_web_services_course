package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		h.handlerProfile(w, r)
	case "/user/create":
		h.handlerCreate(w, r)
	default:
		http.Error(w, "unknown method", http.StatusNotFound)
	}
}

func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		h.handlerCreate(w, r)
	default:
		http.Error(w, "unknown method", http.StatusNotFound)
	}
}

func (h *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var params ProfileParams
	if err := json.Unmarshal(body, &params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if params.Login == "" {
		http.Error(w, "login must me not empty", http.StatusBadRequest)
		return
	}

	res, err := h.Profile(r.Context(), params)
	if err != nil {
		var ae *ApiError
		if errors.As(err, ae) {
			http.Error(w, err.Error(), ae.HTTPStatus)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Auth") != "100500" {
		w.Header().Set("WWW-Authenticate", "Basic realm='api'")
		http.Error(w, "unauthorized", http.StatusForbidden)
		return
	}

	values := r.URL.Query()

	res, err := h.Create(r.Context(), params)
	if err != nil {
		var ae *ApiError
		if errors.As(err, ae) {
			http.Error(w, err.Error(), ae.HTTPStatus)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Auth") != "100500" {
		w.Header().Set("WWW-Authenticate", "Basic realm='api'")
		http.Error(w, "unauthorized", http.StatusForbidden)
		return
	}

	values := r.URL.Query()

	res, err := h.Create(r.Context(), params)
	if err != nil {
		var ae *ApiError
		if errors.As(err, ae) {
			http.Error(w, err.Error(), ae.HTTPStatus)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
