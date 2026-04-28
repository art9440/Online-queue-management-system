package service

import (
	"Online-queue-management-system/libs/auth"
	"Online-queue-management-system/services/branches/internal/domain"
	"context"
	"errors"
	"testing"
)

func TestGetBranchesForUser_WhenBusinessAdmin_ShouldReturnBranchesByBusinessID(t *testing.T) {
	ctx := context.Background()
	repo := newFakeBranchesRepository()
	svc := New(repo)
	repo.byBusinessID[7] = []domain.Branch{
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

	if repo.businessIDCalls != 1 {
		t.Fatalf("expected GetByBusinessID to be called once, got %d", repo.businessIDCalls)
	}
	if repo.lastBusinessID != 7 {
		t.Fatalf("expected business id 7, got %d", repo.lastBusinessID)
	}
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}
}

func TestGetBranchesForUser_WhenManagerHasBranchID_ShouldReturnOnlyManagerBranch(t *testing.T) {
	ctx := context.Background()
	repo := newFakeBranchesRepository()
	svc := New(repo)
	branchID := int64(11)
	repo.byID[branchID] = []domain.Branch{
		{ID: branchID, BusinessID: 7, Name: "Central", Address: "Main street"},
	}

	branches, err := svc.GetBranchesForUser(ctx, &auth.AccessClaims{
		RoleName: string(domain.RoleManager),
		BranchID: &branchID,
	})
	if err != nil {
		t.Fatalf("get branches: %v", err)
	}

	if repo.idCalls != 1 {
		t.Fatalf("expected GetByID to be called once, got %d", repo.idCalls)
	}
	if repo.lastBranchID != branchID {
		t.Fatalf("expected branch id %d, got %d", branchID, repo.lastBranchID)
	}
	if len(branches) != 1 || branches[0].ID != branchID {
		t.Fatalf("unexpected branches: %#v", branches)
	}
}

func TestGetBranchesForUser_WhenManagerHasNoBranchID_ShouldReturnBranchNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeBranchesRepository()
	svc := New(repo)

	_, err := svc.GetBranchesForUser(ctx, &auth.AccessClaims{
		RoleName: string(domain.RoleManager),
	})
	if !errors.Is(err, domain.ErrBranchNotFound) {
		t.Fatalf("expected branch not found, got %v", err)
	}
	if repo.idCalls != 0 || repo.businessIDCalls != 0 {
		t.Fatalf("expected repository not to be called, got idCalls=%d businessIDCalls=%d", repo.idCalls, repo.businessIDCalls)
	}
}

func TestGetBranchesForUser_WhenRoleIsNotAllowed_ShouldReturnForbidden(t *testing.T) {
	ctx := context.Background()
	repo := newFakeBranchesRepository()
	svc := New(repo)

	_, err := svc.GetBranchesForUser(ctx, &auth.AccessClaims{
		RoleName: string(domain.RoleEmployee),
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if repo.idCalls != 0 || repo.businessIDCalls != 0 {
		t.Fatalf("expected repository not to be called, got idCalls=%d businessIDCalls=%d", repo.idCalls, repo.businessIDCalls)
	}
}

func TestGetBranchesForUser_WhenRepositoryFails_ShouldReturnRepositoryError(t *testing.T) {
	ctx := context.Background()
	repo := newFakeBranchesRepository()
	repo.err = errors.New("db failed")
	svc := New(repo)

	_, err := svc.GetBranchesForUser(ctx, &auth.AccessClaims{
		RoleName:   string(domain.RoleBusinessAdmin),
		BusinessID: 7,
	})
	if !errors.Is(err, repo.err) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

type fakeBranchesRepository struct {
	byBusinessID    map[int64][]domain.Branch
	byID            map[int64][]domain.Branch
	lastBusinessID  int64
	lastBranchID    int64
	businessIDCalls int
	idCalls         int
	err             error
}

func newFakeBranchesRepository() *fakeBranchesRepository {
	return &fakeBranchesRepository{
		byBusinessID: make(map[int64][]domain.Branch),
		byID:         make(map[int64][]domain.Branch),
	}
}

func (r *fakeBranchesRepository) GetByBusinessID(_ context.Context, businessID int64) ([]domain.Branch, error) {
	r.businessIDCalls++
	r.lastBusinessID = businessID
	if r.err != nil {
		return nil, r.err
	}
	return r.byBusinessID[businessID], nil
}

func (r *fakeBranchesRepository) GetByID(_ context.Context, branchID int64) ([]domain.Branch, error) {
	r.idCalls++
	r.lastBranchID = branchID
	if r.err != nil {
		return nil, r.err
	}
	branches, ok := r.byID[branchID]
	if !ok {
		return nil, domain.ErrBranchNotFound
	}
	return branches, nil
}
