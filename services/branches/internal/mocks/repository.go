package mocks

import (
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
)

type MockBranchesRepository struct {
	GetByBusinessIDFunc        func(ctx context.Context, businessID int64) ([]domain.Branch, error)
	GetByIDFunc                func(ctx context.Context, branchID int64) ([]domain.Branch, error)
	GetEmployeesByBranchIDFunc func(ctx context.Context, branchID int64) ([]domain.Employee, error)
}

func (m *MockBranchesRepository) GetByBusinessID(ctx context.Context, businessID int64) ([]domain.Branch, error) {
	if m.GetByBusinessIDFunc != nil {
		return m.GetByBusinessIDFunc(ctx, businessID)
	}
	return nil, nil
}

func (m *MockBranchesRepository) GetByID(ctx context.Context, branchID int64) ([]domain.Branch, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, branchID)
	}
	return nil, nil
}

func (m *MockBranchesRepository) GetEmployeesByBranchID(ctx context.Context, branchID int64) ([]domain.Employee, error) {
	if m.GetEmployeesByBranchIDFunc != nil {
		return m.GetEmployeesByBranchIDFunc(ctx, branchID)
	}
	return nil, nil
}
