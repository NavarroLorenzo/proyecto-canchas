package dto

import (
	"time"
)

// CreateCanchaRequest - DTO para crear una cancha (SOLO ADMIN)
type CreateCanchaRequest struct {
	Name        string  `json:"name" binding:"required,min=3"`
	Type        string  `json:"type" binding:"required,oneof=futbol tenis basquet paddle voley"`
	Description string  `json:"description" binding:"required"`
	Number      int     `json:"number" binding:"required,gt=0"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Capacity    int     `json:"capacity" binding:"required,gt=0"`
	Available   bool    `json:"available"`
	ImageURL    string  `json:"image_url"`
}

// UpdateCanchaRequest - DTO para actualizar una cancha (SOLO ADMIN)
type UpdateCanchaRequest struct {
	Name        string  `json:"name" binding:"omitempty,min=3"`
	Type        string  `json:"type" binding:"omitempty,oneof=futbol tenis basquet paddle voley"`
	Description string  `json:"description"`
	Number      int     `json:"number"`
	Price       float64 `json:"price" binding:"omitempty,gt=0"`
	Capacity    int     `json:"capacity" binding:"omitempty,gt=0"`
	Available   *bool   `json:"available"` // Pointer para permitir false
	ImageURL    string  `json:"image_url"`
}

// CanchaResponse - DTO para respuesta de cancha
type CanchaResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Number      int     `json:"number"`
	Price       float64 `json:"price"`
	Capacity    int     `json:"capacity"`
	Available   bool    `json:"available"`
	ImageURL    string  `json:"image_url"`
	// ❌ ELIMINAR: OwnerID     uint      `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CanchasListResponse - DTO para lista de canchas
type CanchasListResponse struct {
	Canchas []CanchaResponse `json:"canchas"`
	Total   int64            `json:"total"`
}
