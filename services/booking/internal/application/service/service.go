package service

import (
	"Online-queue-management-system/libs/auth"
	liberrors "Online-queue-management-system/libs/errors"
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/booking/internal/domain"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

type BookingService struct {
	repoPostgres     BookingRepository
	tokenRepo        CalendarTokenRepository
	exporter         CalendarExporter
	notificationRepo AppointmentNotificationRepository
}

func New(
	repoPostgres BookingRepository,
	tokenRepo CalendarTokenRepository,
	exporter CalendarExporter,
	notificationRepo AppointmentNotificationRepository,
) *BookingService {
	return &BookingService{
		repoPostgres:     repoPostgres,
		tokenRepo:        tokenRepo,
		exporter:         exporter,
		notificationRepo: notificationRepo,
	}
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

	if input.RegistrationSlug == "" {
		return domain.CreateAppointmentOutput{}, domain.ErrInvalidRegistrationSlug
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

	if s.notificationRepo != nil {
		if err := s.notificationRepo.SaveAppointmentNotifications(ctx, &appointment, &input.Client); err != nil {
			logger.From(ctx).Error(
				"failed to create appointment notifications",
				"appointment_id", appointment.AppointmentID,
				"err", err,
			)
		}
	}

	if s.tokenRepo != nil {
		exportToken, err := generateOAuthState()
		if err != nil {
			return domain.CreateAppointmentOutput{}, err
		}
		if err := s.tokenRepo.SavePublicExportToken(ctx, exportToken, appointment.AppointmentID); err != nil {
			return domain.CreateAppointmentOutput{}, err
		}
		appointment.CalendarICSURL = fmt.Sprintf(
			"/public/appointments/%d/calendar.ics?token=%s",
			appointment.AppointmentID,
			exportToken,
		)
		if s.exporter != nil {
			appointment.GoogleCalendarExportURL = fmt.Sprintf(
				"/public/appointments/%d/google-calendar/auth-url?token=%s",
				appointment.AppointmentID,
				exportToken,
			)
		}
	}

	return appointment, nil
}

// GetPublicAppointmentCalendar returns event details for an appointment if exportToken was issued for it.
func (s *BookingService) GetPublicAppointmentCalendar(
	ctx context.Context,
	appointmentID int64,
	exportToken string,
) (domain.AppointmentCalendarDetails, error) {
	if appointmentID <= 0 || exportToken == "" || s.tokenRepo == nil {
		return domain.AppointmentCalendarDetails{}, domain.ErrCalendarExportNotFound
	}

	tokenAppointmentID, err := s.tokenRepo.GetPublicExportToken(ctx, exportToken)
	if errors.Is(err, domain.ErrGoogleCalendarNotLinked) {
		return domain.AppointmentCalendarDetails{}, domain.ErrCalendarExportNotFound
	}
	if err != nil {
		return domain.AppointmentCalendarDetails{}, err
	}
	if tokenAppointmentID != appointmentID {
		return domain.AppointmentCalendarDetails{}, domain.ErrCalendarExportNotFound
	}

	return s.repoPostgres.GetAppointmentCalendarDetails(ctx, appointmentID)
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

func (s *BookingService) GoogleCalendarAuthURL(ctx context.Context, user *auth.AccessClaims) (string, error) {
	if user == nil {
		return "", liberrors.ErrUnauthorized
	}
	if s.tokenRepo == nil || s.exporter == nil {
		return "", domain.ErrGoogleCalendarDisabled
	}

	state, err := generateOAuthState()
	if err != nil {
		return "", err
	}
	if err := s.tokenRepo.SaveOAuthState(ctx, state, user.UserID); err != nil {
		return "", err
	}

	return s.exporter.AuthCodeURL(state)
}

// CompleteGoogleCalendarOAuth handles the Google OAuth callback for both the public and the dashboard flow.
// The result is filled whenever the flow is known, even on error, so the caller can redirect the user back.
func (s *BookingService) CompleteGoogleCalendarOAuth(
	ctx context.Context,
	state, code string,
) (domain.CalendarOAuthResult, error) {
	if s.tokenRepo == nil || s.exporter == nil {
		return domain.CalendarOAuthResult{}, domain.ErrGoogleCalendarDisabled
	}
	if state == "" {
		return domain.CalendarOAuthResult{}, domain.ErrGoogleCalendarNotLinked
	}

	appointmentID, err := s.tokenRepo.GetPublicOAuthState(ctx, state)
	if err == nil {
		return s.completePublicGoogleCalendarOAuth(ctx, state, appointmentID, code)
	}
	if !errors.Is(err, domain.ErrGoogleCalendarNotLinked) {
		return domain.CalendarOAuthResult{}, err
	}

	userID, err := s.tokenRepo.GetOAuthState(ctx, state)
	if err != nil {
		return domain.CalendarOAuthResult{}, err
	}
	defer func() {
		_ = s.tokenRepo.DeleteOAuthState(ctx, state)
	}()

	if code == "" {
		return domain.CalendarOAuthResult{}, domain.ErrGoogleCalendarNotLinked
	}

	token, err := s.exporter.Exchange(ctx, code)
	if err != nil {
		return domain.CalendarOAuthResult{}, err
	}

	return domain.CalendarOAuthResult{}, s.tokenRepo.SaveToken(ctx, userID, token)
}

func (s *BookingService) PublicGoogleCalendarAuthURL(
	ctx context.Context,
	appointmentID int64,
	exportToken string,
) (string, error) {
	if appointmentID <= 0 || exportToken == "" {
		return "", domain.ErrGoogleCalendarNotLinked
	}
	if s.tokenRepo == nil || s.exporter == nil {
		return "", domain.ErrGoogleCalendarDisabled
	}

	tokenAppointmentID, err := s.tokenRepo.GetPublicExportToken(ctx, exportToken)
	if err != nil {
		return "", err
	}
	if tokenAppointmentID != appointmentID {
		return "", domain.ErrGoogleCalendarNotLinked
	}

	state, err := generateOAuthState()
	if err != nil {
		return "", err
	}
	if err := s.tokenRepo.SavePublicOAuthState(ctx, state, appointmentID); err != nil {
		return "", err
	}

	return s.exporter.AuthCodeURL(state)
}

func (s *BookingService) completePublicGoogleCalendarOAuth(
	ctx context.Context,
	state string,
	appointmentID int64,
	code string,
) (domain.CalendarOAuthResult, error) {
	defer func() {
		_ = s.tokenRepo.DeletePublicOAuthState(ctx, state)
	}()

	details, err := s.repoPostgres.GetAppointmentCalendarDetails(ctx, appointmentID)
	if err != nil {
		return domain.CalendarOAuthResult{}, err
	}
	result := domain.CalendarOAuthResult{RegistrationSlug: details.RegistrationSlug}

	// Empty code means the user declined access on the Google consent screen.
	if code == "" {
		return result, domain.ErrGoogleCalendarNotLinked
	}

	token, err := s.exporter.Exchange(ctx, code)
	if err != nil {
		return result, err
	}

	appointment, err := s.repoPostgres.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return result, err
	}

	_, _, err = s.exporter.ExportAppointment(ctx, token, &appointment)
	return result, err
}

func (s *BookingService) ExportAppointmentToGoogleCalendar(
	ctx context.Context,
	user *auth.AccessClaims,
	appointmentID int64,
) (domain.GoogleCalendarEvent, error) {
	if user == nil {
		return domain.GoogleCalendarEvent{}, liberrors.ErrUnauthorized
	}
	if s.tokenRepo == nil || s.exporter == nil {
		return domain.GoogleCalendarEvent{}, domain.ErrGoogleCalendarDisabled
	}

	appointment, err := s.GetAppointmentByID(ctx, user, appointmentID)
	if err != nil {
		return domain.GoogleCalendarEvent{}, err
	}

	token, err := s.tokenRepo.GetToken(ctx, user.UserID)
	if err != nil {
		return domain.GoogleCalendarEvent{}, err
	}

	event, refreshedToken, err := s.exporter.ExportAppointment(ctx, token, &appointment)
	if err != nil {
		return domain.GoogleCalendarEvent{}, err
	}
	if refreshedToken.RefreshToken != "" || refreshedToken.AccessToken != token.AccessToken {
		if err := s.tokenRepo.SaveToken(ctx, user.UserID, refreshedToken); err != nil {
			return domain.GoogleCalendarEvent{}, err
		}
	}

	return event, nil
}

func generateOAuthState() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
