package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Cancha struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Type        string             `bson:"type" json:"type"` // "futbol", "tenis", "basquet", "paddle", "voley"
	Description string             `bson:"description" json:"description"`
	Location    string             `bson:"location" json:"location"`
	Address     string             `bson:"address" json:"address"`
	Number      int                `bson:"number" json:"number"`
	Price       float64            `bson:"price" json:"price"`
	Capacity    int                `bson:"capacity" json:"capacity"`
	Available   bool               `bson:"available" json:"available"`
	ImageURL    string             `bson:"image_url" json:"image_url"`
	// ❌ ELIMINAR: OwnerID     uint               `bson:"owner_id" json:"owner_id"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// CollectionName retorna el nombre de la colección en MongoDB
func (Cancha) CollectionName() string {
	return "canchas"
}
