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

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
