package domain

import "errors"

var (
	ErrInvalidStartTime        = errors.New("invalid start time")
	ErrInvalidClient           = errors.New("invalid client")
	ErrInvalidClientContact    = errors.New("invalid client contact")
	ErrInvalidAppointmentID    = errors.New("invalid appointment id")
	ErrInvalidRegistrationSlug = errors.New("invalid registration slug")
	ErrTimeSlotBusy            = errors.New("time slot is already busy")
	ErrAppointmentNotAvailable = errors.New("appointment is not available for selected employee, service, branch or time")
	ErrAppointmentNotFound     = errors.New("appointment not found")
	ErrAppointmentCancelled    = errors.New("appointment is already cancelled")
	ErrAppointmentCompleted    = errors.New("completed appointment cannot be cancelled")
)
