package errors

import "fmt"

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
	}
}

var (
	ErrTeamExists  = NewAppError("TEAM_EXISTS", "team_name already exists", 400)
	ErrPRExists    = NewAppError("PR_EXISTS", "PR id already exists", 409)
	ErrPRMerged    = NewAppError("PR_MERGED", "cannot reassign on merged PR", 409)
	ErrNotAssigned = NewAppError("NOT_ASSIGNED", "reviewer is not assigned to this PR", 409)
	ErrNoCandidate = NewAppError("NO_CANDIDATE", "no active replacement candidate in team", 409)
	ErrNotFound    = NewAppError("NOT_FOUND", "resource not found", 404)
)
