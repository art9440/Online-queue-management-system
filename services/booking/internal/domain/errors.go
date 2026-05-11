package domain

import "errors"

var (
	ErrInvalidStartTime        = errors.New("invalid start time")
	ErrInvalidClient           = errors.New("invalid client")
	ErrInvalidClientContact    = errors.New("invalid client contact")
	ErrTimeSlotBusy            = errors.New("time slot is already busy")
	ErrAppointmentNotAvailable = errors.New("appointment is not available for selected employee, service, branch or time")
)
