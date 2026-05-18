package httpserver

import (
	"Online-queue-management-system/libs/auth"
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/branches/internal/application/service"
	"Online-queue-management-system/services/branches/internal/domain"
	"Online-queue-management-system/services/branches/internal/mocks"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createContextWithUser(t *testing.T, user *auth.AccessClaims) context.Context {
	ctx := context.Background()
	// Use auth.ContextWithUser to properly set the user in context
	ctx = auth.ContextWithUser(ctx, user)

	// Добавляем логгер в контекст
	log := logger.New(logger.Config{
		Level:  slog.LevelInfo,
		JSON:   false,
		Source: true,
	})
	ctx = logger.With(ctx, log)

	return ctx
}

func TestGetBranches_Success(t *testing.T) {
	// Arrange
	mockBranches := []domain.Branch{
		{ID: 1, BusinessID: 100, Name: "Branch 1", Address: "Address 1"},
		{ID: 2, BusinessID: 100, Name: "Branch 2", Address: "Address 2"},
	}

	mockRepo := &mocks.MockBranchesRepository{
		GetByBusinessIDFunc: func(ctx context.Context, businessID int64) ([]domain.Branch, error) {
			return mockBranches, nil
		},
	}

	svc := service.New(mockRepo)
	server := NewHttpServer(svc)

	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	req := httptest.NewRequest("GET", "/branches", http.NoBody)
	req = req.WithContext(createContextWithUser(t, user))
	rec := httptest.NewRecorder()

	// Act
	server.GetBranches(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var response []map[string]interface{}
	body, _ := io.ReadAll(rec.Body)
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(response))
	}
}

func TestGetBranches_Unauthorized(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockBranchesRepository{}
	svc := service.New(mockRepo)
	server := NewHttpServer(svc)

	req := httptest.NewRequest("GET", "/branches", http.NoBody)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	// Act
	server.GetBranches(rec, req)

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestGetBranches_ServiceError(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockBranchesRepository{
		GetByBusinessIDFunc: func(ctx context.Context, businessID int64) ([]domain.Branch, error) {
			return nil, domain.ErrForbidden
		},
	}

	svc := service.New(mockRepo)
	server := NewHttpServer(svc)

	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	req := httptest.NewRequest("GET", "/branches", http.NoBody)
	req = req.WithContext(createContextWithUser(t, user))
	rec := httptest.NewRecorder()

	// Act
	server.GetBranches(rec, req)

	// Assert
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}

func TestGetBranchEmployees_Success(t *testing.T) {
	// Arrange
	mockBranch := []domain.Branch{
		{ID: 1, BusinessID: 100, Name: "Branch 1", Address: "Address 1"},
	}
	mockEmployees := []domain.Employee{
		{ID: 1, BranchID: 1, Name: "John", Surname: "Doe", Position: "Specialist"},
		{ID: 2, BranchID: 1, Name: "Jane", Surname: "Smith", Position: "Senior"},
	}

	mockRepo := &mocks.MockBranchesRepository{
		GetByIDFunc: func(ctx context.Context, branchID int64) ([]domain.Branch, error) {
			if branchID == 1 {
				return mockBranch, nil
			}
			return nil, domain.ErrBranchNotFound
		},
		GetEmployeesByBranchIDFunc: func(ctx context.Context, branchID int64) ([]domain.Employee, error) {
			if branchID == 1 {
				return mockEmployees, nil
			}
			return nil, nil
		},
	}

	svc := service.New(mockRepo)
	server := NewHttpServer(svc)

	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	// Create a request using the actual Go 1.22 mux to properly parse path parameters
	mux := http.NewServeMux()
	mux.HandleFunc("GET /branches/{id}/employees", server.GetBranchEmployees)

	req := httptest.NewRequest("GET", "/branches/1/employees", http.NoBody)
	req = req.WithContext(createContextWithUser(t, user))
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var response []map[string]interface{}
	body, _ := io.ReadAll(rec.Body)
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response) != 2 {
		t.Fatalf("expected 2 employees, got %d", len(response))
	}
}

func TestGetBranchEmployees_Unauthorized(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockBranchesRepository{}
	svc := service.New(mockRepo)
	server := NewHttpServer(svc)

	req := httptest.NewRequest("GET", "/branches/1/employees", http.NoBody)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	// Act
	server.GetBranchEmployees(rec, req)

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestGetBranchEmployees_InvalidBranchID(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockBranchesRepository{}
	svc := service.New(mockRepo)
	server := NewHttpServer(svc)

	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	// Create mux to properly parse path parameters
	mux := http.NewServeMux()
	mux.HandleFunc("GET /branches/{id}/employees", server.GetBranchEmployees)

	req := httptest.NewRequest("GET", "/branches/invalid/employees", http.NoBody)
	req = req.WithContext(createContextWithUser(t, user))
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestGetBranchEmployees_NegativeBranchID(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockBranchesRepository{}
	svc := service.New(mockRepo)
	server := NewHttpServer(svc)

	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	// Create mux to properly parse path parameters
	mux := http.NewServeMux()
	mux.HandleFunc("GET /branches/{id}/employees", server.GetBranchEmployees)

	req := httptest.NewRequest("GET", "/branches/-1/employees", http.NoBody)
	req = req.WithContext(createContextWithUser(t, user))
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestGetBranchEmployees_ServiceError(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockBranchesRepository{
		GetByIDFunc: func(ctx context.Context, branchID int64) ([]domain.Branch, error) {
			return []domain.Branch{
				{ID: 1, BusinessID: 100, Name: "Branch 1", Address: "Address 1"},
			}, nil
		},
		GetEmployeesByBranchIDFunc: func(ctx context.Context, branchID int64) ([]domain.Employee, error) {
			return nil, domain.ErrForbidden
		},
	}

	svc := service.New(mockRepo)
	server := NewHttpServer(svc)

	user := &auth.AccessClaims{
		UserID:     1,
		RoleID:     1,
		RoleName:   "business_admin",
		BusinessID: 100,
	}

	// Create mux to properly parse path parameters
	mux := http.NewServeMux()
	mux.HandleFunc("GET /branches/{id}/employees", server.GetBranchEmployees)

	req := httptest.NewRequest("GET", "/branches/1/employees", http.NoBody)
	req = req.WithContext(createContextWithUser(t, user))
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}

func TestWriteJSON(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()
	testData := map[string]string{"test": "data"}

	// Act
	writeJSON(rec, http.StatusOK, testData)

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", contentType)
	}

	var result map[string]string
	body, _ := io.ReadAll(rec.Body)
	if err := json.Unmarshal(bytes.TrimSpace(body), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["test"] != "data" {
		t.Fatalf("expected test=data, got test=%s", result["test"])
	}
}

// Tests for public endpoints

func TestGetPublicServices_Success(t *testing.T) {
	// Arrange
	mockServices := []domain.Service{
		{ID: 1, BranchID: 1, Name: "Service 1", DurationMinutes: 30, Price: 500.0},
		{ID: 2, BranchID: 1, Name: "Service 2", DurationMinutes: 45, Price: 750.0},
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
				return mockServices, nil
			}
			return nil, nil
		},
	}

	svc := service.New(mockRepo)
	server := NewHttpServer(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /public/{registrationSlug}/services", server.GetPublicServices)

	req := httptest.NewRequest("GET", "/public/beautiful-salon/services", http.NoBody)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 services, got %d", len(result))
	}
}

func TestGetPublicBranchesWithService_Success(t *testing.T) {
	// Arrange
	mockBranches := []domain.Branch{
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
				return mockBranches, nil
			}
			return nil, nil
		},
	}

	svc := service.New(mockRepo)
	server := NewHttpServer(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /public/{registrationSlug}/services/{serviceId}/branches", server.GetPublicBranchesWithService)

	req := httptest.NewRequest("GET", "/public/beautiful-salon/services/1/branches", http.NoBody)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 branch, got %d", len(result))
	}
}

func TestGetPublicEmployeesForService_Success(t *testing.T) {
	// Arrange
	mockEmployees := []domain.Employee{
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
				return mockEmployees, nil
			}
			return nil, nil
		},
	}

	svc := service.New(mockRepo)
	server := NewHttpServer(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /public/{registrationSlug}/services/{serviceId}/branches/{branchId}/employees", server.GetPublicEmployeesForService)

	req := httptest.NewRequest("GET", "/public/beautiful-salon/services/1/branches/1/employees", http.NoBody)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 employee, got %d", len(result))
	}
}

func TestGetPublicServices_InvalidSlug_ReturnsBadRequest(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockBranchesRepository{}

	svc := service.New(mockRepo)
	server := NewHttpServer(svc)

	req := httptest.NewRequest("GET", "/public//services", http.NoBody)
	rec := httptest.NewRecorder()

	// Act
	server.GetPublicServices(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
