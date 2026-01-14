package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
)

type resp struct {
	Error    string      `json:"error"`
	Response interface{} `json:"response,omitempty"`
}

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

	values := r.URL.Query()

	var params ProfileParams
	// валидирование параметров и заполнение структуры params

	login := values.Get("login")

	if login == "" {
		http.Error(w, "login must me not empty", http.StatusBadRequest)
		return
	}

	params.Login = login

	//Do SomeJob
	res, err := h.Profile(r.Context(), params)
	if err != nil {
		var ae *ApiError
		if errors.As(err, ae) {
			http.Error(w, err.Error(), ae.HTTPStatus)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// прочие обработки
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	answer := resp{Error: "", Response: res}
	if err := json.NewEncoder(w).Encode(answer); err != nil {
		var ae *ApiError
		if errors.As(err, ae) {
			http.Error(w, err.Error(), ae.HTTPStatus)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("X-Auth") != "100500" {
		w.Header().Set("WWW-Authenticate", "Basic realm='api'")
		http.Error(w, "unauthorized", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var params CreateParams
	// валидирование параметров и заполнение структуры params

	login := r.FormValue("login")

	if len(login) < 10 {
		http.Error(w, "login len must be >= 10", http.StatusBadRequest)
		return
	}

	if login == "" {
		http.Error(w, "login must me not empty", http.StatusBadRequest)
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
		http.Error(w, "status must be one of user|moderator|admin", http.StatusBadRequest)
		return
	}

	params.Status = status

	age, err := strconv.Atoi(r.FormValue("age"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if age > 128 {
		http.Error(w, "age must be <= 128", http.StatusBadRequest)
		return
	}

	if age < 0 {
		http.Error(w, "age must be >= 0", http.StatusBadRequest)
		return
	}

	params.Age = age

	//Do SomeJob
	res, err := h.Create(r.Context(), params)
	if err != nil {
		var ae *ApiError
		if errors.As(err, ae) {
			http.Error(w, err.Error(), ae.HTTPStatus)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// прочие обработки
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	answer := resp{Error: "", Response: res}
	if err := json.NewEncoder(w).Encode(answer); err != nil {
		var ae *ApiError
		if errors.As(err, ae) {
			http.Error(w, err.Error(), ae.HTTPStatus)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// writeJSONError возвращает клиенту JSON-ошибку вида {"error": "...", "response": null}
func writeJSONError(w http.ResponseWriter, msg string, code int) {
	// w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error": msg, // ← нет поля response
	})
}

func (h *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	// 1. auth
	if r.Header.Get("X-Auth") != "100500" {
		w.Header().Set("WWW-Authenticate", "Basic realm='api'")
		writeJSONError(w, "unauthorized", http.StatusForbidden)
		return
	}

	// 2. парсим тело
	if err := r.ParseForm(); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 3. валидация + заполнение
	var params OtherCreateParams

	username := r.FormValue("username")
	if username == "" || len(username) < 3 {
		writeJSONError(w, "username len must be >= 3", http.StatusBadRequest)
		return
	}
	params.Username = username

	name := r.FormValue("account_name")
	params.Name = name

	class := r.FormValue("class")
	if !slices.Contains([]string{"warrior", "sorcerer", "rouge"}, class) {
		writeJSONError(w, "class must be one of [warrior, sorcerer, rouge]", http.StatusBadRequest)
		return
	}
	params.Class = class

	level, err := strconv.Atoi(r.FormValue("level"))
	if err != nil || level < 1 || level > 50 {
		writeJSONError(w, "level must be 1-50", http.StatusBadRequest)
		return
	}
	params.Level = level

	// 4. бизнес-логика
	res, err := h.Create(r.Context(), params)
	if err != nil {
		var ae *ApiError
		if errors.As(err, &ae) {
			writeJSONError(w, err.Error(), ae.HTTPStatus)
			return
		}
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. успешный JSON-ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":    "",
		"response": res,
	})
}
