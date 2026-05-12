package service

import (
	"Online-queue-management-system/libs/auth"
	"Online-queue-management-system/services/branches/internal/domain"
	"Online-queue-management-system/services/branches/internal/mocks"
	"context"
	"errors"
	"testing"
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
