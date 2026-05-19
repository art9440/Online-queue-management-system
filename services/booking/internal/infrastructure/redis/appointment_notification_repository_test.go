package redis

import (
	"Online-queue-management-system/services/booking/internal/domain"
	"testing"
	"time"
)

func TestBuildAppointmentNotifications(t *testing.T) {
	email := "client@example.com"
	username := "@ClientUser"
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	appointment := domain.CreateAppointmentOutput{
		AppointmentID:   42,
		BranchName:      "Central",
		EmployeeName:    "Alex",
		EmployeeSurname: "Kirsanov",
		ServiceName:     "Consultation",
		StartTime:       now.Add(48 * time.Hour),
	}
	client := domain.ClientInput{
		Email:      &email,
		Phone:      "8 (999) 000-00-00",
		TgUsername: &username,
	}

	notifications := buildAppointmentNotifications(&appointment, &client, now)

	if len(notifications) != 4 {
		t.Fatalf("expected 4 notifications, got %d", len(notifications))
	}

	first := notifications[0]
	if first.payload.ID != "appointment:42:email:1440" {
		t.Fatalf("unexpected first notification id: %s", first.payload.ID)
	}
	if first.scheduledAt != appointment.StartTime.Add(-24*time.Hour) {
		t.Fatalf("unexpected first notification schedule: %s", first.scheduledAt)
	}

	telegram := notifications[1].payload
	if telegram.Channel != "telegram" {
		t.Fatalf("expected telegram channel, got %s", telegram.Channel)
	}
	if telegram.Phone != "+89990000000" {
		t.Fatalf("unexpected normalized phone: %s", telegram.Phone)
	}
	if telegram.Username != "clientuser" {
		t.Fatalf("unexpected normalized username: %s", telegram.Username)
	}
}

func TestBuildAppointmentNotificationsSkipsPastReminderOffsets(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	appointment := domain.CreateAppointmentOutput{
		AppointmentID: 7,
		StartTime:     now.Add(2 * time.Hour),
	}
	client := domain.ClientInput{Phone: "+79990000000"}

	notifications := buildAppointmentNotifications(&appointment, &client, now)

	if len(notifications) != 1 {
		t.Fatalf("expected only one future reminder, got %d", len(notifications))
	}
	if notifications[0].payload.ID != "appointment:7:telegram:60" {
		t.Fatalf("unexpected notification id: %s", notifications[0].payload.ID)
	}
}
