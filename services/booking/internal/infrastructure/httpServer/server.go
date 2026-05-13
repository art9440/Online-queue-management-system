package httpserver

import (
	"Online-queue-management-system/libs/auth"
	liberrors "Online-queue-management-system/libs/errors"
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/booking/internal/application/service"
	"Online-queue-management-system/services/booking/internal/domain"
	"Online-queue-management-system/services/booking/internal/infrastructure/dto"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

type HttpServer struct {
	svc *service.BookingService
}

func NewHttpServer(svc *service.BookingService) *HttpServer {
	return &HttpServer{svc: svc}
}

func (s *HttpServer) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.From(ctx)

	log.Info("create appointment request started")

	var req dto.CreateAppointmentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("failed to decode create appointment request", "err", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	log.Info(
		"client data received for appointment",
		"client_name", req.Client.Name,
		"client_surname", req.Client.Surname,
		"client_email", stringPtrValue(req.Client.Email),
		"client_phone", req.Client.Phone,
		"client_tg_username", stringPtrValue(req.Client.TgUsername),
	)

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		log.Warn(
			"invalid appointment start_time",
			"start_time", req.StartTime,
			"client_name", req.Client.Name,
			"client_surname", req.Client.Surname,
			"err", err,
		)

		http.Error(w, "invalid start_time format, expected RFC3339", http.StatusBadRequest)
		return
	}

	input := domain.CreateAppointmentInput{
		Client: domain.ClientInput{
			Email:      req.Client.Email,
			Phone:      req.Client.Phone,
			Name:       req.Client.Name,
			Surname:    req.Client.Surname,
			TgUsername: req.Client.TgUsername,
		},
		BranchID:   req.BranchID,
		EmployeeID: req.EmployeeID,
		ServiceID:  req.ServiceID,
		StartTime:  startTime,
		Comment:    req.Comment,
	}

	log.Info(
		"creating appointment for client",
		"client_name", req.Client.Name,
		"client_surname", req.Client.Surname,
		"client_email", stringPtrValue(req.Client.Email),
		"client_phone", req.Client.Phone,
		"client_tg_username", stringPtrValue(req.Client.TgUsername),
		"branch_id", req.BranchID,
		"employee_id", req.EmployeeID,
		"service_id", req.ServiceID,
		"start_time", startTime,
	)

	appointment, err := s.svc.CreateAppointment(ctx, input)
	if err != nil {
		log.Error(
			"failed to create appointment for client",
			"client_name", req.Client.Name,
			"client_surname", req.Client.Surname,
			"client_email", stringPtrValue(req.Client.Email),
			"client_phone", req.Client.Phone,
			"client_tg_username", stringPtrValue(req.Client.TgUsername),
			"branch_id", req.BranchID,
			"employee_id", req.EmployeeID,
			"service_id", req.ServiceID,
			"err", err,
		)

		switch {
		case errors.Is(err, liberrors.ErrUnauthorized):
			http.Error(w, err.Error(), http.StatusUnauthorized)

		case errors.Is(err, liberrors.ErrForbidden):
			http.Error(w, err.Error(), http.StatusForbidden)

		case errors.Is(err, domain.ErrTimeSlotBusy):
			http.Error(w, err.Error(), http.StatusConflict)

		case errors.Is(err, domain.ErrAppointmentNotAvailable):
			http.Error(w, err.Error(), http.StatusConflict)

		case errors.Is(err, liberrors.ErrInvalidBranchID),
			errors.Is(err, liberrors.ErrInvalidEmployeeID),
			errors.Is(err, liberrors.ErrInvalidServiceID),
			errors.Is(err, domain.ErrInvalidClient),
			errors.Is(err, domain.ErrInvalidClientContact),
			errors.Is(err, domain.ErrInvalidStartTime):
			http.Error(w, err.Error(), http.StatusBadRequest)

		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	response := dto.CreateAppointmentResponseFromDomain(appointment)

	log.Info(
		"appointment successfully created for client",
		"appointment_id", appointment.AppointmentID,
		"client_id", appointment.ClientID,
		"client_name", req.Client.Name,
		"client_surname", req.Client.Surname,
		"client_email", stringPtrValue(req.Client.Email),
		"client_phone", req.Client.Phone,
		"client_tg_username", stringPtrValue(req.Client.TgUsername),
		"branch_id", appointment.BranchID,
		"employee_id", appointment.EmployeeID,
		"service_id", appointment.ServiceID,
		"start_time", appointment.StartTime,
		"end_time", appointment.EndTime,
		"status", appointment.Status,
	)

	writeJSON(w, http.StatusCreated, response)
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func (s *HttpServer) GetAppointmentsByEmployeeID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.From(ctx)

	log.Info("get appointments by employee id request started")

	user := auth.FromContext(ctx)
	if user == nil {
		log.Warn("unauthorized get appointments by employee id request")
		http.Error(w, liberrors.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	employeeIDStr := r.PathValue("id")

	employeeID, err := strconv.ParseInt(employeeIDStr, 10, 64)
	if err != nil || employeeID <= 0 {
		log.Warn(
			"invalid employee id",
			"employee_id_raw", employeeIDStr,
			"err", err,
		)

		http.Error(w, liberrors.ErrInvalidEmployeeID.Error(), http.StatusBadRequest)
		return
	}

	log.Info(
		"getting appointments by employee id",
		"user_id", user.UserID,
		"role_id", user.RoleID,
		"role_name", user.RoleName,
		"business_id", user.BusinessID,
		"employee_id", employeeID,
	)

	appointments, err := s.svc.GetAppointmentsByEmployeeID(ctx, user, employeeID)
	if err != nil {
		log.Error(
			"failed to get appointments by employee id",
			"user_id", user.UserID,
			"role_id", user.RoleID,
			"role_name", user.RoleName,
			"business_id", user.BusinessID,
			"employee_id", employeeID,
			"err", err,
		)

		switch {
		case errors.Is(err, liberrors.ErrUnauthorized):
			http.Error(w, err.Error(), http.StatusUnauthorized)

		case errors.Is(err, liberrors.ErrForbidden):
			http.Error(w, err.Error(), http.StatusForbidden)

		case errors.Is(err, liberrors.ErrInvalidEmployeeID):
			http.Error(w, err.Error(), http.StatusBadRequest)

		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	response := dto.AppointmentsFromDomain(appointments)

	log.Info(
		"appointments by employee id successfully received",
		"user_id", user.UserID,
		"business_id", user.BusinessID,
		"employee_id", employeeID,
		"appointments_count", len(response),
	)

	writeJSON(w, http.StatusOK, response)
}

func (s *HttpServer) GetAppointmentByID(w http.ResponseWriter, r *http.Request) {

}

func (s *HttpServer) CancelAppointment(w http.ResponseWriter, r *http.Request) {

}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
