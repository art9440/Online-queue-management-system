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

type EmployeeResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Position string `json:"position"`
}

type ServiceResponse struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	DurationMinutes int     `json:"duration_minutes"`
	Price           float64 `json:"price"`
}

type BusinessRegistrationSlugResponse struct {
	BusinessID       int64  `json:"business_id"`
	RegistrationSlug string `json:"registration_slug"`
}

func BranchesFromDomain(branches []domain.Branch) []BranchResponse {
	result := make([]BranchResponse, 0, len(branches))

	for _, branch := range branches {
		result = append(result, branchFromDomain(branch))
	}

	return result
}

func EmployeesFromDomain(employees []domain.Employee) []EmployeeResponse {
	result := make([]EmployeeResponse, 0, len(employees))

	for _, employee := range employees {
		result = append(result, employeeFromDomain(employee))
	}

	return result
}

func employeeFromDomain(employee domain.Employee) EmployeeResponse {
	return EmployeeResponse{
		ID:       employee.ID,
		Name:     employee.Name,
		Surname:  employee.Surname,
		Position: employee.Position,
	}
}

func branchFromDomain(branch domain.Branch) BranchResponse {
	return BranchResponse{
		ID:      branch.ID,
		Name:    branch.Name,
		Address: branch.Address,
	}
}

func ServicesFromDomain(services []domain.Service) []ServiceResponse {
	result := make([]ServiceResponse, 0, len(services))

	for _, service := range services {
		result = append(result, serviceFromDomain(service))
	}

	return result
}

func serviceFromDomain(service domain.Service) ServiceResponse {
	return ServiceResponse{
		ID:              service.ID,
		Name:            service.Name,
		DurationMinutes: service.DurationMinutes,
		Price:           service.Price,
	}
}
