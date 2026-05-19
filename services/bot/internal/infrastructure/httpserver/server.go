package httpserver

import (
	"Online-queue-management-system/services/bot/internal/application/service"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	bot     *service.Bot
	enabled bool
}

type SendNotificationRequest struct {
	Phone       string `json:"phone"`
	Username    string `json:"username"`
	Business    string `json:"business"`
	Service     string `json:"service"`
	Branch      string `json:"branch"`
	Employee    string `json:"employee"`
	StartTime   string `json:"start_time"`
	Description string `json:"description"`
}

func New(bot *service.Bot, enabled bool) *Server {
	return &Server{
		bot:     bot,
		enabled: enabled,
	}
}

func (s *Server) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) SendNotification(w http.ResponseWriter, r *http.Request) {
	if !s.enabled {
		http.Error(w, "telegram bot is disabled", http.StatusServiceUnavailable)
		return
	}

	var req SendNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	notification, err := notificationFromRequest(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.bot.SendNotification(r.Context(), notification); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func notificationFromRequest(req *SendNotificationRequest) (*service.Notification, error) {
	phone := strings.TrimSpace(req.Phone)
	username := strings.TrimPrefix(strings.TrimSpace(req.Username), "@")
	if phone == "" && username == "" {
		return nil, errMissingRecipient{}
	}

	var startTime time.Time
	if req.StartTime != "" {
		parsed, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			return nil, errInvalidStartTime{}
		}
		startTime = parsed
	}

	return &service.Notification{
		Recipient: service.Recipient{
			Phone:    phone,
			Username: strings.ToLower(username),
		},
		Business:    req.Business,
		Service:     req.Service,
		Branch:      req.Branch,
		Employee:    req.Employee,
		StartTime:   startTime,
		Description: req.Description,
	}, nil
}

type errMissingRecipient struct{}

func (errMissingRecipient) Error() string {
	return "phone or username is required"
}

type errInvalidStartTime struct{}

func (errInvalidStartTime) Error() string {
	return "start_time must be RFC3339"
}
