package models

import (
	"time"
)

type URL struct {
	ID          int       `json:"id"`
	Original    string    `json:"original"`
	ShortCode   string    `json:"shortCode"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	AccessCount int       `json:"accessCount"`
}
