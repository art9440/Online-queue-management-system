package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"Online-queue-management-system/libs/auth"
	"Online-queue-management-system/services/auth/internal/application/service"
	"Online-queue-management-system/services/auth/internal/domain"
)

type Handler struct {
	auth       *service.AuthService
	parser     *auth.TokenParser
	cookies    *CookieManager
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func NewHandler(
	authService *service.AuthService,
	parser *auth.TokenParser,
	cookies *CookieManager,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *Handler {
	return &Handler{
		auth:       authService,
		parser:     parser,
		cookies:    cookies,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/auth/login", h.handleLogin)
	mux.HandleFunc("/auth/refresh", h.handleRefresh)
	mux.HandleFunc("/auth/logout", h.handleLogout)
	mux.HandleFunc("/auth/me", h.handleMe)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, MessageResponse{Message: "method not allowed"})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, MessageResponse{Message: "invalid request"})
		return
	}

	if req.Login == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, MessageResponse{Message: "login and password are required"})
		return
	}

	tokens, err := h.auth.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrBadCredentials) {
			writeJSON(w, http.StatusUnauthorized, MessageResponse{Message: "bad credentials"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, MessageResponse{Message: "internal error"})
		return
	}

	h.cookies.SetAccess(w, tokens.AccessToken, h.accessTTL)
	h.cookies.SetRefresh(w, tokens.RefreshToken, h.refreshTTL)

	writeJSON(w, http.StatusOK, MessageResponse{Message: "ok"})
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, MessageResponse{Message: "method not allowed"})
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		writeJSON(w, http.StatusUnauthorized, MessageResponse{Message: "unauthorized"})
		return
	}

	tokens, err := h.auth.Refresh(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			writeJSON(w, http.StatusUnauthorized, MessageResponse{Message: "unauthorized"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, MessageResponse{Message: "internal error"})
		return
	}

	h.cookies.SetAccess(w, tokens.AccessToken, h.accessTTL)
	h.cookies.SetRefresh(w, tokens.RefreshToken, h.refreshTTL)

	writeJSON(w, http.StatusOK, MessageResponse{Message: "ok"})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, MessageResponse{Message: "method not allowed"})
		return
	}

	if cookie, err := r.Cookie("refresh_token"); err == nil && cookie.Value != "" {
		_ = h.auth.Logout(r.Context(), cookie.Value)
	}

	h.cookies.ClearAccess(w)
	h.cookies.ClearRefresh(w)

	writeJSON(w, http.StatusOK, MessageResponse{Message: "ok"})
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, MessageResponse{Message: "method not allowed"})
		return
	}

	cookie, err := r.Cookie("access_token")
	if err != nil || cookie.Value == "" {
		writeJSON(w, http.StatusUnauthorized, MessageResponse{Message: "unauthorized"})
		return
	}

	claims, err := h.parser.ParseAccessToken(cookie.Value)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, MessageResponse{Message: "unauthorized"})
		return
	}

	writeJSON(w, http.StatusOK, MeResponse{
		UserID:     claims.UserID,
		Login:      claims.Login,
		RoleID:     claims.RoleID,
		BusinessID: claims.BusinessID,
		BranchID:   claims.BranchID,
	})
}
