package main

import (
	"log"
	"net/http"

	"pr-review-manager/internal/handler"
	"pr-review-manager/internal/repository"
	"pr-review-manager/internal/router"
	"pr-review-manager/internal/service"
	"pr-review-manager/pkg/database"
)

func main() {
	cfg := database.LoadConfigFromEnv()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db, "./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	teamRepo := repository.NewTeamRepository(db)
	userRepo := repository.NewUserRepository(db)
	prRepo := repository.NewPRRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	teamService := service.NewTeamService(teamRepo, userRepo, prRepo)
	userService := service.NewUserService(userRepo, prRepo)
	prService := service.NewPRService(prRepo, userRepo)
	statsService := service.NewStatsService(statsRepo)

	teamHandler := handler.NewTeamHandler(teamService)
	userHandler := handler.NewUserHandler(userService)
	prHandler := handler.NewPRHandler(prService)
	statsHandler := handler.NewStatsHandler(statsService)

	r := router.NewRouter(teamHandler, userHandler, prHandler, statsHandler)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
