package handler

import (
	"net/http"

	"pr-review-manager/internal/service"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
