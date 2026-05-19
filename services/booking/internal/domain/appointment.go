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
