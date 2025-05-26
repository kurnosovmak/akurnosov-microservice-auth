package models

import "time"

type User struct {
	ID                string    `json:"id"`
	Email             string    `json:"email"`
	Password          string    `json:"-"` // Не сериализуется в JSON
	IsVerified        bool      `json:"is_verified"`
	VerificationToken string    `json:"-"` // Не сериализуется в JSON
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
