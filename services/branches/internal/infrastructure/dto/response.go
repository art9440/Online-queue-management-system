package dto

import (
	"Online-queue-management-system/services/branches/internal/domain"
	"time"
)

type BranchResponse struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type ClientResponse struct {
	ID         int64     `json:"id"`
	Email      *string   `json:"email,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
	Name       string    `json:"name"`
	Surname    string    `json:"surname"`
	TgUsername *string   `json:"tg_username,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type AppointmentResponse struct {
	ID              int64          `json:"id"`
	BranchID        int64          `json:"branch_id"`
	Client          ClientResponse `json:"client"`
	EmployeeID      int64          `json:"employee_id"`
	EmployeeName    string         `json:"employee_name"`
	EmployeeSurname string         `json:"employee_surname"`
	ServiceID       int64          `json:"service_id"`
	ServiceName     string         `json:"service_name"`
	StartTime       time.Time      `json:"start_time"`
	EndTime         time.Time      `json:"end_time"`
	Status          string         `json:"status"`
	Comment         *string        `json:"comment,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
}

func ClientsFromDomain(clients []domain.Client) []ClientResponse {
	result := make([]ClientResponse, 0, len(clients))
	for i := range clients {
		result = append(result, ClientFromDomain(&clients[i]))
	}
	return result
}

func ClientFromDomain(client *domain.Client) ClientResponse {
	return ClientResponse{
		ID:         client.ID,
		Email:      client.Email,
		Phone:      client.Phone,
		Name:       client.Name,
		Surname:    client.Surname,
		TgUsername: client.TgUsername,
		CreatedAt:  client.CreatedAt,
	}
}

func AppointmentsFromDomain(appointments []domain.Appointment) []AppointmentResponse {
	result := make([]AppointmentResponse, 0, len(appointments))
	for i := range appointments {
		result = append(result, AppointmentFromDomain(&appointments[i]))
	}
	return result
}

func AppointmentFromDomain(appointment *domain.Appointment) AppointmentResponse {
	return AppointmentResponse{
		ID:              appointment.ID,
		BranchID:        appointment.BranchID,
		Client:          ClientFromDomain(&appointment.Client),
		EmployeeID:      appointment.EmployeeID,
		EmployeeName:    appointment.EmployeeName,
		EmployeeSurname: appointment.EmployeeSurname,
		ServiceID:       appointment.ServiceID,
		ServiceName:     appointment.ServiceName,
		StartTime:       appointment.StartTime,
		EndTime:         appointment.EndTime,
		Status:          appointment.Status,
		Comment:         appointment.Comment,
		CreatedAt:       appointment.CreatedAt,
	}
}
