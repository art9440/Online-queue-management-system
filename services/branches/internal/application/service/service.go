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

	switch auth.Role(user.RoleName) {

	case auth.RoleBusinessAdmin:
		return s.repoPostgres.GetByBusinessID(ctx, user.BusinessID)

	case auth.RoleManager:
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
	switch auth.Role(user.RoleName) {
	case auth.RoleBusinessAdmin:
		ok, err := s.repoPostgres.BranchBelongsToBusiness(ctx, branchID, user.BusinessID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrForbidden
		}
		return nil
	case auth.RoleManager:
		if user.BranchID == nil || *user.BranchID != branchID {
			return domain.ErrForbidden
		}
		return nil
	default:
		return domain.ErrForbidden
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

	switch auth.Role(user.RoleName) {
	case auth.RoleBusinessAdmin:
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

	case auth.RoleManager:
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

func (s *BranchesService) GetServicesForBusiness(
	ctx context.Context,
	user *auth.AccessClaims,
) ([]domain.Service, error) {
	if user == nil {
		return nil, domain.ErrUnauthorized
	}

	switch auth.Role(user.RoleName) {
	case auth.RoleBusinessAdmin:
		return s.repoPostgres.GetServicesByBusinessID(ctx, user.BusinessID)

	default:
		return nil, domain.ErrForbidden
	}
}

func (s *BranchesService) GetBranchesWithServiceForBusiness(
	ctx context.Context,
	user *auth.AccessClaims,
	serviceID int64,
) ([]domain.Branch, error) {
	if user == nil {
		return nil, domain.ErrUnauthorized
	}

	if serviceID <= 0 {
		return nil, domain.ErrInvalidServiceID
	}

	switch auth.Role(user.RoleName) {
	case auth.RoleBusinessAdmin:
		return s.repoPostgres.GetBranchesWithService(ctx, user.BusinessID, serviceID)

	default:
		return nil, domain.ErrForbidden
	}
}

func (s *BranchesService) GetEmployeesForServiceAndBranch(
	ctx context.Context,
	user *auth.AccessClaims,
	serviceID int64,
	branchID int64,
) ([]domain.Employee, error) {
	if user == nil {
		return nil, domain.ErrUnauthorized
	}

	if serviceID <= 0 {
		return nil, domain.ErrInvalidServiceID
	}

	if branchID <= 0 {
		return nil, domain.ErrInvalidBranchID
	}

	switch auth.Role(user.RoleName) {
	case auth.RoleBusinessAdmin:
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

		return s.repoPostgres.GetEmployeesByServiceAndBranch(ctx, serviceID, branchID)

	case auth.RoleManager:
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

		return s.repoPostgres.GetEmployeesByServiceAndBranch(ctx, serviceID, branchID)

	default:
		return nil, domain.ErrForbidden
	}
}

// Public methods for unauthenticated clients via registration slug

func (s *BranchesService) GetPublicServicesForSlug(
	ctx context.Context,
	registrationSlug string,
) ([]domain.Service, error) {
	if registrationSlug == "" {
		return nil, domain.ErrInvalidRegistrationSlug
	}

	businessID, err := s.repoPostgres.GetBusinessIDByRegistrationSlug(ctx, registrationSlug)
	if err != nil {
		return nil, err
	}

	return s.repoPostgres.GetServicesByBusinessID(ctx, businessID)
}

func (s *BranchesService) GetPublicBranchesWithServiceForSlug(
	ctx context.Context,
	registrationSlug string,
	serviceID int64,
) ([]domain.Branch, error) {
	if registrationSlug == "" {
		return nil, domain.ErrInvalidRegistrationSlug
	}

	if serviceID <= 0 {
		return nil, domain.ErrInvalidServiceID
	}

	businessID, err := s.repoPostgres.GetBusinessIDByRegistrationSlug(ctx, registrationSlug)
	if err != nil {
		return nil, err
	}

	return s.repoPostgres.GetBranchesWithService(ctx, businessID, serviceID)
}

func (s *BranchesService) GetPublicEmployeesForServiceAndBranchSlug(
	ctx context.Context,
	registrationSlug string,
	serviceID int64,
	branchID int64,
) ([]domain.Employee, error) {
	if registrationSlug == "" {
		return nil, domain.ErrInvalidRegistrationSlug
	}

	if serviceID <= 0 {
		return nil, domain.ErrInvalidServiceID
	}

	if branchID <= 0 {
		return nil, domain.ErrInvalidBranchID
	}

	businessID, err := s.repoPostgres.GetBusinessIDByRegistrationSlug(ctx, registrationSlug)
	if err != nil {
		return nil, err
	}

	// Verify that the branch belongs to this business
	branch, err := s.repoPostgres.GetByID(ctx, branchID)
	if err != nil {
		return nil, err
	}

	if len(branch) == 0 {
		return nil, domain.ErrBranchNotFound
	}

	if branch[0].BusinessID != businessID {
		return nil, domain.ErrForbidden
	}

	return s.repoPostgres.GetEmployeesByServiceAndBranch(ctx, serviceID, branchID)
}
