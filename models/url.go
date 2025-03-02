package models

import (
	"time"
)

// URL represents a shortened URL entry in the database
type URL struct {
	ID          int       `json:"id"`
	Original    string    `json:"original"`
	ShortCode   string    `json:"shortCode"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	AccessCount int       `json:"accessCount"`
}
