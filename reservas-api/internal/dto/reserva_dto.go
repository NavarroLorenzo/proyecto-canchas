package dto

import (
	"time"
)

// CreateReservaRequest - DTO para crear una reserva
type CreateReservaRequest struct {
	CanchaID  string `json:"cancha_id" binding:"required"`
	UserID    uint   `json:"user_id" binding:"required"`
	Date      string `json:"date" binding:"required"`             // Formato: "2025-11-15"
	StartTime string `json:"start_time" binding:"required,len=5"` // Formato: "18:00"
	EndTime   string `json:"end_time" binding:"required,len=5"`   // Formato: "19:00"
}

// UpdateReservaRequest - DTO para actualizar una reserva
type UpdateReservaRequest struct {
	Date      string `json:"date"`
	StartTime string `json:"start_time" binding:"omitempty,len=5"`
	EndTime   string `json:"end_time" binding:"omitempty,len=5"`
	Status    string `json:"status" binding:"omitempty,oneof=pending confirmed cancelled"`
}

// ReservaResponse - DTO para respuesta de reserva
type ReservaResponse struct {
	ID         string    `json:"id"`
	CanchaID   string    `json:"cancha_id"`
	CanchaName string    `json:"cancha_name"`
	UserID     uint      `json:"user_id"`
	UserName   string    `json:"user_name"`
	Date       string    `json:"date"` // Formato: "2025-11-15"
	StartTime  string    `json:"start_time"`
	EndTime    string    `json:"end_time"`
	Duration   int       `json:"duration"`
	Status     string    `json:"status"`
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ReservasListResponse - DTO para lista de reservas
type ReservasListResponse struct {
	Reservas []ReservaResponse `json:"reservas"`
	Total    int64             `json:"total"`
}

// ValidationResult - Resultado de validación concurrente
type ValidationResult struct {
	Valid   bool
	Message string
	Data    interface{}
}
