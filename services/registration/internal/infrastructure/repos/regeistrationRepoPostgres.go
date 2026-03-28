package repos

import (
	"Online-queue-management-system/services/registration/internal/domain/pending"
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

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

func (r *RegistrationRepoPostgres) CreateUserWithBusiness(ctx context.Context, p pending.PendingRegistration) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. создать бизнес
	var businessID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO businesses (name, type)
		VALUES ($1, $2)
		RETURNING id
	`, p.BusinessName, p.BusinessType).Scan(&businessID)
	if err != nil {
		return fmt.Errorf("insert business: %w", err)
	}

	// 2. получить role_id для "business_admin"
	var roleID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM roles WHERE name = $1
	`, "business_admin").Scan(&roleID)
	if err != nil {
		return fmt.Errorf("get role: %w", err)
	}

	// 3. создать пользователя
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (login, password_hash, role_id, business_id)
		VALUES ($1, $2, $3, $4)
	`, p.Email, p.PasswordHash, roleID, businessID)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return tx.Commit()
}
