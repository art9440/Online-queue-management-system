package service

import (
	"Online-queue-management-system/libs/auth"
	"Online-queue-management-system/services/branches/internal/domain"
	"Online-queue-management-system/services/branches/internal/mocks"
	"context"
	"errors"
	"testing"
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
