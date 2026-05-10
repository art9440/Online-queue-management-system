package service

import (
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
)

type BranchesRepository interface {
	GetByBusinessID(ctx context.Context, businessID int64) ([]domain.Branch, error)
	GetByID(ctx context.Context, branchID int64) ([]domain.Branch, error)
	GetEmployeesByBranchID(ctx context.Context, branchID int64) ([]domain.Employee, error)
}
