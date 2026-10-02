package service

import (
	"Online-queue-management-system/services/booking/internal/domain"
	"context"
	"errors"
	"testing"
	"time"
)

func TestCreateAppointment_WhenRegistrationSlugIsEmpty_ShouldReject(t *testing.T) {
	repo := &fakeBookingRepo{}
	svc := New(repo, nil, nil, nil)

	_, err := svc.CreateAppointment(context.Background(), &domain.CreateAppointmentInput{
		Client:     domain.ClientInput{Phone: "+79990000000", Name: "Ivan", Surname: "Ivanov"},
		BranchID:   1,
		EmployeeID: 1,
		ServiceID:  1,
		StartTime:  time.Now().Add(time.Hour),
	})

	if !errors.Is(err, domain.ErrInvalidRegistrationSlug) {
		t.Fatalf("expected ErrInvalidRegistrationSlug, got %v", err)
	}
	if repo.calls != 0 {
		t.Fatalf("expected no repository calls, got %d", repo.calls)
	}
}

type fakeBookingRepo struct {
	BookingRepository
	calls int
}

func (r *fakeBookingRepo) CheckClientExists(context.Context, domain.ClientInput) (int64, bool, error) {
	r.calls++
	return 0, false, nil
}

func (r *fakeBookingRepo) CreateClient(context.Context, domain.ClientInput) (int64, error) {
	r.calls++
	return 1, nil
}

func (r *fakeBookingRepo) CreateAppointment(
	context.Context,
	int64,
	*domain.CreateAppointmentInput,
) (domain.CreateAppointmentOutput, error) {
	r.calls++
	return domain.CreateAppointmentOutput{}, nil
}
