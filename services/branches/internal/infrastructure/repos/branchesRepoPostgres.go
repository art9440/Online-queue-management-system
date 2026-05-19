package repos

import (
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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

func (r *BranchesRepoPostgres) BranchBelongsToBusiness(ctx context.Context, branchID, businessID int64) (bool, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM branches
			WHERE id = $1 AND business_id = $2
		)`,
		branchID,
		businessID,
	)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *BranchesRepoPostgres) GetClientsByBranchID(ctx context.Context, branchID int64) ([]domain.Client, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT c.id, c.email, c.phone, c.name, c.surname, c.tg_username, c.created_at
		 FROM clients c
		 JOIN appointments a ON a.client_id = c.id
		 WHERE a.branch_id = $1
		 ORDER BY c.surname, c.name, c.id`,
		branchID,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	var clients []domain.Client
	for rows.Next() {
		client, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clients, nil
}

func (r *BranchesRepoPostgres) GetAppointmentsByBranchIDAndDate(
	ctx context.Context,
	branchID int64,
	date time.Time,
) ([]domain.Appointment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
			a.id,
			a.branch_id,
			c.id,
			c.email,
			c.phone,
			c.name,
			c.surname,
			c.tg_username,
			c.created_at,
			e.id,
			e.name,
			e.surname,
			s.id,
			s.name,
			a.start_time,
			a.end_time,
			a.status,
			a.comment,
			a.created_at
		 FROM appointments a
		 JOIN clients c ON c.id = a.client_id
		 JOIN employees e ON e.id = a.employee_id
		 JOIN services s ON s.id = a.service_id
		 WHERE a.branch_id = $1
		   AND a.start_time >= $2::date
		   AND a.start_time < $2::date + INTERVAL '1 day'
		 ORDER BY a.start_time, a.id`,
		branchID,
		date.Format(time.DateOnly),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			_ = err
		}
	}()

	appointments := make([]domain.Appointment, 0)

	for rows.Next() {
		var (
			appointment domain.Appointment
			client      domain.Client
			employee    domain.Employee
			service     domain.Service
		)

		if err := rows.Scan(
			&appointment.ID,
			&appointment.BranchID,
			&client.ID,
			&client.Email,
			&client.Phone,
			&client.Name,
			&client.Surname,
			&client.TgUsername,
			&client.CreatedAt,
			&employee.ID,
			&employee.Name,
			&employee.Surname,
			&service.ID,
			&service.Name,
			&appointment.StartTime,
			&appointment.EndTime,
			&appointment.Status,
			&appointment.Comment,
			&appointment.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan appointment: %w", err)
		}

		appointment.Client = client

		appointments = append(appointments, appointment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate appointments: %w", err)
	}

	return appointments, nil
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
		_ = rows.Close()
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
		return nil, fmt.Errorf("iterate employees: %w", err)
	}

	return employees, nil
}

type clientScanner interface {
	Scan(dest ...any) error
}

func scanClient(scanner clientScanner) (domain.Client, error) {
	var client domain.Client
	var email sql.NullString
	var phone sql.NullString
	var tgUsername sql.NullString

	err := scanner.Scan(
		&client.ID,
		&email,
		&phone,
		&client.Name,
		&client.Surname,
		&tgUsername,
		&client.CreatedAt,
	)
	if err != nil {
		return domain.Client{}, err
	}

	client.Email = nullableString(email)
	client.Phone = nullableString(phone)
	client.TgUsername = nullableString(tgUsername)

	return client, nil
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
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
	defer func() {
		_ = rows.Close()
	}()

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
	defer func() {
		_ = rows.Close()
	}()

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
	defer func() {
		_ = rows.Close()
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
