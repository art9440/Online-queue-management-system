package domain

import "time"

type Client struct {
	ID         int64     `json:"id"`
	Email      *string   `json:"email,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
	Name       string    `json:"name"`
	Surname    string    `json:"surname"`
	TgUsername *string   `json:"tg_username,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
