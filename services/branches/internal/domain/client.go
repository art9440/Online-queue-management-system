package domain

import "time"

type Client struct {
	ID         int64
	Email      *string
	Phone      *string
	Name       string
	Surname    string
	TgUsername *string
	CreatedAt  time.Time
}
