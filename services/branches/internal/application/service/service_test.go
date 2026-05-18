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

func TestGetBranchesForUser_WhenBusinessAdmin_ShouldReturnBranchesByBusinessID(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewBranchesRepository()
	svc := New(repo)
	repo.ByBusinessID[7] = []domain.Branch{
		{ID: 1, BusinessID: 7, Name: "Central", Address: "Main street"},
		{ID: 2, BusinessID: 7, Name: "Left Bank", Address: "Second street"},
	}

	branches, err := svc.GetBranchesForUser(ctx, &auth.AccessClaims{
		RoleName:   string(domain.RoleBusinessAdmin),
		BusinessID: 7,
	})
	if err != nil {
		t.Fatalf("get branches: %v", err)
	}

	if repo.BusinessIDCalls != 1 {
		t.Fatalf("expected GetByBusinessID to be called once, got %d", repo.BusinessIDCalls)
	}
	if repo.LastBusinessID != 7 {
		t.Fatalf("expected business id 7, got %d", repo.LastBusinessID)
	}
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}
}

func TestGetBranchesForUser_WhenManagerHasBranchID_ShouldReturnOnlyManagerBranch(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewBranchesRepository()
	svc := New(repo)
	branchID := int64(11)
	repo.ByID[branchID] = []domain.Branch{
		{ID: branchID, BusinessID: 7, Name: "Central", Address: "Main street"},
	}

	branches, err := svc.GetBranchesForUser(ctx, &auth.AccessClaims{
		RoleName: string(domain.RoleManager),
		BranchID: &branchID,
	})
	if err != nil {
		t.Fatalf("get branches: %v", err)
	}

	if repo.IDCalls != 1 {
		t.Fatalf("expected GetByID to be called once, got %d", repo.IDCalls)
	}
	if repo.LastBranchID != branchID {
		t.Fatalf("expected branch id %d, got %d", branchID, repo.LastBranchID)
	}
	if len(branches) != 1 || branches[0].ID != branchID {
		t.Fatalf("unexpected branches: %#v", branches)
	}
}

func TestGetBranchesForUser_WhenManagerHasNoBranchID_ShouldReturnBranchNotFound(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewBranchesRepository()
	svc := New(repo)

	_, err := svc.GetBranchesForUser(ctx, &auth.AccessClaims{
		RoleName: string(domain.RoleManager),
	})
	if !errors.Is(err, domain.ErrBranchNotFound) {
		t.Fatalf("expected branch not found, got %v", err)
	}
	if repo.IDCalls != 0 || repo.BusinessIDCalls != 0 {
		t.Fatalf("expected repository not to be called, got idCalls=%d businessIDCalls=%d", repo.IDCalls, repo.BusinessIDCalls)
	}
}

func TestGetBranchesForUser_WhenRoleIsNotAllowed_ShouldReturnForbidden(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewBranchesRepository()
	svc := New(repo)

	_, err := svc.GetBranchesForUser(ctx, &auth.AccessClaims{
		RoleName: string(domain.RoleEmployee),
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if repo.IDCalls != 0 || repo.BusinessIDCalls != 0 {
		t.Fatalf("expected repository not to be called, got idCalls=%d businessIDCalls=%d", repo.IDCalls, repo.BusinessIDCalls)
	}
}

func TestGetBranchesForUser_WhenRepositoryFails_ShouldReturnRepositoryError(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewBranchesRepository()
	repo.Err = errors.New("db failed")
	svc := New(repo)

	_, err := svc.GetBranchesForUser(ctx, &auth.AccessClaims{
		RoleName:   string(domain.RoleBusinessAdmin),
		BusinessID: 7,
	})
	if !errors.Is(err, repo.Err) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestGetBranchClients_WhenBusinessAdminOwnsBranch_ShouldReturnClients(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewBranchesRepository()
	svc := New(repo)
	repo.BranchBusiness[11] = 7
	repo.ClientsByBranchID[11] = []domain.Client{{ID: 5, Name: "Alex", Surname: "Stone"}}

	clients, err := svc.GetBranchClients(ctx, &auth.AccessClaims{
		RoleName:   string(domain.RoleBusinessAdmin),
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

func TestGetBranchBookings_WhenManagerRequestsAnotherBranch_ShouldReturnForbidden(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewBranchesRepository()
	svc := New(repo)
	managerBranchID := int64(11)

	_, err := svc.GetBranchBookings(ctx, &auth.AccessClaims{
		RoleName: string(domain.RoleManager),
		BranchID: &managerBranchID,
	}, 12, time.Now())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if repo.BookingsByBranchCalls != 0 {
		t.Fatalf("expected bookings repository not to be called, got %d", repo.BookingsByBranchCalls)
	}
}
