package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
)

type Response struct {
	Error string `json:"error"`
	Body  any    `json:"response,omitempty"`
}

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {

	case "/user/profile":
		switch r.Method {

		case "GET":
			h.handlerProfile(w, r)
		case "POST":
			h.handlerProfile(w, r)
		default:
			w.WriteHeader(http.StatusNotAcceptable)
			resp := Response{
				Error: "bad method",
			}
			json.NewEncoder(w).Encode(resp)
			return

		}

	case "/user/create":
		switch r.Method {

		case "POST":
			h.handlerCreate(w, r)
		default:
			w.WriteHeader(http.StatusNotAcceptable)
			resp := Response{
				Error: "bad method",
			}
			json.NewEncoder(w).Encode(resp)
			return

		}

	default:
		w.WriteHeader(http.StatusNotFound)
		resp := Response{
			Error: "unknown method",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
}

func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {

	case "/user/create":
		switch r.Method {

		case "POST":
			h.handlerCreate(w, r)
		default:
			w.WriteHeader(http.StatusNotAcceptable)
			resp := Response{
				Error: "bad method",
			}
			json.NewEncoder(w).Encode(resp)
			return

		}

	default:
		w.WriteHeader(http.StatusNotFound)
		resp := Response{
			Error: "unknown method",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
}

func (h *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var params ProfileParams
	// валидирование параметров и заполнение структуры params

	login := r.FormValue("login")

	if login == "" {
		w.WriteHeader(http.StatusBadRequest)
		resp := Response{
			Error: "login must me not empty",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	params.Login = login

	//Do SomeJob
	res, err := h.Profile(r.Context(), params)
	if err != nil {
		var ae ApiError
		if errors.As(err, &ae) {
			w.WriteHeader(ae.HTTPStatus)
			resp := Response{
				Error: err.Error(),
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		resp := Response{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// прочие обработки
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	answer := Response{Error: "", Body: res}
	if err := json.NewEncoder(w).Encode(answer); err != nil {
		var ae *ApiError
		if errors.As(err, ae) {
			w.WriteHeader(ae.HTTPStatus)
			resp := Response{
				Error: err.Error(),
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("X-Auth") != "100500" {
		w.Header().Set("WWW-Authenticate", "Basic realm='api'")
		w.WriteHeader(http.StatusForbidden)
		resp := Response{
			Error: "unauthorized",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var params CreateParams
	// валидирование параметров и заполнение структуры params

	login := r.FormValue("login")

	if login == "" {
		w.WriteHeader(http.StatusBadRequest)
		resp := Response{
			Error: "login must me not empty",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if len(login) < 10 {
		w.WriteHeader(http.StatusBadRequest)
		resp := Response{
			Error: "login len must be >= 10",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	params.Login = login

	name := r.FormValue("full_name")

	params.Name = name

	status := r.FormValue("status")

	if status == "" {
		status = "user"
	}

	if !slices.Contains([]string{"user", "moderator", "admin"}, status) {
		w.WriteHeader(http.StatusBadRequest)
		resp := Response{
			Error: "status must be one of [user, moderator, admin]",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	params.Status = status

	age, err := strconv.Atoi(r.FormValue("age"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := Response{
			Error: "age must be int",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if age > 128 {
		w.WriteHeader(http.StatusBadRequest)
		resp := Response{
			Error: "age must be <= 128",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if age < 0 {
		w.WriteHeader(http.StatusBadRequest)
		resp := Response{
			Error: "age must be >= 0",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	params.Age = age

	//Do SomeJob
	res, err := h.Create(r.Context(), params)
	if err != nil {
		var ae ApiError
		if errors.As(err, &ae) {
			w.WriteHeader(ae.HTTPStatus)
			resp := Response{
				Error: err.Error(),
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		resp := Response{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// прочие обработки
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	answer := Response{Error: "", Body: res}
	if err := json.NewEncoder(w).Encode(answer); err != nil {
		var ae *ApiError
		if errors.As(err, ae) {
			w.WriteHeader(ae.HTTPStatus)
			resp := Response{
				Error: err.Error(),
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("X-Auth") != "100500" {
		w.Header().Set("WWW-Authenticate", "Basic realm='api'")
		w.WriteHeader(http.StatusForbidden)
		resp := Response{
			Error: "unauthorized",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var params OtherCreateParams
	// валидирование параметров и заполнение структуры params

	username := r.FormValue("username")

	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		resp := Response{
			Error: "username must me not empty",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if len(username) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		resp := Response{
			Error: "username len must be >= 3",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	params.Username = username

	name := r.FormValue("account_name")

	params.Name = name

	class := r.FormValue("class")

	if class == "" {
		class = "warrior"
	}

	if !slices.Contains([]string{"warrior", "sorcerer", "rouge"}, class) {
		w.WriteHeader(http.StatusBadRequest)
		resp := Response{
			Error: "class must be one of [warrior, sorcerer, rouge]",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	params.Class = class

	level, err := strconv.Atoi(r.FormValue("level"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := Response{
			Error: "level must be int",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if level > 50 {
		w.WriteHeader(http.StatusBadRequest)
		resp := Response{
			Error: "level must be <= 50",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if level < 1 {
		w.WriteHeader(http.StatusBadRequest)
		resp := Response{
			Error: "level must be >= 1",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	params.Level = level

	//Do SomeJob
	res, err := h.Create(r.Context(), params)
	if err != nil {
		var ae ApiError
		if errors.As(err, &ae) {
			w.WriteHeader(ae.HTTPStatus)
			resp := Response{
				Error: err.Error(),
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		resp := Response{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// прочие обработки
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	answer := Response{Error: "", Body: res}
	if err := json.NewEncoder(w).Encode(answer); err != nil {
		var ae *ApiError
		if errors.As(err, ae) {
			w.WriteHeader(ae.HTTPStatus)
			resp := Response{
				Error: err.Error(),
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
