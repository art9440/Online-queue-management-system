package http

import (
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"time"

	"Online-queue-management-system/services/auth/internal/application/service"
	"Online-queue-management-system/services/auth/internal/domain"
)

type Handler struct {
	auth       *service.AuthService
	cookies    *CookieManager
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func NewHandler(
	auth *service.AuthService,
	cookies *CookieManager,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *Handler {
	return &Handler{
		auth:       auth,
		cookies:    cookies,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (h *Handler) Register(mux *stdhttp.ServeMux) {
	mux.HandleFunc("/auth/login", h.handleLogin)
	mux.HandleFunc("/auth/refresh", h.handleRefresh)
	mux.HandleFunc("/auth/logout", h.handleLogout)
	mux.HandleFunc("/auth/me", h.handleMe)
}

func (h *Handler) handleLogin(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodPost {
		writeJSON(w, stdhttp.StatusMethodNotAllowed, MessageResponse{Message: "method not allowed"})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, stdhttp.StatusBadRequest, MessageResponse{Message: "invalid request"})
		return
	}

	if req.Login == "" || req.Password == "" {
		writeJSON(w, stdhttp.StatusBadRequest, MessageResponse{Message: "login and password are required"})
		return
	}

	tokens, err := h.auth.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrBadCredentials) {
			writeJSON(w, stdhttp.StatusUnauthorized, MessageResponse{Message: "bad credentials"})
			return
		}
		writeJSON(w, stdhttp.StatusInternalServerError, MessageResponse{Message: "internal error"})
		return
	}

	h.cookies.SetAccess(w, tokens.AccessToken, h.accessTTL)
	h.cookies.SetRefresh(w, tokens.RefreshToken, h.refreshTTL)

	writeJSON(w, stdhttp.StatusOK, MessageResponse{Message: "ok"})
}

func (h *Handler) handleRefresh(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodPost {
		writeJSON(w, stdhttp.StatusMethodNotAllowed, MessageResponse{Message: "method not allowed"})
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		writeJSON(w, stdhttp.StatusUnauthorized, MessageResponse{Message: "unauthorized"})
		return
	}

	tokens, err := h.auth.Refresh(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			writeJSON(w, stdhttp.StatusUnauthorized, MessageResponse{Message: "unauthorized"})
			return
		}
		writeJSON(w, stdhttp.StatusInternalServerError, MessageResponse{Message: "internal error"})
		return
	}

	h.cookies.SetAccess(w, tokens.AccessToken, h.accessTTL)
	h.cookies.SetRefresh(w, tokens.RefreshToken, h.refreshTTL)

	writeJSON(w, stdhttp.StatusOK, MessageResponse{Message: "ok"})
}

func (h *Handler) handleLogout(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodPost {
		writeJSON(w, stdhttp.StatusMethodNotAllowed, MessageResponse{Message: "method not allowed"})
		return
	}

	if cookie, err := r.Cookie("refresh_token"); err == nil && cookie.Value != "" {
		_ = h.auth.Logout(r.Context(), cookie.Value)
	}

	h.cookies.ClearAccess(w)
	h.cookies.ClearRefresh(w)

	writeJSON(w, stdhttp.StatusOK, MessageResponse{Message: "ok"})
}

func (h *Handler) handleMe(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodGet {
		writeJSON(w, stdhttp.StatusMethodNotAllowed, MessageResponse{Message: "method not allowed"})
		return
	}

	cookie, err := r.Cookie("access_token")
	if err != nil || cookie.Value == "" {
		writeJSON(w, stdhttp.StatusUnauthorized, MessageResponse{Message: "unauthorized"})
		return
	}

	claims, err := h.auth.Me(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			writeJSON(w, stdhttp.StatusUnauthorized, MessageResponse{Message: "unauthorized"})
			return
		}
		writeJSON(w, stdhttp.StatusInternalServerError, MessageResponse{Message: "internal error"})
		return
	}

	writeJSON(w, stdhttp.StatusOK, MeResponse{
		UserID:     claims.UserID,
		Login:      claims.Login,
		RoleID:     claims.RoleID,
		BusinessID: claims.BusinessID,
		BranchID:   claims.BranchID,
	})
}