package mocks

import (
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
)

type MockBranchesRepository struct {
	GetByBusinessIDFunc            func(ctx context.Context, businessID int64) ([]domain.Branch, error)
	GetByIDFunc                    func(ctx context.Context, branchID int64) ([]domain.Branch, error)
	GetEmployeesByBranchIDFunc     func(ctx context.Context, branchID int64) ([]domain.Employee, error)
	GetServicesByBusinessIDFunc    func(ctx context.Context, businessID int64) ([]domain.Service, error)
	GetBranchesWithServiceFunc     func(ctx context.Context, businessID int64, serviceID int64) ([]domain.Branch, error)
	GetEmployeesByServiceAndBranchFunc func(ctx context.Context, serviceID int64, branchID int64) ([]domain.Employee, error)
	GetBusinessIDByRegistrationSlugFunc func(ctx context.Context, registrationSlug string) (int64, error)
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

func (m *MockBranchesRepository) GetServicesByBusinessID(ctx context.Context, businessID int64) ([]domain.Service, error) {
	if m.GetServicesByBusinessIDFunc != nil {
		return m.GetServicesByBusinessIDFunc(ctx, businessID)
	}
	return nil, nil
}

func (m *MockBranchesRepository) GetBranchesWithService(ctx context.Context, businessID int64, serviceID int64) ([]domain.Branch, error) {
	if m.GetBranchesWithServiceFunc != nil {
		return m.GetBranchesWithServiceFunc(ctx, businessID, serviceID)
	}
	return nil, nil
}

func (m *MockBranchesRepository) GetEmployeesByServiceAndBranch(ctx context.Context, serviceID int64, branchID int64) ([]domain.Employee, error) {
	if m.GetEmployeesByServiceAndBranchFunc != nil {
		return m.GetEmployeesByServiceAndBranchFunc(ctx, serviceID, branchID)
	}
	return nil, nil
}

func (m *MockBranchesRepository) GetBusinessIDByRegistrationSlug(ctx context.Context, registrationSlug string) (int64, error) {
	if m.GetBusinessIDByRegistrationSlugFunc != nil {
		return m.GetBusinessIDByRegistrationSlugFunc(ctx, registrationSlug)
	}
	return 0, nil
}
