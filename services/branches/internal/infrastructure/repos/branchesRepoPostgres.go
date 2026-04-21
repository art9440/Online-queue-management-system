package repos

import "database/sql"

type RegistrationRepoPostgres struct {
	db *sql.DB
}

func NewRegistrationRepoPostgres(dsn string) (*RegistrationRepoPostgres, error) {

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &RegistrationRepoPostgres{db: db}, nil
}
