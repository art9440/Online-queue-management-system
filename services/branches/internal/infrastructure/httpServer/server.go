package httpserver

import (
	"Online-queue-management-system/libs/auth"
	"Online-queue-management-system/services/branches/internal/application/service"
	"Online-queue-management-system/services/branches/internal/domain"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
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
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	branches, err := s.svc.GetBranchesForUser(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, branches)
}

func (s *HttpServer) GetBranchClients(w http.ResponseWriter, r *http.Request) {
	user := auth.FromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	branchID, err := branchIDFromRequest(r)
	if err != nil {
		http.Error(w, "invalid branch id", http.StatusBadRequest)
		return
	}

	clients, err := s.svc.GetBranchClients(r.Context(), user, branchID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, clients)
}

func (s *HttpServer) GetBranchBookings(w http.ResponseWriter, r *http.Request) {
	user := auth.FromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	branchID, err := branchIDFromRequest(r)
	if err != nil {
		http.Error(w, "invalid branch id", http.StatusBadRequest)
		return
	}

	date, err := time.Parse(time.DateOnly, r.URL.Query().Get("date"))
	if err != nil {
		http.Error(w, "invalid date", http.StatusBadRequest)
		return
	}

	bookings, err := s.svc.GetBranchBookings(r.Context(), user, branchID, date)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, bookings)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func branchIDFromRequest(r *http.Request) (int64, error) {
	rawID := r.PathValue("id")
	if rawID == "" {
		rawID = strings.TrimPrefix(r.URL.Path, "/branches/")
		rawID = strings.Split(rawID, "/")[0]
	}

	branchID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, err
	}
	if branchID <= 0 {
		return 0, errors.New("invalid branch id")
	}

	return branchID, nil
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrForbidden):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, domain.ErrBranchNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
