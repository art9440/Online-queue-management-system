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
	CreateAppointment(ctx context.Context, clientID int64, input domain.CreateAppointmentInput) (domain.CreateAppointmentOutput, error)
}
