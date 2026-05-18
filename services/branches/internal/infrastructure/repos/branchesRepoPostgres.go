package repos

import (
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
	"database/sql"
	"errors"
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

func (r *BranchesRepoPostgres) GetBookingsByBranchIDAndDate(
	ctx context.Context,
	branchID int64,
	date time.Time,
) ([]domain.Booking, error) {
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
		_ = rows.Close()
	}()

	var bookings []domain.Booking
	for rows.Next() {
		var booking domain.Booking
		var email sql.NullString
		var phone sql.NullString
		var tgUsername sql.NullString
		var comment sql.NullString

		err := rows.Scan(
			&booking.ID,
			&booking.BranchID,
			&booking.Client.ID,
			&email,
			&phone,
			&booking.Client.Name,
			&booking.Client.Surname,
			&tgUsername,
			&booking.Client.CreatedAt,
			&booking.EmployeeID,
			&booking.EmployeeName,
			&booking.EmployeeSurname,
			&booking.ServiceID,
			&booking.ServiceName,
			&booking.StartTime,
			&booking.EndTime,
			&booking.Status,
			&comment,
			&booking.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		booking.Client.Email = nullableString(email)
		booking.Client.Phone = nullableString(phone)
		booking.Client.TgUsername = nullableString(tgUsername)
		booking.Comment = nullableString(comment)
		bookings = append(bookings, booking)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
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
