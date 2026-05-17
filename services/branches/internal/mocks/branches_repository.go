package mocks

import (
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
)

type BranchesRepository struct {
	ByBusinessID map[int64][]domain.Branch
	ByID         map[int64][]domain.Branch

	EmployeesByBranchID map[int64][]domain.Employee

	ServicesByBusinessID map[int64][]domain.Service
	BranchesByService    map[int64][]domain.Branch
	EmployeesByService   map[int64][]domain.Employee
	BusinessBySlug       map[string]int64

	LastBusinessID int64
	LastBranchID   int64

	LastEmployeesBranchID int64

	BusinessIDCalls int
	IDCalls         int
	EmployeesCalls  int

	Err error
}

func NewBranchesRepository() *BranchesRepository {
	return &BranchesRepository{
		ByBusinessID:         make(map[int64][]domain.Branch),
		ByID:                 make(map[int64][]domain.Branch),
		EmployeesByBranchID:  make(map[int64][]domain.Employee),
		ServicesByBusinessID: make(map[int64][]domain.Service),
		BranchesByService:    make(map[int64][]domain.Branch),
		EmployeesByService:   make(map[int64][]domain.Employee),
		BusinessBySlug:       make(map[string]int64),
	}
}

func (r *BranchesRepository) GetByBusinessID(
	_ context.Context,
	businessID int64,
) ([]domain.Branch, error) {
	r.BusinessIDCalls++
	r.LastBusinessID = businessID

	if r.Err != nil {
		return nil, r.Err
	}

	return r.ByBusinessID[businessID], nil
}

func (r *BranchesRepository) GetByID(
	_ context.Context,
	branchID int64,
) ([]domain.Branch, error) {
	r.IDCalls++
	r.LastBranchID = branchID

	if r.Err != nil {
		return nil, r.Err
	}

	branches, ok := r.ByID[branchID]
	if !ok {
		return nil, domain.ErrBranchNotFound
	}

	return branches, nil
}

func (r *BranchesRepository) GetEmployeesByBranchID(
	_ context.Context,
	branchID int64,
) ([]domain.Employee, error) {
	r.EmployeesCalls++
	r.LastEmployeesBranchID = branchID

	if r.Err != nil {
		return nil, r.Err
	}

	employees, ok := r.EmployeesByBranchID[branchID]
	if !ok {
		return []domain.Employee{}, nil
	}

	return employees, nil
}

func (r *BranchesRepository) GetServicesByBusinessID(
	_ context.Context,
	businessID int64,
) ([]domain.Service, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	return r.ServicesByBusinessID[businessID], nil
}

func (r *BranchesRepository) GetBranchesWithService(
	_ context.Context,
	businessID int64,
	serviceID int64,
) ([]domain.Branch, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	return r.BranchesByService[serviceID], nil
}

func (r *BranchesRepository) GetEmployeesByServiceAndBranch(
	_ context.Context,
	serviceID int64,
	branchID int64,
) ([]domain.Employee, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	return r.EmployeesByService[serviceID], nil
}

func (r *BranchesRepository) GetBusinessIDByRegistrationSlug(
	_ context.Context,
	registrationSlug string,
) (int64, error) {
	if r.Err != nil {
		return 0, r.Err
	}

	if businessID, ok := r.BusinessBySlug[registrationSlug]; ok {
		return businessID, nil
	}

	return 0, domain.ErrBranchNotFound
}
