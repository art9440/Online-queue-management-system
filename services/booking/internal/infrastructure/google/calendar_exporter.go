package google

import (
	"Online-queue-management-system/services/booking/internal/domain"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	googloauth "golang.org/x/oauth2/google"
)

const calendarScope = "https://www.googleapis.com/auth/calendar.events"

type CalendarExporter struct {
	config *oauth2.Config
}

type eventRequest struct {
	Summary     string    `json:"summary"`
	Description string    `json:"description,omitempty"`
	Start       eventTime `json:"start"`
	End         eventTime `json:"end"`
}

type eventTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone,omitempty"`
}

type eventResponse struct {
	ID       string `json:"id"`
	HTMLLink string `json:"htmlLink"`
}

func NewCalendarExporter(clientID, clientSecret, redirectURL string) *CalendarExporter {
	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return nil
	}

	return &CalendarExporter{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{calendarScope},
			Endpoint:     googloauth.Endpoint,
		},
	}
}

func (e *CalendarExporter) AuthCodeURL(state string) (string, error) {
	if e == nil || e.config == nil {
		return "", domain.ErrGoogleCalendarDisabled
	}

	return e.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce), nil
}

func (e *CalendarExporter) Exchange(ctx context.Context, code string) (domain.GoogleCalendarToken, error) {
	if e == nil || e.config == nil {
		return domain.GoogleCalendarToken{}, domain.ErrGoogleCalendarDisabled
	}

	token, err := e.config.Exchange(ctx, code)
	if err != nil {
		return domain.GoogleCalendarToken{}, err
	}

	return tokenFromOAuth(token), nil
}

func (e *CalendarExporter) ExportAppointment(
	ctx context.Context,
	storedToken domain.GoogleCalendarToken,
	appointment *domain.Appointment,
) (domain.GoogleCalendarEvent, domain.GoogleCalendarToken, error) {
	if e == nil || e.config == nil {
		return domain.GoogleCalendarEvent{}, domain.GoogleCalendarToken{}, domain.ErrGoogleCalendarDisabled
	}

	oauthToken := oauthTokenFromDomain(storedToken)
	tokenSource := e.config.TokenSource(ctx, oauthToken)
	freshToken, err := tokenSource.Token()
	if err != nil {
		return domain.GoogleCalendarEvent{}, domain.GoogleCalendarToken{}, err
	}

	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(freshToken))

	event, err := insertEvent(ctx, client, appointment)
	if err != nil {
		return domain.GoogleCalendarEvent{}, domain.GoogleCalendarToken{}, err
	}

	refreshedToken := tokenFromOAuth(freshToken)
	if refreshedToken.RefreshToken == "" {
		refreshedToken.RefreshToken = storedToken.RefreshToken
	}

	return event, refreshedToken, nil
}

func insertEvent(ctx context.Context, client *http.Client, appointment *domain.Appointment) (domain.GoogleCalendarEvent, error) {
	payload := eventRequest{
		Summary:     fmt.Sprintf("Appointment #%d", appointment.ID),
		Description: eventDescription(appointment),
		Start: eventTime{
			DateTime: appointment.StartTime.Format("2006-01-02T15:04:05Z07:00"),
		},
		End: eventTime{
			DateTime: appointment.EndTime.Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return domain.GoogleCalendarEvent{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://www.googleapis.com/calendar/v3/calendars/primary/events",
		bytes.NewReader(raw),
	)
	if err != nil {
		return domain.GoogleCalendarEvent{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return domain.GoogleCalendarEvent{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return domain.GoogleCalendarEvent{}, fmt.Errorf("google calendar returned status %d", resp.StatusCode)
	}

	var response eventResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return domain.GoogleCalendarEvent{}, err
	}

	return domain.GoogleCalendarEvent{
		ID:       response.ID,
		HTMLLink: response.HTMLLink,
	}, nil
}

func eventDescription(appointment *domain.Appointment) string {
	if appointment.Comment == nil || *appointment.Comment == "" {
		return "Exported from Online Queue."
	}
	return "Comment: " + *appointment.Comment
}

func oauthTokenFromDomain(token domain.GoogleCalendarToken) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}
}

func tokenFromOAuth(token *oauth2.Token) domain.GoogleCalendarToken {
	return domain.GoogleCalendarToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}
}
