package mocks

import (
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
	"time"
)

type BranchesRepository struct {
	ByBusinessID          map[int64][]domain.Branch
	ByID                  map[int64][]domain.Branch
	BranchBusiness        map[int64]int64
	ClientsByBranchID     map[int64][]domain.Client
	BookingsByBranchDate  map[branchDateKey][]domain.Booking
	LastBusinessID        int64
	LastBranchID          int64
	LastBookingDate       time.Time
	BusinessIDCalls       int
	IDCalls               int
	BranchBelongsCalls    int
	ClientsByBranchCalls  int
	BookingsByBranchCalls int
	Err                   error
}

func NewBranchesRepository() *BranchesRepository {
	return &BranchesRepository{
		ByBusinessID:         make(map[int64][]domain.Branch),
		ByID:                 make(map[int64][]domain.Branch),
		BranchBusiness:       make(map[int64]int64),
		ClientsByBranchID:    make(map[int64][]domain.Client),
		BookingsByBranchDate: make(map[branchDateKey][]domain.Booking),
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

func (r *BranchesRepository) GetBookingsByBranchIDAndDate(
	_ context.Context,
	branchID int64,
	date time.Time,
) ([]domain.Booking, error) {
	r.BookingsByBranchCalls++
	r.LastBranchID = branchID
	r.LastBookingDate = date
	if r.Err != nil {
		return nil, r.Err
	}
	return r.BookingsByBranchDate[branchDateKey{branchID: branchID, date: date.Format(time.DateOnly)}], nil
}

func (r *BranchesRepository) SetBookings(branchID int64, date time.Time, bookings []domain.Booking) {
	r.BookingsByBranchDate[branchDateKey{branchID: branchID, date: date.Format(time.DateOnly)}] = bookings
}

type branchDateKey struct {
	branchID int64
	date     string
}
