package repos

import (
	"Online-queue-management-system/services/booking/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

type BookingRepoPostgres struct {
	db *sql.DB
}

func NewBookingRepoPostgres(dsn string) (*BookingRepoPostgres, error) {

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &BookingRepoPostgres{db: db}, nil
}

func (r *BookingRepoPostgres) GetAppointmentsByEmployeeID(
	ctx context.Context,
	employeeID int64,
) ([]domain.Appointment, error) {
	const query = `
		SELECT
			id,
			client_id,
			branch_id,
			employee_id,
			service_id,
			start_time,
			end_time,
			status,
			comment
		FROM appointments
		WHERE employee_id = $1
		ORDER BY start_time
	`

	rows, err := r.db.QueryContext(ctx, query, employeeID)
	if err != nil {
		return nil, fmt.Errorf("query appointments by employee id: %w", err)
	}
	defer rows.Close()

	appointments := make([]domain.Appointment, 0)

	for rows.Next() {
		var appointment domain.Appointment

		if err := rows.Scan(
			&appointment.ID,
			&appointment.ClientID,
			&appointment.BranchID,
			&appointment.EmployeeID,
			&appointment.ServiceID,
			&appointment.StartTime,
			&appointment.EndTime,
			&appointment.Status,
			&appointment.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan appointment: %w", err)
		}

		appointments = append(appointments, appointment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate appointments rows: %w", err)
	}

	return appointments, nil
}

func (r *BookingRepoPostgres) GetAppointmentByID(ctx context.Context, appointmentID int64) (domain.Appointment, error) {
	return domain.Appointment{}, nil
}

func (r *BookingRepoPostgres) CancelAppointment(ctx context.Context, appointmentID int64) error {
	return nil
}

func (r *BookingRepoPostgres) CheckClientExists(
	ctx context.Context,
	client domain.ClientInput,
) (int64, bool, error) {
	const query = `
		SELECT id
		FROM clients
		WHERE phone = $1
		LIMIT 1
	`

	var clientID int64

	err := r.db.QueryRowContext(ctx, query, client.Phone).Scan(&clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}

		return 0, false, fmt.Errorf("check client exists: %w", err)
	}

	return clientID, true, nil
}

func (r *BookingRepoPostgres) CreateClient(
	ctx context.Context,
	client domain.ClientInput,
) (int64, error) {
	const query = `
		INSERT INTO clients (
			email,
			phone,
			name,
			surname,
			tg_username,
			password_hash
		)
		VALUES ($1, $2, $3, $4, $5, NULL)
		RETURNING id
	`

	var clientID int64

	err := r.db.QueryRowContext(
		ctx,
		query,
		nullableString(client.Email),
		client.Phone,
		client.Name,
		client.Surname,
		nullableString(client.TgUsername),
	).Scan(&clientID)

	if err != nil {
		existingID, exists, findErr := r.CheckClientExists(ctx, client)
		if findErr == nil && exists {
			return existingID, nil
		}

		return 0, fmt.Errorf("create client: %w", err)
	}

	return clientID, nil
}

func (r *BookingRepoPostgres) CreateAppointment(
	ctx context.Context,
	clientID int64,
	input domain.CreateAppointmentInput,
) (domain.CreateAppointmentOutput, error) {
	const query = `
		WITH created_appointment AS (
			INSERT INTO appointments (
				client_id,
				branch_id,
				employee_id,
				service_id,
				start_time,
				end_time,
				status,
				comment
			)
			SELECT
				$1,
				$2,
				$3,
				$4,
				$5,
				$5::timestamptz + (s.duration_minutes * interval '1 minute'),
				$6,
				$7
			FROM services s
			JOIN employees e ON e.id = $3
			JOIN employee_services es 
				ON es.employee_id = e.id 
				AND es.service_id = s.id
			WHERE s.id = $4
			  AND s.branch_id = $2
			  AND e.branch_id = $2
			  AND EXISTS (
				  SELECT 1
				  FROM employee_schedules sch
				  WHERE sch.employee_id = e.id
				    AND sch.starts_at <= $5
				    AND sch.ends_at >= ($5::timestamptz + (s.duration_minutes * interval '1 minute'))
			  )
			RETURNING
				id,
				client_id,
				branch_id,
				employee_id,
				service_id,
				start_time,
				end_time,
				status,
				comment
		)
		SELECT
			a.id,
			a.client_id,

			a.branch_id,
			b.name,

			a.employee_id,
			e.name,
			e.surname,

			a.service_id,
			s.name,

			a.start_time,
			a.end_time,
			a.status,
			a.comment
		FROM created_appointment a
		JOIN branches b ON b.id = a.branch_id
		JOIN employees e ON e.id = a.employee_id
		JOIN services s ON s.id = a.service_id
	`

	var output domain.CreateAppointmentOutput

	err := r.db.QueryRowContext(
		ctx,
		query,
		clientID,
		input.BranchID,
		input.EmployeeID,
		input.ServiceID,
		input.StartTime,
		domain.AppointmentStatusPending,
		input.Comment,
	).Scan(
		&output.AppointmentID,
		&output.ClientID,

		&output.BranchID,
		&output.BranchName,

		&output.EmployeeID,
		&output.EmployeeName,
		&output.EmployeeSurname,

		&output.ServiceID,
		&output.ServiceName,

		&output.StartTime,
		&output.EndTime,
		&output.Status,
		&output.Comment,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23P01" {
				return domain.CreateAppointmentOutput{}, domain.ErrTimeSlotBusy
			}
		}

		if errors.Is(err, sql.ErrNoRows) {
			return domain.CreateAppointmentOutput{}, domain.ErrAppointmentNotAvailable
		}

		return domain.CreateAppointmentOutput{}, fmt.Errorf("create appointment: %w", err)
	}

	return output, nil
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}

	return *value
}
