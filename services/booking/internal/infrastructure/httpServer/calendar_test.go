package httpserver

import (
	"Online-queue-management-system/services/booking/internal/application/service"
	"Online-queue-management-system/services/booking/internal/domain"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testFrontendURL = "http://localhost:5173"

func TestDownloadPublicAppointmentICS(t *testing.T) {
	tokens := &fakeTokenRepo{exportTokens: map[string]int64{"good": 7, "other": 8}}
	srv := NewHttpServer(service.New(&fakeRepo{}, tokens, nil, nil), testFrontendURL)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /public/appointments/{id}/calendar.ics", srv.DownloadPublicAppointmentICS)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"valid token", "/public/appointments/7/calendar.ics?token=good", http.StatusOK},
		{"unknown token", "/public/appointments/7/calendar.ics?token=bad", http.StatusNotFound},
		{"token of another appointment", "/public/appointments/7/calendar.ics?token=other", http.StatusNotFound},
		{"missing token", "/public/appointments/7/calendar.ics", http.StatusNotFound},
		{"invalid id", "/public/appointments/abc/calendar.ics?token=good", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/calendar") {
				t.Errorf("Content-Type = %q, want text/calendar", got)
			}
			if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="appointment-7.ics"` {
				t.Errorf("Content-Disposition = %q", got)
			}
			if !strings.Contains(rec.Body.String(), "SUMMARY:Consultation") {
				t.Errorf("body does not contain the event: %s", rec.Body.String())
			}
		})
	}
}

func TestGoogleCalendarCallbackRedirectsToFrontend(t *testing.T) {
	tests := []struct {
		name     string
		exporter service.CalendarExporter
		query    string
		want     string
	}{
		{
			name:     "integration disabled",
			exporter: nil,
			query:    "?state=public-state&code=abc",
			want:     testFrontendURL + "/admin?calendar=error",
		},
		{
			name:     "public flow succeeded",
			exporter: &fakeExporter{},
			query:    "?state=public-state&code=abc",
			want:     testFrontendURL + "/public/demo-business?calendar=success",
		},
		{
			name:     "public flow declined by user",
			exporter: &fakeExporter{},
			query:    "?state=public-state&error=access_denied",
			want:     testFrontendURL + "/public/demo-business?calendar=error",
		},
		{
			name:     "unknown state",
			exporter: &fakeExporter{},
			query:    "?state=unknown&code=abc",
			want:     testFrontendURL + "/admin?calendar=error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := &fakeTokenRepo{publicStates: map[string]int64{"public-state": 7}}
			srv := NewHttpServer(service.New(&fakeRepo{}, tokens, tt.exporter, nil), testFrontendURL)

			rec := httptest.NewRecorder()
			srv.GoogleCalendarCallback(rec, httptest.NewRequest(http.MethodGet, "/google-calendar/callback"+tt.query, nil))

			if rec.Code != http.StatusFound {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
			}
			if got := rec.Header().Get("Location"); got != tt.want {
				t.Errorf("Location = %q, want %q", got, tt.want)
			}
		})
	}
}

type fakeRepo struct {
	service.BookingRepository
}

func (r *fakeRepo) GetAppointmentCalendarDetails(
	_ context.Context,
	appointmentID int64,
) (domain.AppointmentCalendarDetails, error) {
	start := time.Date(2030, 1, 10, 3, 0, 0, 0, time.UTC)
	return domain.AppointmentCalendarDetails{
		AppointmentID:    appointmentID,
		RegistrationSlug: "demo-business",
		ServiceName:      "Consultation",
		EmployeeName:     "Anna",
		EmployeeSurname:  "Petrova",
		BranchName:       "Central Branch",
		StartTime:        start,
		EndTime:          start.Add(30 * time.Minute),
	}, nil
}

func (r *fakeRepo) GetAppointmentByID(_ context.Context, appointmentID int64) (domain.Appointment, error) {
	return domain.Appointment{ID: appointmentID}, nil
}

type fakeTokenRepo struct {
	service.CalendarTokenRepository
	exportTokens map[string]int64
	publicStates map[string]int64
}

func (r *fakeTokenRepo) GetPublicExportToken(_ context.Context, token string) (int64, error) {
	if id, ok := r.exportTokens[token]; ok {
		return id, nil
	}
	return 0, domain.ErrGoogleCalendarNotLinked
}

func (r *fakeTokenRepo) GetPublicOAuthState(_ context.Context, state string) (int64, error) {
	if id, ok := r.publicStates[state]; ok {
		return id, nil
	}
	return 0, domain.ErrGoogleCalendarNotLinked
}

func (r *fakeTokenRepo) DeletePublicOAuthState(context.Context, string) error {
	return nil
}

func (r *fakeTokenRepo) GetOAuthState(context.Context, string) (int64, error) {
	return 0, domain.ErrGoogleCalendarNotLinked
}

type fakeExporter struct{}

func (e *fakeExporter) AuthCodeURL(string) (string, error) {
	return "https://accounts.google.com/o/oauth2/auth", nil
}

func (e *fakeExporter) Exchange(context.Context, string) (domain.GoogleCalendarToken, error) {
	return domain.GoogleCalendarToken{AccessToken: "token"}, nil
}

func (e *fakeExporter) ExportAppointment(
	context.Context,
	domain.GoogleCalendarToken,
	*domain.Appointment,
) (domain.GoogleCalendarEvent, domain.GoogleCalendarToken, error) {
	return domain.GoogleCalendarEvent{ID: "event"}, domain.GoogleCalendarToken{}, nil
}
