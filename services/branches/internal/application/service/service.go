package service

import (
	"Online-queue-management-system/libs/auth"
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
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

func (s *BranchesService) GetEmployeesForBranch(
	ctx context.Context,
	user *auth.AccessClaims,
	branchID int64,
) ([]domain.Employee, error) {
	if user == nil {
		return nil, domain.ErrUnauthorized
	}

	if branchID <= 0 {
		return nil, domain.ErrInvalidBranchID
	}

	switch domain.Role(user.RoleName) {
	case domain.RoleBusinessAdmin:
		branch, err := s.repoPostgres.GetByID(ctx, branchID)
		if err != nil {
			return nil, err
		}

		if len(branch) == 0 {
			return nil, domain.ErrBranchNotFound
		}

		if branch[0].BusinessID != user.BusinessID {
			return nil, domain.ErrForbidden
		}

		return s.repoPostgres.GetEmployeesByBranchID(ctx, branchID)

	case domain.RoleManager:
		if user.BranchID == nil {
			return nil, domain.ErrForbidden
		}

		if *user.BranchID != branchID {
			return nil, domain.ErrForbidden
		}

		branch, err := s.repoPostgres.GetByID(ctx, branchID)
		if err != nil {
			return nil, err
		}

		if len(branch) == 0 {
			return nil, domain.ErrBranchNotFound
		}

		if branch[0].BusinessID != user.BusinessID {
			return nil, domain.ErrForbidden
		}

		return s.repoPostgres.GetEmployeesByBranchID(ctx, branchID)

	default:
		return nil, domain.ErrForbidden
	}
}
