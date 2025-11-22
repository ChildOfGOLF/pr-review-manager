package repository

import "database/sql"

type PRRepository struct {
	db *sql.DB
}

func NewPRRepository(db *sql.DB) *PRRepository {
	return &PRRepository{db: db}
}
