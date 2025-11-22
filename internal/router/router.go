package router

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"pr-review-manager/internal/handler"
)

func NewRouter(teamHandler *handler.TeamHandler, userHandler *handler.UserHandler, prHandler *handler.PRHandler, statsHandler *handler.StatsHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Get("/stats", statsHandler.GetStats)

	r.Route("/team", func(r chi.Router) {
		r.Post("/add", teamHandler.AddTeam)
		r.Get("/get", teamHandler.GetTeam)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", userHandler.SetIsActive)
		r.Get("/getReview", userHandler.GetReview)
	})

	r.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", prHandler.CreatePR)
		r.Post("/merge", prHandler.MergePR)
		r.Post("/reassign", prHandler.ReassignReviewer)
	})

	return r
}
