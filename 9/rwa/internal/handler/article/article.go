package article

import (
	"net/http"
)

type Service interface {
	RegisterUser()
}

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
}

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {

}
