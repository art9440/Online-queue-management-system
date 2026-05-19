package service

import (
	"Online-queue-management-system/libs/auth"
	"Online-queue-management-system/services/branches/internal/domain"
	"Online-queue-management-system/services/branches/internal/mocks"
	"context"
	"errors"
	"testing"
	"time"
)

func TestGetBranchesForUser_BusinessAdmin_ReturnsBusinessBranches(t *testing.T) {
	// Arrange
	expectedBranches := []domain.Branch{
		{ID: 1, BusinessID: 100, Name: "Branch 1", Address: "Address 1"},
		{ID: 2, BusinessID: 100, Name: "Branch 2", Address: "Address 2"},
	}

	mockRepo := &mocks.MockBranchesRepository{
		GetByBusinessIDFunc: func(ctx context.Context, businessID int64) ([]domain.Branch, error) {
			if businessID == 100 {
				return expectedBranches, nil
			}
			return nil, nil
		},
	}

	svc := New(mockRepo)
	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	// Act
	branches, err := svc.GetBranchesForUser(context.Background(), user)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != len(expectedBranches) {
		t.Fatalf("expected %d branches, got %d", len(expectedBranches), len(branches))
	}
	if branches[0].ID != expectedBranches[0].ID {
		t.Fatalf("expected branch ID %d, got %d", expectedBranches[0].ID, branches[0].ID)
	}
}

func TestGetBranchesForUser_Manager_ReturnsManagerBranch(t *testing.T) {
	// Arrange
	expectedBranches := []domain.Branch{
		{ID: 1, BusinessID: 100, Name: "Branch 1", Address: "Address 1"},
	}

	mockRepo := &mocks.MockBranchesRepository{
		GetByIDFunc: func(ctx context.Context, branchID int64) ([]domain.Branch, error) {
			if branchID == 1 {
				return expectedBranches, nil
			}
			return nil, domain.ErrBranchNotFound
		},
	}

	svc := New(mockRepo)
	branchID := int64(1)
	user := &auth.AccessClaims{
		UserID:     2,
		RoleID:     3,
		RoleName:   "manager",
		BusinessID: 100,
		BranchID:   &branchID,
	}

	// Act
	branches, err := svc.GetBranchesForUser(context.Background(), user)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 1 {
		t.Fatalf("expected 1 branch, got %d", len(branches))
	}
	if branches[0].ID != 1 {
		t.Fatalf("expected branch ID 1, got %d", branches[0].ID)
	}
}

func TestGetBranchesForUser_Manager_NoBranchID_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockBranchesRepository{}

	svc := New(mockRepo)
	user := &auth.AccessClaims{
		UserID:     2,
		RoleID:     3,
		RoleName:   "manager",
		BusinessID: 100,
		BranchID:   nil,
	}

	// Act
	branches, err := svc.GetBranchesForUser(context.Background(), user)

	// Assert
	if err != domain.ErrBranchNotFound {
		t.Fatalf("expected ErrBranchNotFound, got %v", err)
	}
	if branches != nil {
		t.Fatalf("expected nil branches, got %v", branches)
	}
}

func TestGetBranchesForUser_InvalidRole_ReturnsForbidden(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockBranchesRepository{}

	svc := New(mockRepo)
	user := &auth.AccessClaims{
		UserID:     3,
		RoleID:     4,
		RoleName:   "employee",
		BusinessID: 100,
	}

	// Act
	branches, err := svc.GetBranchesForUser(context.Background(), user)

	// Assert
	if err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if branches != nil {
		t.Fatalf("expected nil branches, got %v", branches)
	}
}

func TestGetEmployeesForBranch_BusinessAdmin_Success(t *testing.T) {
	// Arrange
	branch := []domain.Branch{
		{ID: 1, BusinessID: 100, Name: "Branch 1", Address: "Address 1"},
	}
	employees := []domain.Employee{
		{ID: 1, BranchID: 1, Name: "John", Surname: "Doe", Position: "Specialist"},
		{ID: 2, BranchID: 1, Name: "Jane", Surname: "Smith", Position: "Senior"},
	}

	mockRepo := &mocks.MockBranchesRepository{
		GetByIDFunc: func(ctx context.Context, branchID int64) ([]domain.Branch, error) {
			if branchID == 1 {
				return branch, nil
			}
			return nil, domain.ErrBranchNotFound
		},
		GetEmployeesByBranchIDFunc: func(ctx context.Context, branchID int64) ([]domain.Employee, error) {
			if branchID == 1 {
				return employees, nil
			}
			return nil, nil
		},
	}

	svc := New(mockRepo)
	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	// Act
	result, err := svc.GetEmployeesForBranch(context.Background(), user, 1)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != len(employees) {
		t.Fatalf("expected %d employees, got %d", len(employees), len(result))
	}
	if result[0].Name != "John" {
		t.Fatalf("expected employee name John, got %s", result[0].Name)
	}
}

func TestGetEmployeesForBranch_BusinessAdmin_WrongBusiness_ReturnsForbidden(t *testing.T) {
	// Arrange
	branch := []domain.Branch{
		{ID: 1, BusinessID: 200, Name: "Branch 1", Address: "Address 1"},
	}

	mockRepo := &mocks.MockBranchesRepository{
		GetByIDFunc: func(ctx context.Context, branchID int64) ([]domain.Branch, error) {
			if branchID == 1 {
				return branch, nil
			}
			return nil, domain.ErrBranchNotFound
		},
	}

	svc := New(mockRepo)
	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	// Act
	result, err := svc.GetEmployeesForBranch(context.Background(), user, 1)

	// Assert
	if err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil employees, got %v", result)
	}
}

func TestGetEmployeesForBranch_Manager_Success(t *testing.T) {
	// Arrange
	branch := []domain.Branch{
		{ID: 1, BusinessID: 100, Name: "Branch 1", Address: "Address 1"},
	}
	employees := []domain.Employee{
		{ID: 1, BranchID: 1, Name: "John", Surname: "Doe", Position: "Specialist"},
	}

	mockRepo := &mocks.MockBranchesRepository{
		GetByIDFunc: func(ctx context.Context, branchID int64) ([]domain.Branch, error) {
			if branchID == 1 {
				return branch, nil
			}
			return nil, domain.ErrBranchNotFound
		},
		GetEmployeesByBranchIDFunc: func(ctx context.Context, branchID int64) ([]domain.Employee, error) {
			if branchID == 1 {
				return employees, nil
			}
			return nil, nil
		},
	}

	svc := New(mockRepo)
	branchID := int64(1)
	user := &auth.AccessClaims{
		UserID:     2,
		RoleID:     3,
		RoleName:   "manager",
		BusinessID: 100,
		BranchID:   &branchID,
	}

	// Act
	result, err := svc.GetEmployeesForBranch(context.Background(), user, 1)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 employee, got %d", len(result))
	}
}

func TestGetEmployeesForBranch_Manager_DifferentBranch_ReturnsForbidden(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockBranchesRepository{}

	svc := New(mockRepo)
	branchID := int64(1)
	user := &auth.AccessClaims{
		UserID:     2,
		RoleID:     3,
		RoleName:   "manager",
		BusinessID: 100,
		BranchID:   &branchID,
	}

	// Act
	result, err := svc.GetEmployeesForBranch(context.Background(), user, 2)

	// Assert
	if err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil employees, got %v", result)
	}
}

func TestGetEmployeesForBranch_InvalidBranchID_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockBranchesRepository{}

	svc := New(mockRepo)
	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	// Act
	result, err := svc.GetEmployeesForBranch(context.Background(), user, -1)

	// Assert
	if err != domain.ErrInvalidBranchID {
		t.Fatalf("expected ErrInvalidBranchID, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil employees, got %v", result)
	}
}

func TestGetEmployeesForBranch_NilUser_ReturnsUnauthorized(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockBranchesRepository{}

	svc := New(mockRepo)

	// Act
	result, err := svc.GetEmployeesForBranch(context.Background(), nil, 1)

	// Assert
	if err != domain.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil employees, got %v", result)
	}
}

func TestGetEmployeesForBranch_BranchNotFound_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockBranchesRepository{
		GetByIDFunc: func(ctx context.Context, branchID int64) ([]domain.Branch, error) {
			return nil, domain.ErrBranchNotFound
		},
	}

	svc := New(mockRepo)
	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	// Act
	result, err := svc.GetEmployeesForBranch(context.Background(), user, 1)

	// Assert
	if err != domain.ErrBranchNotFound {
		t.Fatalf("expected ErrBranchNotFound, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil employees, got %v", result)
	}
}

func TestGetEmployeesForBranch_RepositoryError_ReturnsError(t *testing.T) {
	// Arrange
	repoErr := errors.New("database error")
	mockRepo := &mocks.MockBranchesRepository{
		GetEmployeesByBranchIDFunc: func(ctx context.Context, branchID int64) ([]domain.Employee, error) {
			return nil, repoErr
		},
		GetByIDFunc: func(ctx context.Context, branchID int64) ([]domain.Branch, error) {
			return []domain.Branch{
				{ID: 1, BusinessID: 100, Name: "Branch 1", Address: "Address 1"},
			}, nil
		},
	}

	svc := New(mockRepo)
	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	// Act
	result, err := svc.GetEmployeesForBranch(context.Background(), user, 1)

	// Assert
	if err != repoErr {
		t.Fatalf("expected error %v, got %v", repoErr, err)
	}
	if result != nil {
		t.Fatalf("expected nil employees, got %v", result)
	}
}

// Tests for new service-related methods

func TestGetServicesForBusiness_ReturnsServices(t *testing.T) {
	// Arrange
	expectedServices := []domain.Service{
		{ID: 1, BranchID: 1, Name: "Service 1", DurationMinutes: 30, Price: 500.0},
		{ID: 2, BranchID: 1, Name: "Service 2", DurationMinutes: 45, Price: 750.0},
	}

	mockRepo := &mocks.MockBranchesRepository{
		GetServicesByBusinessIDFunc: func(ctx context.Context, businessID int64) ([]domain.Service, error) {
			if businessID == 100 {
				return expectedServices, nil
			}
			return nil, nil
		},
	}

	svc := New(mockRepo)
	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	// Act
	services, err := svc.GetServicesForBusiness(context.Background(), user)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(services) != len(expectedServices) {
		t.Fatalf("expected %d services, got %d", len(expectedServices), len(services))
	}
	if services[0].ID != expectedServices[0].ID {
		t.Fatalf("expected service ID %d, got %d", expectedServices[0].ID, services[0].ID)
	}
}

func TestGetBranchesWithServiceForBusiness_ReturnsServiceBranches(t *testing.T) {
	// Arrange
	expectedBranches := []domain.Branch{
		{ID: 1, BusinessID: 100, Name: "Branch 1", Address: "Address 1"},
		{ID: 2, BusinessID: 100, Name: "Branch 2", Address: "Address 2"},
	}

	mockRepo := &mocks.MockBranchesRepository{
		GetBranchesWithServiceFunc: func(ctx context.Context, businessID int64, serviceID int64) ([]domain.Branch, error) {
			if businessID == 100 && serviceID == 1 {
				return expectedBranches, nil
			}
			return nil, nil
		},
	}

	svc := New(mockRepo)
	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	// Act
	branches, err := svc.GetBranchesWithServiceForBusiness(context.Background(), user, 1)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != len(expectedBranches) {
		t.Fatalf("expected %d branches, got %d", len(expectedBranches), len(branches))
	}
}

func TestGetEmployeesForServiceAndBranch_ReturnsEmployees(t *testing.T) {
	// Arrange
	expectedEmployees := []domain.Employee{
		{ID: 1, BranchID: 1, Name: "John", Surname: "Doe", Position: "Master"},
		{ID: 2, BranchID: 1, Name: "Jane", Surname: "Smith", Position: "Master"},
	}

	mockRepo := &mocks.MockBranchesRepository{
		GetByIDFunc: func(ctx context.Context, branchID int64) ([]domain.Branch, error) {
			return []domain.Branch{
				{ID: 1, BusinessID: 100, Name: "Branch 1", Address: "Address 1"},
			}, nil
		},
		GetEmployeesByServiceAndBranchFunc: func(ctx context.Context, serviceID int64, branchID int64) ([]domain.Employee, error) {
			if serviceID == 1 && branchID == 1 {
				return expectedEmployees, nil
			}
			return nil, nil
		},
	}

	svc := New(mockRepo)
	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	// Act
	employees, err := svc.GetEmployeesForServiceAndBranch(context.Background(), user, 1, 1)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(employees) != len(expectedEmployees) {
		t.Fatalf("expected %d employees, got %d", len(expectedEmployees), len(employees))
	}
}

// Tests for public methods

func TestGetPublicServicesForSlug_ReturnsServices(t *testing.T) {
	// Arrange
	expectedServices := []domain.Service{
		{ID: 1, BranchID: 1, Name: "Service 1", DurationMinutes: 30, Price: 500.0},
	}

	mockRepo := &mocks.MockBranchesRepository{
		GetBusinessIDByRegistrationSlugFunc: func(ctx context.Context, registrationSlug string) (int64, error) {
			if registrationSlug == "beautiful-salon" {
				return 100, nil
			}
			return 0, errors.New("business not found")
		},
		GetServicesByBusinessIDFunc: func(ctx context.Context, businessID int64) ([]domain.Service, error) {
			if businessID == 100 {
				return expectedServices, nil
			}
			return nil, nil
		},
	}

	svc := New(mockRepo)

	// Act
	services, err := svc.GetPublicServicesForSlug(context.Background(), "beautiful-salon")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
}

func TestGetPublicBranchesWithServiceForSlug_ReturnsBranches(t *testing.T) {
	// Arrange
	expectedBranches := []domain.Branch{
		{ID: 1, BusinessID: 100, Name: "Branch 1", Address: "Address 1"},
	}

	mockRepo := &mocks.MockBranchesRepository{
		GetBusinessIDByRegistrationSlugFunc: func(ctx context.Context, registrationSlug string) (int64, error) {
			if registrationSlug == "beautiful-salon" {
				return 100, nil
			}
			return 0, errors.New("business not found")
		},
		GetBranchesWithServiceFunc: func(ctx context.Context, businessID int64, serviceID int64) ([]domain.Branch, error) {
			if businessID == 100 && serviceID == 1 {
				return expectedBranches, nil
			}
			return nil, nil
		},
	}

	svc := New(mockRepo)

	// Act
	branches, err := svc.GetPublicBranchesWithServiceForSlug(context.Background(), "beautiful-salon", 1)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 1 {
		t.Fatalf("expected 1 branch, got %d", len(branches))
	}
}

func TestGetPublicEmployeesForServiceAndBranchSlug_ReturnsEmployees(t *testing.T) {
	// Arrange
	expectedEmployees := []domain.Employee{
		{ID: 1, BranchID: 1, Name: "John", Surname: "Doe", Position: "Master"},
	}

	mockRepo := &mocks.MockBranchesRepository{
		GetBusinessIDByRegistrationSlugFunc: func(ctx context.Context, registrationSlug string) (int64, error) {
			if registrationSlug == "beautiful-salon" {
				return 100, nil
			}
			return 0, errors.New("business not found")
		},
		GetByIDFunc: func(ctx context.Context, branchID int64) ([]domain.Branch, error) {
			return []domain.Branch{
				{ID: 1, BusinessID: 100, Name: "Branch 1", Address: "Address 1"},
			}, nil
		},
		GetEmployeesByServiceAndBranchFunc: func(ctx context.Context, serviceID int64, branchID int64) ([]domain.Employee, error) {
			if serviceID == 1 && branchID == 1 {
				return expectedEmployees, nil
			}
			return nil, nil
		},
	}

	svc := New(mockRepo)

	// Act
	employees, err := svc.GetPublicEmployeesForServiceAndBranchSlug(context.Background(), "beautiful-salon", 1, 1)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(employees) != 1 {
		t.Fatalf("expected 1 employee, got %d", len(employees))
	}
}

func TestGetBranchClients_WhenBusinessAdminOwnsBranch_ShouldReturnClients(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewBranchesRepository()
	svc := New(repo)
	repo.BranchBusiness[11] = 7
	repo.ClientsByBranchID[11] = []domain.Client{{ID: 5, Name: "Alex", Surname: "Stone"}}

	clients, err := svc.GetBranchClients(ctx, &auth.AccessClaims{
		RoleName:   string(auth.RoleBusinessAdmin),
		BusinessID: 7,
	}, 11)
	if err != nil {
		t.Fatalf("get branch clients: %v", err)
	}

	if repo.BranchBelongsCalls != 1 || repo.ClientsByBranchCalls != 1 {
		t.Fatalf("unexpected calls: belongs=%d clients=%d", repo.BranchBelongsCalls, repo.ClientsByBranchCalls)
	}
	if len(clients) != 1 || clients[0].ID != 5 {
		t.Fatalf("unexpected clients: %#v", clients)
	}
}

func TestGetBranchAppointments_WhenManagerRequestsAnotherBranch_ShouldReturnForbidden(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewBranchesRepository()
	svc := New(repo)
	managerBranchID := int64(11)

	_, err := svc.GetBranchAppointments(ctx, &auth.AccessClaims{
		RoleName: string(auth.RoleManager),
		BranchID: &managerBranchID,
	}, 12, time.Now())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if repo.AppointmentsByBranchCalls != 0 {
		t.Fatalf("expected appointments repository not to be called, got %d", repo.AppointmentsByBranchCalls)
	}
}

func TestGetRegistrationSlugForBusiness_WhenBusinessAdminOwnsBusiness_ShouldReturnSlug(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewBranchesRepository()
	svc := New(repo)
	repo.SlugByBusinessID[7] = "beautiful-salon"

	registrationSlug, err := svc.GetRegistrationSlugForBusiness(ctx, &auth.AccessClaims{
		RoleName:   string(auth.RoleBusinessAdmin),
		BusinessID: 7,
	}, 7)
	if err != nil {
		t.Fatalf("get registration slug: %v", err)
	}

	if registrationSlug != "beautiful-salon" {
		t.Fatalf("expected registration slug beautiful-salon, got %q", registrationSlug)
	}
	if repo.BusinessIDCalls != 1 || repo.LastBusinessID != 7 {
		t.Fatalf("unexpected repo call state: calls=%d business_id=%d", repo.BusinessIDCalls, repo.LastBusinessID)
	}
}

func TestGetRegistrationSlugForBusiness_WhenUserIsNotBusinessAdmin_ShouldReturnForbidden(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewBranchesRepository()
	svc := New(repo)

	_, err := svc.GetRegistrationSlugForBusiness(ctx, &auth.AccessClaims{
		RoleName:   string(auth.RoleManager),
		BusinessID: 7,
	}, 7)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if repo.BusinessIDCalls != 0 {
		t.Fatalf("expected repository not to be called, got %d", repo.BusinessIDCalls)
	}
}

func TestGetRegistrationSlugForBusiness_WhenBusinessAdminRequestsAnotherBusiness_ShouldReturnForbidden(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewBranchesRepository()
	svc := New(repo)

	_, err := svc.GetRegistrationSlugForBusiness(ctx, &auth.AccessClaims{
		RoleName:   string(auth.RoleBusinessAdmin),
		BusinessID: 7,
	}, 8)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if repo.BusinessIDCalls != 0 {
		t.Fatalf("expected repository not to be called, got %d", repo.BusinessIDCalls)
	}
}
