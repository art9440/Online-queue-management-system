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
}
