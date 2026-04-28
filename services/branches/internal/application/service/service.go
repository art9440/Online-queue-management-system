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
