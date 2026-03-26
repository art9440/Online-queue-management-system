package postgres

import (
	"context"

	"Online-queue-management-system/services/auth/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	const query = `
		SELECT id, login, password_hash, role_id, business_id, branch_id
		FROM users
		WHERE login = $1
		LIMIT 1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.RoleID,
		&user.BusinessID,
		&user.BranchID,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	const query = `
		SELECT id, login, password_hash, role_id, business_id, branch_id
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.RoleID,
		&user.BusinessID,
		&user.BranchID,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}