package repos

import (
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"

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

	if err := db.Ping(); err != nil {
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

	defer func() {
		_ = rows.Close()
	}()

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

func (r *BranchesRepoPostgres) GetEmployeesByBranchID(
	ctx context.Context,
	branchID int64,
) ([]domain.Employee, error) {
	const query = `
		SELECT
			id,
			branch_id,
			name,
			surname,
			position
		FROM employees
		WHERE branch_id = $1
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, branchID)
	if err != nil {
		return nil, fmt.Errorf("query employees by branch id: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Silently ignore close errors in defer as there's nothing we can do
			_ = err
		}
	}()

	employees := make([]domain.Employee, 0)

	for rows.Next() {
		var employee domain.Employee

		if err := rows.Scan(
			&employee.ID,
			&employee.BranchID,
			&employee.Name,
			&employee.Surname,
			&employee.Position,
		); err != nil {
			return nil, fmt.Errorf("scan employee: %w", err)
		}

		employees = append(employees, employee)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate employees rows: %w", err)
	}

	return employees, nil
}
