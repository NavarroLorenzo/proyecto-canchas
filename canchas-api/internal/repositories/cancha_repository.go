package repositories

import (
	"canchas-api/internal/domain"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CanchaRepository interface {
	Create(cancha *domain.Cancha) error
	GetByID(id string) (*domain.Cancha, error)
	GetAll() ([]domain.Cancha, error)
	Update(id string, cancha *domain.Cancha) error
	Delete(id string) error
	GetByOwnerID(ownerID uint) ([]domain.Cancha, error)
}

type canchaRepository struct {
	collection *mongo.Collection
}

// NewCanchaRepository crea una nueva instancia del repositorio
func NewCanchaRepository(db *mongo.Database) CanchaRepository {
	return &canchaRepository{
		collection: db.Collection(domain.Cancha{}.CollectionName()),
	}
}

// Create crea una nueva cancha en MongoDB
func (r *canchaRepository) Create(cancha *domain.Cancha) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cancha.ID = primitive.NewObjectID()
	cancha.CreatedAt = time.Now()
	cancha.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, cancha)
	return err
}

// GetByID obtiene una cancha por su ID
func (r *canchaRepository) GetByID(id string) (*domain.Cancha, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	var cancha domain.Cancha
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&cancha)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("cancha not found")
		}
		return nil, err
	}

	return &cancha, nil
}

// GetAll obtiene todas las canchas
func (r *canchaRepository) GetAll() ([]domain.Cancha, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var canchas []domain.Cancha
	if err := cursor.All(ctx, &canchas); err != nil {
		return nil, err
	}

	return canchas, nil
}

// Update actualiza una cancha existente
func (r *canchaRepository) Update(id string, cancha *domain.Cancha) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid ID format")
	}

	cancha.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"name":        cancha.Name,
			"type":        cancha.Type,
			"description": cancha.Description,
			"location":    cancha.Location,
			"address":     cancha.Address,
			"price":       cancha.Price,
			"capacity":    cancha.Capacity,
			"available":   cancha.Available,
			"image_url":   cancha.ImageURL,
			"updated_at":  cancha.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("cancha not found")
	}

	return nil
}

// Delete elimina una cancha
func (r *canchaRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid ID format")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("cancha not found")
	}

	return nil
}

// GetByOwnerID obtiene todas las canchas de un owner específico
func (r *canchaRepository) GetByOwnerID(ownerID uint) ([]domain.Cancha, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"owner_id": ownerID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var canchas []domain.Cancha
	if err := cursor.All(ctx, &canchas); err != nil {
		return nil, err
	}

	return canchas, nil
}
