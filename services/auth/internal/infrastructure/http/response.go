package http

import (
	"encoding/json"
	"net/http"
)

type MessageResponse struct {
	Message string `json:"message"`
}

type MeResponse struct {
	UserID     int64  `json:"user_id"`
	Login      string `json:"login"`
	RoleID     int64  `json:"role_id"`
	BusinessID int64  `json:"business_id"`
	BranchID   *int64 `json:"branch_id,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}