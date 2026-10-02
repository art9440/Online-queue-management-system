package domain

import "time"

type AppointmentStatus string

const (
	AppointmentStatusPending   AppointmentStatus = "pending"
	AppointmentStatusConfirmed AppointmentStatus = "confirmed"
	AppointmentStatusCompleted AppointmentStatus = "completed"
	AppointmentStatusCancelled AppointmentStatus = "cancelled"
)

type Appointment struct {
	ID         int64
	ClientID   int64
	BusinessID int64
	BranchID   int64
	EmployeeID int64
	ServiceID  int64
	StartTime  time.Time
	EndTime    time.Time
	Status     AppointmentStatus
	Comment    *string
}

type CreateAppointmentInput struct {
	Client ClientInput

	RegistrationSlug string
	BranchID         int64
	EmployeeID       int64
	ServiceID        int64
	StartTime        time.Time

	Comment *string
}

type ClientInput struct {
	Email      *string
	Phone      string
	Name       string
	Surname    string
	TgUsername *string
}

type CreateAppointmentOutput struct {
	AppointmentID int64
	ClientID      int64

	BranchID   int64
	BranchName string

	EmployeeID      int64
	EmployeeName    string
	EmployeeSurname string

	ServiceID   int64
	ServiceName string

	StartTime time.Time
	EndTime   time.Time
	Status    AppointmentStatus
	Comment   *string

	GoogleCalendarExportURL string
	CalendarICSURL          string
}

// AppointmentCalendarDetails holds what is needed to describe an appointment as a calendar event.
type AppointmentCalendarDetails struct {
	AppointmentID    int64
	RegistrationSlug string

	ServiceName     string
	EmployeeName    string
	EmployeeSurname string
	BranchName      string
	BranchAddress   string
	Timezone        string

	StartTime time.Time
	EndTime   time.Time
	Comment   *string
}

// CalendarOAuthResult tells where to send the user after the Google OAuth callback.
// RegistrationSlug is empty for the dashboard (non-public) flow.
type CalendarOAuthResult struct {
	RegistrationSlug string
}

type AvailableSlotsInput struct {
	RegistrationSlug string
	ServiceID        int64
	BranchID         int64
	EmployeeID       int64
	Date             time.Time
}

type AvailableSlot struct {
	StartTime time.Time
	EndTime   time.Time
	Timezone  string
}

type GoogleCalendarToken struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expiry       time.Time
}

type GoogleCalendarEvent struct {
	ID       string
	HTMLLink string
}
