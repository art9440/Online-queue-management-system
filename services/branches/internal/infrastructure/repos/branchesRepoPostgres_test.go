package repos

import (
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetByBusinessID_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	expectedBranches := []domain.Branch{
		{ID: 1, BusinessID: 100, Name: "Branch 1", Address: "Address 1"},
		{ID: 2, BusinessID: 100, Name: "Branch 2", Address: "Address 2"},
	}

	rows := sqlmock.NewRows([]string{"id", "business_id", "name", "address"}).
		AddRow(expectedBranches[0].ID, expectedBranches[0].BusinessID, expectedBranches[0].Name, expectedBranches[0].Address).
		AddRow(expectedBranches[1].ID, expectedBranches[1].BusinessID, expectedBranches[1].Name, expectedBranches[1].Address)

	mock.ExpectQuery("SELECT id, business_id, name, address FROM branches WHERE business_id = \\$1").
		WithArgs(int64(100)).
		WillReturnRows(rows)

	repo := &BranchesRepoPostgres{db: db}

	// Act
	branches, err := repo.GetByBusinessID(context.Background(), 100)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}
	if branches[0].Name != "Branch 1" {
		t.Fatalf("expected Branch 1, got %s", branches[0].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestGetByBusinessID_Empty(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	rows := sqlmock.NewRows([]string{"id", "business_id", "name", "address"})

	mock.ExpectQuery("SELECT id, business_id, name, address FROM branches WHERE business_id = \\$1").
		WithArgs(int64(100)).
		WillReturnRows(rows)

	repo := &BranchesRepoPostgres{db: db}

	// Act
	branches, err := repo.GetByBusinessID(context.Background(), 100)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 0 {
		t.Fatalf("expected 0 branches, got %d", len(branches))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestGetByBusinessID_QueryError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	expectedErr := errors.New("database error")

	mock.ExpectQuery("SELECT id, business_id, name, address FROM branches WHERE business_id = \\$1").
		WithArgs(int64(100)).
		WillReturnError(expectedErr)

	repo := &BranchesRepoPostgres{db: db}

	// Act
	branches, err := repo.GetByBusinessID(context.Background(), 100)

	// Assert
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if branches != nil {
		t.Fatalf("expected nil branches, got %v", branches)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestGetByID_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	expectedBranch := domain.Branch{ID: 1, BusinessID: 100, Name: "Branch 1", Address: "Address 1"}

	rows := sqlmock.NewRows([]string{"id", "business_id", "name", "address"}).
		AddRow(expectedBranch.ID, expectedBranch.BusinessID, expectedBranch.Name, expectedBranch.Address)

	mock.ExpectQuery("SELECT id, business_id, name, address.*FROM branches.*WHERE id = \\$1").
		WithArgs(int64(1)).
		WillReturnRows(rows)

	repo := &BranchesRepoPostgres{db: db}

	// Act
	branches, err := repo.GetByID(context.Background(), 1)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 1 {
		t.Fatalf("expected 1 branch, got %d", len(branches))
	}
	if branches[0].Name != "Branch 1" {
		t.Fatalf("expected Branch 1, got %s", branches[0].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	mock.ExpectQuery("SELECT id, business_id, name, address.*FROM branches.*WHERE id = \\$1").
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	repo := &BranchesRepoPostgres{db: db}

	// Act
	branches, err := repo.GetByID(context.Background(), 999)

	// Assert
	if err != domain.ErrBranchNotFound {
		t.Fatalf("expected ErrBranchNotFound, got %v", err)
	}
	if branches != nil {
		t.Fatalf("expected nil branches, got %v", branches)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestGetByID_QueryError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	expectedErr := errors.New("database error")

	mock.ExpectQuery("SELECT id, business_id, name, address.*FROM branches.*WHERE id = \\$1").
		WithArgs(int64(1)).
		WillReturnError(expectedErr)

	repo := &BranchesRepoPostgres{db: db}

	// Act
	branches, err := repo.GetByID(context.Background(), 1)

	// Assert
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if branches != nil {
		t.Fatalf("expected nil branches, got %v", branches)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestGetEmployeesByBranchID_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	expectedEmployees := []domain.Employee{
		{ID: 1, BranchID: 1, Name: "John", Surname: "Doe", Position: "Specialist"},
		{ID: 2, BranchID: 1, Name: "Jane", Surname: "Smith", Position: "Senior"},
	}

	rows := sqlmock.NewRows([]string{"id", "branch_id", "name", "surname", "position"}).
		AddRow(expectedEmployees[0].ID, expectedEmployees[0].BranchID, expectedEmployees[0].Name, expectedEmployees[0].Surname, expectedEmployees[0].Position).
		AddRow(expectedEmployees[1].ID, expectedEmployees[1].BranchID, expectedEmployees[1].Name, expectedEmployees[1].Surname, expectedEmployees[1].Position)

	mock.ExpectQuery("SELECT.*FROM employees.*WHERE branch_id = \\$1.*ORDER BY id").
		WithArgs(int64(1)).
		WillReturnRows(rows)

	repo := &BranchesRepoPostgres{db: db}

	// Act
	employees, err := repo.GetEmployeesByBranchID(context.Background(), 1)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(employees) != 2 {
		t.Fatalf("expected 2 employees, got %d", len(employees))
	}
	if employees[0].Name != "John" {
		t.Fatalf("expected John, got %s", employees[0].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestGetEmployeesByBranchID_Empty(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	rows := sqlmock.NewRows([]string{"id", "branch_id", "name", "surname", "position"})

	mock.ExpectQuery("SELECT.*FROM employees.*WHERE branch_id = \\$1.*ORDER BY id").
		WithArgs(int64(1)).
		WillReturnRows(rows)

	repo := &BranchesRepoPostgres{db: db}

	// Act
	employees, err := repo.GetEmployeesByBranchID(context.Background(), 1)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(employees) != 0 {
		t.Fatalf("expected 0 employees, got %d", len(employees))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestGetEmployeesByBranchID_QueryError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	expectedErr := errors.New("database error")

	mock.ExpectQuery("SELECT.*FROM employees.*WHERE branch_id = \\$1.*ORDER BY id").
		WithArgs(int64(1)).
		WillReturnError(expectedErr)

	repo := &BranchesRepoPostgres{db: db}

	// Act
	employees, err := repo.GetEmployeesByBranchID(context.Background(), 1)

	// Assert
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if employees != nil {
		t.Fatalf("expected nil employees, got %v", employees)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestGetEmployeesByBranchID_ScanError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	rows := sqlmock.NewRows([]string{"id", "branch_id", "name", "surname", "position"}).
		AddRow("invalid", 1, "John", "Doe", "Specialist")

	mock.ExpectQuery("SELECT.*FROM employees.*WHERE branch_id = \\$1.*ORDER BY id").
		WithArgs(int64(1)).
		WillReturnRows(rows)

	repo := &BranchesRepoPostgres{db: db}

	// Act
	employees, err := repo.GetEmployeesByBranchID(context.Background(), 1)

	// Assert
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if employees != nil {
		t.Fatalf("expected nil employees, got %v", employees)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}
