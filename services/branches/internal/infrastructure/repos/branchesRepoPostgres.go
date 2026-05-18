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
	defer rows.Close()

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

func (r *BranchesRepoPostgres) GetServicesByBusinessID(
	ctx context.Context,
	businessID int64,
) ([]domain.Service, error) {
	const query = `
		SELECT DISTINCT
			s.id,
			s.branch_id,
			s.name,
			s.duration_minutes,
			s.price
		FROM services s
		JOIN branches b ON s.branch_id = b.id
		WHERE b.business_id = $1
		ORDER BY s.name
	`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, fmt.Errorf("query services by business id: %w", err)
	}
	defer rows.Close()

	services := make([]domain.Service, 0)

	for rows.Next() {
		var service domain.Service

		if err := rows.Scan(
			&service.ID,
			&service.BranchID,
			&service.Name,
			&service.DurationMinutes,
			&service.Price,
		); err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}

		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate services rows: %w", err)
	}

	return services, nil
}

func (r *BranchesRepoPostgres) GetBranchesWithService(
	ctx context.Context,
	businessID int64,
	serviceID int64,
) ([]domain.Branch, error) {
	const query = `
		SELECT DISTINCT
			b.id,
			b.business_id,
			b.name,
			b.address
		FROM branches b
		JOIN services s ON b.id = s.branch_id
		WHERE b.business_id = $1 AND s.id = $2
		ORDER BY b.name
	`

	rows, err := r.db.QueryContext(ctx, query, businessID, serviceID)
	if err != nil {
		return nil, fmt.Errorf("query branches with service: %w", err)
	}
	defer rows.Close()

	branches := make([]domain.Branch, 0)

	for rows.Next() {
		var branch domain.Branch

		if err := rows.Scan(
			&branch.ID,
			&branch.BusinessID,
			&branch.Name,
			&branch.Address,
		); err != nil {
			return nil, fmt.Errorf("scan branch: %w", err)
		}

		branches = append(branches, branch)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate branches rows: %w", err)
	}

	return branches, nil
}

func (r *BranchesRepoPostgres) GetEmployeesByServiceAndBranch(
	ctx context.Context,
	serviceID int64,
	branchID int64,
) ([]domain.Employee, error) {
	const query = `
		SELECT
			e.id,
			e.branch_id,
			e.name,
			e.surname,
			e.position
		FROM employees e
		JOIN employee_services es ON e.id = es.employee_id
		WHERE es.service_id = $1 AND e.branch_id = $2
		ORDER BY e.name
	`

	rows, err := r.db.QueryContext(ctx, query, serviceID, branchID)
	if err != nil {
		return nil, fmt.Errorf("query employees by service and branch: %w", err)
	}
	defer rows.Close()

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

func (r *BranchesRepoPostgres) GetBusinessIDByRegistrationSlug(
	ctx context.Context,
	registrationSlug string,
) (int64, error) {
	const query = `
		SELECT id
		FROM businesses
		WHERE registration_slug = $1
	`

	var businessID int64

	err := r.db.QueryRowContext(ctx, query, registrationSlug).Scan(&businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("business not found with slug: %s", registrationSlug)
		}
		return 0, fmt.Errorf("query business by slug: %w", err)
	}

	return businessID, nil
}
