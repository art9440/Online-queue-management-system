package repos

import (
	"database/sql"
	"log/slog"
)

type RegistrationRepo struct {
	db  *sql.DB
	log *slog.Logger
}

func NewRegistrationRepo(log *slog.Logger, dsn string) (*RegistrationRepo, error) {

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &RegistrationRepo{db: db, log: log}, nil
}
