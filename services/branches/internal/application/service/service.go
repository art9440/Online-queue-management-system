package service

import (
	"Online-queue-management-system/libs/auth"
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
	"time"
)

type BranchesService struct {
	repoPostgres BranchesRepository
}

func New(repoPostgres BranchesRepository) *BranchesService {
	return &BranchesService{repoPostgres: repoPostgres}
}

func (s *BranchesService) GetBranchesForUser(ctx context.Context, user *auth.AccessClaims) ([]domain.Branch, error) {

	switch domain.Role(user.RoleName) {

	case domain.RoleBusinessAdmin:
		return s.repoPostgres.GetByBusinessID(ctx, user.BusinessID)

	case domain.RoleManager:
		if user.BranchID == nil {
			return nil, domain.ErrBranchNotFound
		}
		return s.repoPostgres.GetByID(ctx, *user.BranchID)

	default:
		return nil, domain.ErrForbidden
	}
}

func (s *BranchesService) GetBranchClients(ctx context.Context, user *auth.AccessClaims, branchID int64) ([]domain.Client, error) {
	if err := s.ensureBranchAccess(ctx, user, branchID); err != nil {
		return nil, err
	}

	return s.repoPostgres.GetClientsByBranchID(ctx, branchID)
}

func (s *BranchesService) GetBranchAppointments(
	ctx context.Context,
	user *auth.AccessClaims,
	branchID int64,
	date time.Time,
) ([]domain.Appointment, error) {
	if err := s.ensureBranchAccess(ctx, user, branchID); err != nil {
		return nil, err
	}

	return s.repoPostgres.GetAppointmentsByBranchIDAndDate(ctx, branchID, date)
}

func (s *BranchesService) ensureBranchAccess(ctx context.Context, user *auth.AccessClaims, branchID int64) error {
	switch domain.Role(user.RoleName) {
	case domain.RoleBusinessAdmin:
		ok, err := s.repoPostgres.BranchBelongsToBusiness(ctx, branchID, user.BusinessID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrForbidden
		}
		return nil
	case domain.RoleManager:
		if user.BranchID == nil || *user.BranchID != branchID {
			return domain.ErrForbidden
		}
		return nil
	default:
		return domain.ErrForbidden
	}
}
