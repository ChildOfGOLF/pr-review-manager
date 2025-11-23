package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"pr-review-manager/internal/domain"
	"pr-review-manager/internal/handler"
	"pr-review-manager/internal/repository"
	"pr-review-manager/internal/router"
	"pr-review-manager/internal/service"
	"pr-review-manager/pkg/database"
)

func setup() (http.Handler, func()) {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "pruser")
	os.Setenv("DB_PASSWORD", "prpass")
	os.Setenv("DB_NAME", "pr_review_db")

	cfg := database.LoadConfigFromEnv()
	db, err := database.Connect(cfg)
	if err != nil {
		panic("failed to connect to db: " + err.Error())
	}

	err = database.RunMigrations(db, "../../migrations")
	if err != nil {
		panic("failed to run migrations: " + err.Error())
	}

	_, _ = db.Exec("TRUNCATE TABLE pull_requests, users, teams CASCADE")

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

	return r, func() {
		db.Close()
	}
}

func TestTeamAndPRFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	r, teardown := setup()
	defer teardown()

	teamPayload := domain.Team{
		TeamName: "Backend",
		Members: []domain.TeamMember{
			{UserID: "u1", Username: "Golf", IsActive: true},
			{UserID: "u2", Username: "Lebron", IsActive: true},
			{UserID: "u3", Username: "Cat", IsActive: true},
		},
	}
	
	body, _ := json.Marshal(teamPayload)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	prPayload := map[string]string{
		"pull_request_id":   "pr-101",
		"pull_request_name": "Fix login bug",
		"author_id":         "u1",
	}

	body, _ = json.Marshal(prPayload)
	req = httptest.NewRequest("POST", "/pullRequest/create", bytes.NewBuffer(body))
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	prData, ok := resp["pr"].(map[string]interface{})
	if !ok {
		t.Fatal("Response does not contain pr object")
	}

	reviewers := prData["assigned_reviewers"].([]interface{})
	if len(reviewers) == 0 {
		t.Error("Expected reviewers to be assigned got 0")
	}

	for _, rev := range reviewers {
		if rev == "u1" {
			t.Error("Author should not be a reviewer")
		}
	}
}

func TestDeactivateTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	r, teardown := setup()
	defer teardown()

	teamPayload := domain.Team{
		TeamName: "Frontend",
		Members: []domain.TeamMember{
			{UserID: "f1", Username: "Frank", IsActive: true},
			{UserID: "f2", Username: "Gofer", IsActive: true},
		},
	}
	body, _ := json.Marshal(teamPayload)
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/team/add", bytes.NewBuffer(body)))

	deactivatePayload := map[string]string{"team_name": "Frontend"}
	body, _ = json.Marshal(deactivatePayload)
	
	req := httptest.NewRequest("POST", "/team/deactivate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]int
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["deactivated_users_count"] != 2 {
		t.Errorf("Expected 2 deactivated users, got %d", resp["deactivated_users_count"])
	}
}
