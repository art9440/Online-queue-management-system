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
}
