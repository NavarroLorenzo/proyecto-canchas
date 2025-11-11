package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Reserva struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CanchaID   string             `bson:"cancha_id" json:"cancha_id"`     // ID de la cancha (string ObjectID)
	UserID     uint               `bson:"user_id" json:"user_id"`         // ID del usuario (MySQL)
	Date       time.Time          `bson:"date" json:"date"`               // Fecha de la reserva (YYYY-MM-DD)
	StartTime  string             `bson:"start_time" json:"start_time"`   // Hora inicio (HH:MM)
	EndTime    string             `bson:"end_time" json:"end_time"`       // Hora fin (HH:MM)
	Duration   int                `bson:"duration" json:"duration"`       // Duración en minutos
	Status     string             `bson:"status" json:"status"`           // "pending", "confirmed", "cancelled"
	TotalPrice float64            `bson:"total_price" json:"total_price"` // Precio total calculado
	CanchaName string             `bson:"cancha_name" json:"cancha_name"` // Nombre de la cancha (cache)
	UserName   string             `bson:"user_name" json:"user_name"`     // Nombre del usuario (cache)
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// CollectionName retorna el nombre de la colección en MongoDB
func (Reserva) CollectionName() string {
	return "reservas"
}
