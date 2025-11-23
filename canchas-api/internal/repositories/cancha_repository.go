package repositories

import (
	"canchas-api/internal/domain"
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CanchaRepository interface {
	Create(cancha *domain.Cancha) error
	GetByID(id string) (*domain.Cancha, error)
	GetAll() ([]domain.Cancha, error)
	GetByNumberAndType(number int, tipo string) (*domain.Cancha, error)
	GetByName(name string) (*domain.Cancha, error)
	Update(id string, cancha *domain.Cancha) error
	Delete(id string) error
	// ❌ ELIMINAR: GetByOwnerID(ownerID uint) ([]domain.Cancha, error)
}

type canchaRepository struct {
	collection *mongo.Collection
}

func NewCanchaRepository(db *mongo.Database) CanchaRepository {
	coll := db.Collection(domain.Cancha{}.CollectionName())
	r := &canchaRepository{collection: coll}

	// Índice compuesto único number+type para asegurar unicidad en base de datos
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "number", Value: 1},
			{Key: "type", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := coll.Indexes().CreateOne(ctx, indexModel); err != nil {
		log.Printf("Warning: failed to create unique index on cancha.number+type: %v", err)
	}

	// Índice único en name para evitar nombres duplicados
	indexName := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := coll.Indexes().CreateOne(ctx, indexName); err != nil {
		log.Printf("Warning: failed to create unique index on cancha.name: %v", err)
	}

	return r
}

func (r *canchaRepository) Create(cancha *domain.Cancha) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cancha.ID = primitive.NewObjectID()
	cancha.CreatedAt = time.Now()
	cancha.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, cancha)
	if err != nil {
		// Traducir error de clave duplicada de Mongo a un mensaje entendible
		var we mongo.WriteException
		if errors.As(err, &we) {
			// Revisar qué índice disparó el duplicado
			msg := err.Error()
			for _, e := range we.WriteErrors {
				if e.Code == 11000 {
					if strings.Contains(msg, "name") || strings.Contains(msg, "name_1") {
						return errors.New("Ya existe una cancha con ese nombre.")
					}
					if strings.Contains(msg, "number") || strings.Contains(msg, "number_1") {
						return errors.New("Ya existe una cancha con ese número y de ese tipo.")
					}
					// fallback
					return errors.New("error de clave duplicada")
				}
			}
		}
		return err
	}
	return nil
}

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
			"number":      cancha.Number,
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

// GetByNumber devuelve la cancha que tiene el número indicado.
// Retorna (nil, nil) si no existe.
func (r *canchaRepository) GetByNumberAndType(number int, tipo string) (*domain.Cancha, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cancha domain.Cancha
	err := r.collection.FindOne(ctx, bson.M{"number": number, "type": tipo}).Decode(&cancha)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &cancha, nil
}

// GetByName devuelve la cancha que tiene el nombre indicado.
// Retorna (nil, nil) si no existe.
func (r *canchaRepository) GetByName(name string) (*domain.Cancha, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cancha domain.Cancha
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&cancha)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &cancha, nil
}

// ❌ ELIMINAR método GetByOwnerID
