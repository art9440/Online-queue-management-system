package dto

import (
	"Online-queue-management-system/services/booking/internal/domain"
	"time"
)

type AppointmentResponse struct {
	ID         int64   `json:"id"`
	ClientID   int64   `json:"client_id"`
	BranchID   int64   `json:"branch_id"`
	EmployeeID int64   `json:"employee_id"`
	ServiceID  int64   `json:"service_id"`
	StartTime  string  `json:"start_time"`
	EndTime    string  `json:"end_time"`
	Status     string  `json:"status"`
	Comment    *string `json:"comment,omitempty"`
}

type AvailableSlotResponse struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Timezone  string `json:"timezone"`
}

func AvailableSlotFromDomain(slot domain.AvailableSlot) AvailableSlotResponse {
	return AvailableSlotResponse{
		StartTime: slot.StartTime.Format(time.RFC3339),
		EndTime:   slot.EndTime.Format(time.RFC3339),
		Timezone:  slot.Timezone,
	}
}

func AvailableSlotsFromDomain(slots []domain.AvailableSlot) []AvailableSlotResponse {
	result := make([]AvailableSlotResponse, 0, len(slots))

	for _, slot := range slots {
		result = append(result, AvailableSlotFromDomain(slot))
	}

	return result
}

func AppointmentFromDomain(appointment domain.Appointment) AppointmentResponse {
	return AppointmentResponse{
		ID:         appointment.ID,
		ClientID:   appointment.ClientID,
		BranchID:   appointment.BranchID,
		EmployeeID: appointment.EmployeeID,
		ServiceID:  appointment.ServiceID,
		StartTime:  appointment.StartTime.Format(time.RFC3339),
		EndTime:    appointment.EndTime.Format(time.RFC3339),
		Status:     string(appointment.Status),
		Comment:    appointment.Comment,
	}
}

func AppointmentsFromDomain(appointments []domain.Appointment) []AppointmentResponse {
	result := make([]AppointmentResponse, 0, len(appointments))

	for _, appointment := range appointments {
		result = append(result, AppointmentFromDomain(appointment))
	}

	return result
}

type CreateAppointmentResponse struct {
	AppointmentID int64 `json:"appointment_id"`
	ClientID      int64 `json:"client_id"`

	Branch BranchResponse `json:"branch"`

	Employee EmployeeResponse `json:"employee"`

	Service ServiceResponse `json:"service"`

	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	Status    string  `json:"status"`
	Comment   *string `json:"comment,omitempty"`
}

type BranchResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type EmployeeResponse struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
}

type ServiceResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func CreateAppointmentResponseFromDomain(
	output domain.CreateAppointmentOutput,
) CreateAppointmentResponse {
	return CreateAppointmentResponse{
		AppointmentID: output.AppointmentID,
		ClientID:      output.ClientID,

		Branch: BranchResponse{
			ID:   output.BranchID,
			Name: output.BranchName,
		},

		Employee: EmployeeResponse{
			ID:      output.EmployeeID,
			Name:    output.EmployeeName,
			Surname: output.EmployeeSurname,
		},

		Service: ServiceResponse{
			ID:   output.ServiceID,
			Name: output.ServiceName,
		},

		StartTime: output.StartTime.Format(time.RFC3339),
		EndTime:   output.EndTime.Format(time.RFC3339),
		Status:    string(output.Status),
		Comment:   output.Comment,
	}
}
