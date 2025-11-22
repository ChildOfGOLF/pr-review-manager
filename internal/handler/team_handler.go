package handler

import (
	"encoding/json"
	"net/http"

	"pr-review-manager/internal/domain"
	"pr-review-manager/internal/errors"
	"pr-review-manager/internal/service"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var team domain.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	result, err := h.teamService.AddTeam(&team)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"team": result,
	})
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "team_name is required")
		return
	}

	team, err := h.teamService.GetTeam(teamName)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, team)
}

func (h *TeamHandler) DeactivateTeam(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TeamName string `json:"team_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.TeamName == "" {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "team_name is required")
		return
	}

	deactivatedCount, affectedPRs, err := h.teamService.DeactivateTeam(req.TeamName)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"deactivated_users_count": deactivatedCount,
		"affected_prs_count":      affectedPRs,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func handleServiceError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		respondError(w, appErr.HTTPStatus, appErr.Code, appErr.Message)
	} else {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
