package repos

import (
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

type BranchesRepoPostgres struct {
	db *sql.DB
}

func NewBranchesRepoPostgres(dsn string) (*BranchesRepoPostgres, error) {

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &BranchesRepoPostgres{db: db}, nil
}

func (r *BranchesRepoPostgres) GetByBusinessID(ctx context.Context, businessID int64) ([]domain.Branch, error) {

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, business_id, name, address FROM branches WHERE business_id = $1`,
		businessID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var res []domain.Branch

	for rows.Next() {
		var b domain.Branch
		if err := rows.Scan(&b.ID, &b.BusinessID, &b.Name, &b.Address); err != nil {
			return nil, err
		}
		res = append(res, b)
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (r *BranchesRepoPostgres) GetByID(ctx context.Context, branchID int64) ([]domain.Branch, error) {

	row := r.db.QueryRowContext(ctx,
		`SELECT id, business_id, name, address 
		 FROM branches 
		 WHERE id = $1`,
		branchID,
	)

	var b domain.Branch

	err := row.Scan(
		&b.ID,
		&b.BusinessID,
		&b.Name,
		&b.Address,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrBranchNotFound
		}
		return nil, err
	}

	return []domain.Branch{b}, nil
}
