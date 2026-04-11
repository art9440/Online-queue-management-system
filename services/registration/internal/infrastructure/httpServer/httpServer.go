package httpserver

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/registration/internal/application/service"
	"Online-queue-management-system/services/registration/internal/infrastructure/dto"
	"encoding/json"
	"fmt"
	"net/http"
)

type HttpServer struct {
	svc *service.RegistrationService
}

func NewHttpServer(svc *service.RegistrationService) *HttpServer {
	return &HttpServer{svc: svc}
}

func (h *HttpServer) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	ctx := r.Context()
	log := logger.From(ctx)

	log.Info("handling register request")
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	input := service.RegisterInput{
		Email:        req.Email,
		Password:     req.Password,
		BusinessName: req.BusinessName,
		BusinessType: req.BusinessType,
	}

	resp, err := h.svc.Register(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dto.Response{
		Status:         resp.Status,
		RegistrationID: resp.RegistrationID,
	})
}

func (h *HttpServer) ResendCode(w http.ResponseWriter, r *http.Request) {
	var req dto.ResendCodeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	input := service.ResendInput{
		RegistrationID: req.RegistrationID,
	}

	err := h.svc.ResendCode(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dto.Response{
		Status: "resended",
	})
}

func (h *HttpServer) Verify(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	input := service.VerifyInput{
		RegistrationID: req.RegistrationID,
		Code:           req.Code,
	}

	err := h.svc.Verify(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dto.Response{
		Status: "verified",
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{
		"error": msg,
	})
}

func RecoverMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("PANIC:", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		next(w, r)
	}
}
