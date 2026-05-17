package httpserver

import (
	"Online-queue-management-system/libs/auth"
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/branches/internal/application/service"
	"Online-queue-management-system/services/branches/internal/domain"
	"Online-queue-management-system/services/branches/internal/infrastructure/dto"
	"encoding/json"
	"net/http"
	"strconv"
)

type HttpServer struct {
	svc *service.BranchesService
}

func NewHttpServer(svc *service.BranchesService) *HttpServer {
	return &HttpServer{svc: svc}
}

func (s *HttpServer) GetBranches(w http.ResponseWriter, r *http.Request) {

	user := auth.FromContext(r.Context())
	if user == nil {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	branches, err := s.svc.GetBranchesForUser(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.BranchesFromDomain(branches)

	writeJSON(w, http.StatusOK, response)
}

func (s *HttpServer) GetBranchEmployees(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.From(ctx)

	log.Info("get branch employees request started")

	user := auth.FromContext(ctx)
	if user == nil {
		log.Warn("unauthorized get branch employees request")
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	branchIDStr := r.PathValue("id")

	branchID, err := strconv.ParseInt(branchIDStr, 10, 64)
	if err != nil || branchID <= 0 {
		log.Warn(
			"invalid branch id",
			"branch_id_raw", branchIDStr,
			"err", err,
		)

		http.Error(w, domain.ErrInvalidBranchID.Error(), http.StatusBadRequest)
		return
	}

	log.Info(
		"getting employees for branch",
		"user_id", user.UserID,
		"role_id", user.RoleID,
		"business_id", user.BusinessID,
		"branch_id", branchID,
	)

	employees, err := s.svc.GetEmployeesForBranch(ctx, user, branchID)
	if err != nil {
		log.Error(
			"failed to get employees for branch",
			"user_id", user.UserID,
			"role_id", user.RoleID,
			"business_id", user.BusinessID,
			"branch_id", branchID,
			"err", err,
		)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.EmployeesFromDomain(employees)

	log.Info(
		"employees for branch successfully received",
		"user_id", user.UserID,
		"business_id", user.BusinessID,
		"branch_id", branchID,
		"employees_count", len(response),
	)

	writeJSON(w, http.StatusOK, response)
}

func (s *HttpServer) GetServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.From(ctx)

	log.Info("get services request started")

	user := auth.FromContext(ctx)
	if user == nil {
		log.Warn("unauthorized get services request")
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	services, err := s.svc.GetServicesForBusiness(ctx, user)
	if err != nil {
		log.Error(
			"failed to get services",
			"user_id", user.UserID,
			"business_id", user.BusinessID,
			"err", err,
		)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.ServicesFromDomain(services)

	log.Info(
		"services successfully received",
		"user_id", user.UserID,
		"business_id", user.BusinessID,
		"services_count", len(response),
	)

	writeJSON(w, http.StatusOK, response)
}

func (s *HttpServer) GetBranchesWithService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.From(ctx)

	log.Info("get branches with service request started")

	user := auth.FromContext(ctx)
	if user == nil {
		log.Warn("unauthorized get branches with service request")
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	serviceIDStr := r.PathValue("serviceId")

	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil || serviceID <= 0 {
		log.Warn(
			"invalid service id",
			"service_id_raw", serviceIDStr,
			"err", err,
		)

		http.Error(w, domain.ErrInvalidServiceID.Error(), http.StatusBadRequest)
		return
	}

	log.Info(
		"getting branches with service",
		"user_id", user.UserID,
		"business_id", user.BusinessID,
		"service_id", serviceID,
	)

	branches, err := s.svc.GetBranchesWithServiceForBusiness(ctx, user, serviceID)
	if err != nil {
		log.Error(
			"failed to get branches with service",
			"user_id", user.UserID,
			"business_id", user.BusinessID,
			"service_id", serviceID,
			"err", err,
		)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.BranchesFromDomain(branches)

	log.Info(
		"branches with service successfully received",
		"user_id", user.UserID,
		"business_id", user.BusinessID,
		"service_id", serviceID,
		"branches_count", len(response),
	)

	writeJSON(w, http.StatusOK, response)
}

func (s *HttpServer) GetEmployeesForService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.From(ctx)

	log.Info("get employees for service request started")

	user := auth.FromContext(ctx)
	if user == nil {
		log.Warn("unauthorized get employees for service request")
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	serviceIDStr := r.PathValue("serviceId")
	branchIDStr := r.PathValue("branchId")

	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil || serviceID <= 0 {
		log.Warn(
			"invalid service id",
			"service_id_raw", serviceIDStr,
			"err", err,
		)

		http.Error(w, domain.ErrInvalidServiceID.Error(), http.StatusBadRequest)
		return
	}

	branchID, err := strconv.ParseInt(branchIDStr, 10, 64)
	if err != nil || branchID <= 0 {
		log.Warn(
			"invalid branch id",
			"branch_id_raw", branchIDStr,
			"err", err,
		)

		http.Error(w, domain.ErrInvalidBranchID.Error(), http.StatusBadRequest)
		return
	}

	log.Info(
		"getting employees for service and branch",
		"user_id", user.UserID,
		"business_id", user.BusinessID,
		"service_id", serviceID,
		"branch_id", branchID,
	)

	employees, err := s.svc.GetEmployeesForServiceAndBranch(ctx, user, serviceID, branchID)
	if err != nil {
		log.Error(
			"failed to get employees for service and branch",
			"user_id", user.UserID,
			"business_id", user.BusinessID,
			"service_id", serviceID,
			"branch_id", branchID,
			"err", err,
		)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.EmployeesFromDomain(employees)

	log.Info(
		"employees for service successfully received",
		"user_id", user.UserID,
		"business_id", user.BusinessID,
		"service_id", serviceID,
		"branch_id", branchID,
		"employees_count", len(response),
	)

	writeJSON(w, http.StatusOK, response)
}

// Public handlers for unauthenticated clients via registration slug

func (s *HttpServer) GetPublicServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.From(ctx)

	log.Info("get public services request started")

	registrationSlug := r.PathValue("registrationSlug")

	if registrationSlug == "" {
		log.Warn("empty registration slug")
		http.Error(w, domain.ErrInvalidRegistrationSlug.Error(), http.StatusBadRequest)
		return
	}

	log.Info("getting services for registration slug", "slug", registrationSlug)

	services, err := s.svc.GetPublicServicesForSlug(ctx, registrationSlug)
	if err != nil {
		log.Error(
			"failed to get public services",
			"slug", registrationSlug,
			"err", err,
		)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.ServicesFromDomain(services)

	log.Info(
		"public services successfully received",
		"slug", registrationSlug,
		"services_count", len(response),
	)

	writeJSON(w, http.StatusOK, response)
}

func (s *HttpServer) GetPublicBranchesWithService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.From(ctx)

	log.Info("get public branches with service request started")

	registrationSlug := r.PathValue("registrationSlug")
	serviceIDStr := r.PathValue("serviceId")

	if registrationSlug == "" {
		log.Warn("empty registration slug")
		http.Error(w, domain.ErrInvalidRegistrationSlug.Error(), http.StatusBadRequest)
		return
	}

	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil || serviceID <= 0 {
		log.Warn(
			"invalid service id",
			"service_id_raw", serviceIDStr,
			"err", err,
		)

		http.Error(w, domain.ErrInvalidServiceID.Error(), http.StatusBadRequest)
		return
	}

	log.Info(
		"getting branches with service",
		"slug", registrationSlug,
		"service_id", serviceID,
	)

	branches, err := s.svc.GetPublicBranchesWithServiceForSlug(ctx, registrationSlug, serviceID)
	if err != nil {
		log.Error(
			"failed to get public branches with service",
			"slug", registrationSlug,
			"service_id", serviceID,
			"err", err,
		)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.BranchesFromDomain(branches)

	log.Info(
		"public branches with service successfully received",
		"slug", registrationSlug,
		"service_id", serviceID,
		"branches_count", len(response),
	)

	writeJSON(w, http.StatusOK, response)
}

func (s *HttpServer) GetPublicEmployeesForService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.From(ctx)

	log.Info("get public employees for service request started")

	registrationSlug := r.PathValue("registrationSlug")
	serviceIDStr := r.PathValue("serviceId")
	branchIDStr := r.PathValue("branchId")

	if registrationSlug == "" {
		log.Warn("empty registration slug")
		http.Error(w, domain.ErrInvalidRegistrationSlug.Error(), http.StatusBadRequest)
		return
	}

	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil || serviceID <= 0 {
		log.Warn(
			"invalid service id",
			"service_id_raw", serviceIDStr,
			"err", err,
		)

		http.Error(w, domain.ErrInvalidServiceID.Error(), http.StatusBadRequest)
		return
	}

	branchID, err := strconv.ParseInt(branchIDStr, 10, 64)
	if err != nil || branchID <= 0 {
		log.Warn(
			"invalid branch id",
			"branch_id_raw", branchIDStr,
			"err", err,
		)

		http.Error(w, domain.ErrInvalidBranchID.Error(), http.StatusBadRequest)
		return
	}

	log.Info(
		"getting employees for service and branch",
		"slug", registrationSlug,
		"service_id", serviceID,
		"branch_id", branchID,
	)

	employees, err := s.svc.GetPublicEmployeesForServiceAndBranchSlug(ctx, registrationSlug, serviceID, branchID)
	if err != nil {
		log.Error(
			"failed to get public employees for service and branch",
			"slug", registrationSlug,
			"service_id", serviceID,
			"branch_id", branchID,
			"err", err,
		)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.EmployeesFromDomain(employees)

	log.Info(
		"public employees for service successfully received",
		"slug", registrationSlug,
		"service_id", serviceID,
		"branch_id", branchID,
		"employees_count", len(response),
	)

	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
