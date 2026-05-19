package mocks

import (
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
	"time"
)

type BranchesRepository struct {
	ByBusinessID              map[int64][]domain.Branch
	ByID                      map[int64][]domain.Branch
	BranchBusiness            map[int64]int64
	ClientsByBranchID         map[int64][]domain.Client
	AppointmentsByBranchDate  map[branchDateKey][]domain.Appointment
	LastBusinessID            int64
	LastBranchID              int64
	LastAppointmentDate       time.Time
	BusinessIDCalls           int
	IDCalls                   int
	BranchBelongsCalls        int
	ClientsByBranchCalls      int
	AppointmentsByBranchCalls int
	EmployeesCalls            int
	LastEmployeesBranchID     int64
	Err                       error

	EmployeesByBranchID map[int64][]domain.Employee

	ServicesByBusinessID map[int64][]domain.Service
	BranchesByService    map[int64][]domain.Branch
	EmployeesByService   map[int64][]domain.Employee
	BusinessBySlug       map[string]int64
}

func NewBranchesRepository() *BranchesRepository {
	return &BranchesRepository{
		ByBusinessID:             make(map[int64][]domain.Branch),
		ByID:                     make(map[int64][]domain.Branch),
		BranchBusiness:           make(map[int64]int64),
		ClientsByBranchID:        make(map[int64][]domain.Client),
		AppointmentsByBranchDate: make(map[branchDateKey][]domain.Appointment),
		EmployeesByBranchID:      make(map[int64][]domain.Employee),
		ServicesByBusinessID:     make(map[int64][]domain.Service),
		BranchesByService:        make(map[int64][]domain.Branch),
		EmployeesByService:       make(map[int64][]domain.Employee),
		BusinessBySlug:           make(map[string]int64),
		EmployeesCalls:           0,
		LastEmployeesBranchID:    0,
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

func (r *BranchesRepository) BranchBelongsToBusiness(_ context.Context, branchID, businessID int64) (bool, error) {
	r.BranchBelongsCalls++
	r.LastBranchID = branchID
	r.LastBusinessID = businessID
	if r.Err != nil {
		return false, r.Err
	}
	return r.BranchBusiness[branchID] == businessID, nil
}

func (r *BranchesRepository) GetClientsByBranchID(_ context.Context, branchID int64) ([]domain.Client, error) {
	r.ClientsByBranchCalls++
	r.LastBranchID = branchID
	if r.Err != nil {
		return nil, r.Err
	}
	return r.ClientsByBranchID[branchID], nil
}

func (r *BranchesRepository) GetAppointmentsByBranchIDAndDate(
	_ context.Context,
	branchID int64,
	date time.Time,
) ([]domain.Appointment, error) {
	r.AppointmentsByBranchCalls++
	r.LastBranchID = branchID
	r.LastAppointmentDate = date
	if r.Err != nil {
		return nil, r.Err
	}
	return r.AppointmentsByBranchDate[branchDateKey{branchID: branchID, date: date.Format(time.DateOnly)}], nil
}

func (r *BranchesRepository) SetAppointments(branchID int64, date time.Time, appointments []domain.Appointment) {
	r.AppointmentsByBranchDate[branchDateKey{branchID: branchID, date: date.Format(time.DateOnly)}] = appointments
}

type branchDateKey struct {
	branchID int64
	date     string
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
