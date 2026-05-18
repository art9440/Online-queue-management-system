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
	Err                       error
}

func NewBranchesRepository() *BranchesRepository {
	return &BranchesRepository{
		ByBusinessID:             make(map[int64][]domain.Branch),
		ByID:                     make(map[int64][]domain.Branch),
		BranchBusiness:           make(map[int64]int64),
		ClientsByBranchID:        make(map[int64][]domain.Client),
		AppointmentsByBranchDate: make(map[branchDateKey][]domain.Appointment),
	}
}

func (r *BranchesRepository) GetByBusinessID(_ context.Context, businessID int64) ([]domain.Branch, error) {
	r.BusinessIDCalls++
	r.LastBusinessID = businessID
	if r.Err != nil {
		return nil, r.Err
	}
	return r.ByBusinessID[businessID], nil
}

func (r *BranchesRepository) GetByID(_ context.Context, branchID int64) ([]domain.Branch, error) {
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
