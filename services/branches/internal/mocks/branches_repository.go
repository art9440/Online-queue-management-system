package mocks

import (
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
)

type BranchesRepository struct {
	ByBusinessID    map[int64][]domain.Branch
	ByID            map[int64][]domain.Branch
	LastBusinessID  int64
	LastBranchID    int64
	BusinessIDCalls int
	IDCalls         int
	Err             error
}

func NewBranchesRepository() *BranchesRepository {
	return &BranchesRepository{
		ByBusinessID: make(map[int64][]domain.Branch),
		ByID:         make(map[int64][]domain.Branch),
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
