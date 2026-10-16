package domain

import "time"

type Appointment struct {
	ID              int64
	BranchID        int64
	Client          Client
	EmployeeID      int64
	EmployeeName    string
	EmployeeSurname string
	ServiceID       int64
	ServiceName     string
	StartTime       time.Time
	EndTime         time.Time
	Status          string
	Comment         *string
	CreatedAt       time.Time
}
