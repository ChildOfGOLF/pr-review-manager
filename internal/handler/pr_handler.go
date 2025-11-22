package handler

import (
	"net/http"

	"pr-review-manager/internal/service"
)

type PRHandler struct {
	prService *service.PRService
}

func NewPRHandler(prService *service.PRService) *PRHandler {
	return &PRHandler{prService: prService}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
