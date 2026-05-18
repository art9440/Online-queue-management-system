package domain

import "time"

type Booking struct {
	ID              int64     `json:"id"`
	BranchID        int64     `json:"branch_id"`
	Client          Client    `json:"client"`
	EmployeeID      int64     `json:"employee_id"`
	EmployeeName    string    `json:"employee_name"`
	EmployeeSurname string    `json:"employee_surname"`
	ServiceID       int64     `json:"service_id"`
	ServiceName     string    `json:"service_name"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Status          string    `json:"status"`
	Comment         *string   `json:"comment,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}
