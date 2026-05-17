package dto

import (
	"Online-queue-management-system/services/branches/internal/domain"
)

type BranchResponse struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
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
