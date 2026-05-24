package service

import (
	"Online-queue-management-system/services/booking/internal/domain"
	"context"
)

type BookingRepository interface {
	GetAppointmentsByEmployeeID(ctx context.Context, employeeID int64) ([]domain.Appointment, error)
	GetAppointmentByID(ctx context.Context, appointmentID int64) (domain.Appointment, error)
	CancelAppointment(ctx context.Context, appointmentID int64) error
	GetAvailableSlots(ctx context.Context, input domain.AvailableSlotsInput) ([]domain.AvailableSlot, error)
	CheckClientExists(ctx context.Context, client domain.ClientInput) (int64, bool, error)
	CreateClient(ctx context.Context, client domain.ClientInput) (int64, error)
	CreateAppointment(ctx context.Context, clientID int64, input *domain.CreateAppointmentInput) (domain.CreateAppointmentOutput, error)
}

type CalendarTokenRepository interface {
	SaveOAuthState(ctx context.Context, state string, userID int64) error
	GetOAuthState(ctx context.Context, state string) (int64, error)
	DeleteOAuthState(ctx context.Context, state string) error
	SaveToken(ctx context.Context, userID int64, token domain.GoogleCalendarToken) error
	GetToken(ctx context.Context, userID int64) (domain.GoogleCalendarToken, error)
	SavePublicExportToken(ctx context.Context, token string, appointmentID int64) error
	GetPublicExportToken(ctx context.Context, token string) (int64, error)
	SavePublicOAuthState(ctx context.Context, state string, appointmentID int64) error
	GetPublicOAuthState(ctx context.Context, state string) (int64, error)
	DeletePublicOAuthState(ctx context.Context, state string) error
}

type CalendarExporter interface {
	AuthCodeURL(state string) (string, error)
	Exchange(ctx context.Context, code string) (domain.GoogleCalendarToken, error)
	ExportAppointment(ctx context.Context, token domain.GoogleCalendarToken, appointment *domain.Appointment) (domain.GoogleCalendarEvent, domain.GoogleCalendarToken, error)
}

type AppointmentNotificationRepository interface {
	SaveAppointmentNotifications(
		ctx context.Context,
		appointment *domain.CreateAppointmentOutput,
		client *domain.ClientInput,
	) error
}
