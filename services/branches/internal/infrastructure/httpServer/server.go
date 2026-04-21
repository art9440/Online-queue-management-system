package httpserver

import (
	"Online-queue-management-system/libs/auth"
	"Online-queue-management-system/services/branches/internal/application/service"
	"encoding/json"
	"net/http"
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

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
