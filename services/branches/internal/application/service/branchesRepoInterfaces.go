package service

import (
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
	"time"
)

type BranchesRepository interface {
	GetByBusinessID(ctx context.Context, businessID int64) ([]domain.Branch, error)
	GetByID(ctx context.Context, branchID int64) ([]domain.Branch, error)
	BranchBelongsToBusiness(ctx context.Context, branchID, businessID int64) (bool, error)
	GetClientsByBranchID(ctx context.Context, branchID int64) ([]domain.Client, error)
	GetAppointmentsByBranchIDAndDate(ctx context.Context, branchID int64, date time.Time) ([]domain.Appointment, error)
	GetEmployeesByBranchID(ctx context.Context, branchID int64) ([]domain.Employee, error)
	GetServicesByBusinessID(ctx context.Context, businessID int64) ([]domain.Service, error)
	GetBranchesWithService(ctx context.Context, businessID int64, serviceID int64) ([]domain.Branch, error)
	GetEmployeesByServiceAndBranch(ctx context.Context, serviceID int64, branchID int64) ([]domain.Employee, error)
	GetBusinessIDByRegistrationSlug(ctx context.Context, registrationSlug string) (int64, error)
	GetRegistrationSlugByBusinessID(ctx context.Context, businessID int64) (string, error)
}
