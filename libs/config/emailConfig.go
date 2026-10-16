package config

import "time"

type EmailSenderConfig struct {
	SMTPHost    string
	SMTPPort    int
	SMTPUser    string
	SMTPPass    string
	SendTimeOut time.Duration
}

type QueueConfig struct {
	NumWorkers int
	RateLimit  time.Duration
	WrkTimeOut time.Duration
	// RetryAfter: how long an unacknowledged email waits before another worker retries it.
	RetryAfter time.Duration
	// MaxAttempts: after this many failed deliveries the email goes to the dead letter stream.
	MaxAttempts int
}
