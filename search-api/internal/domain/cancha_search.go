package domain

import (
	"time"
)

// CanchaSearch representa una cancha indexada en SolR
type CanchaSearch struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Address     string    `json:"address"`
	Number      int       `json:"number"`
	Price       float64   `json:"price"`
	Capacity    int       `json:"capacity"`
	Available   bool      `json:"available"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
