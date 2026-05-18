package service

import (
	"Online-queue-management-system/libs/auth"
	liberrors "Online-queue-management-system/libs/errors"
	"Online-queue-management-system/services/booking/internal/domain"
	"context"
)

type BookingService struct {
	repoPostgres BookingRepository
}

func New(repoPostgres BookingRepository) *BookingService {
	return &BookingService{repoPostgres: repoPostgres}
}

func (s *BookingService) GetAppointmentsByEmployeeID(
	ctx context.Context,
	user *auth.AccessClaims,
	employeeID int64,
) ([]domain.Appointment, error) {
	if user == nil {
		return nil, liberrors.ErrUnauthorized
	}

	if employeeID <= 0 {
		return nil, liberrors.ErrInvalidEmployeeID
	}

	switch auth.Role(user.RoleName) {
	case auth.RoleBusinessAdmin, auth.RoleManager:
		return s.repoPostgres.GetAppointmentsByEmployeeID(ctx, employeeID)

	default:
		return nil, liberrors.ErrForbidden
	}
}

func (s *BookingService) GetAppointmentByID(
	ctx context.Context,
	user *auth.AccessClaims,
	appointmentID int64,
) (domain.Appointment, error) {
	if user == nil {
		return domain.Appointment{}, liberrors.ErrUnauthorized
	}

	if appointmentID <= 0 {
		return domain.Appointment{}, domain.ErrInvalidAppointmentID
	}

	switch auth.Role(user.RoleName) {
	case auth.RoleBusinessAdmin, auth.RoleManager:
		appointment, err := s.repoPostgres.GetAppointmentByID(ctx, appointmentID)
		if err != nil {
			return domain.Appointment{}, err
		}

		switch auth.Role(user.RoleName) {
		case auth.RoleBusinessAdmin:
			if appointment.BusinessID != user.BusinessID {
				return domain.Appointment{}, liberrors.ErrForbidden
			}

		case auth.RoleManager:
			if user.BranchID == nil || appointment.BranchID != *user.BranchID {
				return domain.Appointment{}, liberrors.ErrForbidden
			}
		}

		return appointment, nil

	default:
		return domain.Appointment{}, liberrors.ErrForbidden
	}
}

func (s *BookingService) CancelAppointment(
	ctx context.Context,
	user *auth.AccessClaims,
	appointmentID int64,
) error {
	if user == nil {
		return liberrors.ErrUnauthorized
	}

	if appointmentID <= 0 {
		return domain.ErrInvalidAppointmentID
	}

	switch auth.Role(user.RoleName) {
	case auth.RoleBusinessAdmin, auth.RoleManager:
		appointment, err := s.repoPostgres.GetAppointmentByID(ctx, appointmentID)
		if err != nil {
			return err
		}

		switch auth.Role(user.RoleName) {
		case auth.RoleBusinessAdmin:
			if appointment.BusinessID != user.BusinessID {
				return liberrors.ErrForbidden
			}

		case auth.RoleManager:
			if user.BranchID == nil || appointment.BranchID != *user.BranchID {
				return liberrors.ErrForbidden
			}
		}

		return s.repoPostgres.CancelAppointment(ctx, appointmentID)

	default:
		return liberrors.ErrForbidden
	}
}

func (s *BookingService) CreateAppointment(
	ctx context.Context,
	input *domain.CreateAppointmentInput,
) (domain.CreateAppointmentOutput, error) {
	if input == nil {
		return domain.CreateAppointmentOutput{}, domain.ErrInvalidClient
	}

	if input.BranchID <= 0 {
		return domain.CreateAppointmentOutput{}, liberrors.ErrInvalidBranchID
	}

	if input.EmployeeID <= 0 {
		return domain.CreateAppointmentOutput{}, liberrors.ErrInvalidEmployeeID
	}

	if input.ServiceID <= 0 {
		return domain.CreateAppointmentOutput{}, liberrors.ErrInvalidServiceID
	}

	if input.Client.Name == "" || input.Client.Surname == "" {
		return domain.CreateAppointmentOutput{}, domain.ErrInvalidClient
	}

	if input.Client.Phone == "" {
		return domain.CreateAppointmentOutput{}, domain.ErrInvalidClientContact
	}

	if input.StartTime.IsZero() {
		return domain.CreateAppointmentOutput{}, domain.ErrInvalidStartTime
	}

	clientID, exists, err := s.repoPostgres.CheckClientExists(ctx, input.Client)
	if err != nil {
		return domain.CreateAppointmentOutput{}, err
	}

	if !exists {
		clientID, err = s.repoPostgres.CreateClient(ctx, input.Client)
		if err != nil {
			return domain.CreateAppointmentOutput{}, err
		}
	}

	appointment, err := s.repoPostgres.CreateAppointment(ctx, clientID, input)
	if err != nil {
		return domain.CreateAppointmentOutput{}, err
	}

	return appointment, nil
}

func (s *BookingService) GetAvailableSlots(
	ctx context.Context,
	input domain.AvailableSlotsInput,
) ([]domain.AvailableSlot, error) {
	if input.RegistrationSlug == "" {
		return nil, domain.ErrInvalidRegistrationSlug
	}

	if input.BranchID <= 0 {
		return nil, liberrors.ErrInvalidBranchID
	}

	if input.EmployeeID <= 0 {
		return nil, liberrors.ErrInvalidEmployeeID
	}

	if input.ServiceID <= 0 {
		return nil, liberrors.ErrInvalidServiceID
	}

	if input.Date.IsZero() {
		return nil, domain.ErrInvalidStartTime
	}

	return s.repoPostgres.GetAvailableSlots(ctx, input)
}
