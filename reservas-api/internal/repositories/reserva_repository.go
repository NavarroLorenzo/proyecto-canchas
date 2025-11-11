package repositories

import (
	"context"
	"errors"
	"reservas-api/internal/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReservaRepository interface {
	Create(reserva *domain.Reserva) error
	GetByID(id string) (*domain.Reserva, error)
	GetAll() ([]domain.Reserva, error)
	GetByUserID(userID uint) ([]domain.Reserva, error)
	GetByCanchaID(canchaID string) ([]domain.Reserva, error)
	Update(id string, reserva *domain.Reserva) error
	Delete(id string) error
	CheckAvailability(canchaID string, date time.Time, startTime, endTime string) (bool, error)
}

type reservaRepository struct {
	collection *mongo.Collection
}

// NewReservaRepository crea una nueva instancia del repositorio
func NewReservaRepository(db *mongo.Database) ReservaRepository {
	return &reservaRepository{
		collection: db.Collection(domain.Reserva{}.CollectionName()),
	}
}

// Create crea una nueva reserva en MongoDB
func (r *reservaRepository) Create(reserva *domain.Reserva) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reserva.ID = primitive.NewObjectID()
	reserva.CreatedAt = time.Now()
	reserva.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, reserva)
	return err
}

// GetByID obtiene una reserva por su ID
func (r *reservaRepository) GetByID(id string) (*domain.Reserva, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	var reserva domain.Reserva
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&reserva)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("reserva not found")
		}
		return nil, err
	}

	return &reserva, nil
}

// GetAll obtiene todas las reservas
func (r *reservaRepository) GetAll() ([]domain.Reserva, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reservas []domain.Reserva
	if err := cursor.All(ctx, &reservas); err != nil {
		return nil, err
	}

	return reservas, nil
}

// GetByUserID obtiene todas las reservas de un usuario
func (r *reservaRepository) GetByUserID(userID uint) ([]domain.Reserva, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reservas []domain.Reserva
	if err := cursor.All(ctx, &reservas); err != nil {
		return nil, err
	}

	return reservas, nil
}

// GetByCanchaID obtiene todas las reservas de una cancha
func (r *reservaRepository) GetByCanchaID(canchaID string) ([]domain.Reserva, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"cancha_id": canchaID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reservas []domain.Reserva
	if err := cursor.All(ctx, &reservas); err != nil {
		return nil, err
	}

	return reservas, nil
}

// Update actualiza una reserva existente
func (r *reservaRepository) Update(id string, reserva *domain.Reserva) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid ID format")
	}

	reserva.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"date":        reserva.Date,
			"start_time":  reserva.StartTime,
			"end_time":    reserva.EndTime,
			"duration":    reserva.Duration,
			"status":      reserva.Status,
			"total_price": reserva.TotalPrice,
			"updated_at":  reserva.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("reserva not found")
	}

	return nil
}

// Delete elimina una reserva (cambia estado a cancelled)
func (r *reservaRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid ID format")
	}

	// En lugar de eliminar, cambiar el estado a "cancelled"
	update := bson.M{
		"$set": bson.M{
			"status":     "cancelled",
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("reserva not found")
	}

	return nil
}

// CheckAvailability verifica si hay conflicto de horarios para una cancha en una fecha específica
func (r *reservaRepository) CheckAvailability(canchaID string, date time.Time, startTime, endTime string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Buscar reservas de la misma cancha en la misma fecha que NO estén canceladas
	filter := bson.M{
		"cancha_id": canchaID,
		"date":      date,
		"status":    bson.M{"$ne": "cancelled"},
		"$or": []bson.M{
			// Caso 1: La nueva reserva empieza durante una reserva existente
			{
				"start_time": bson.M{"$lte": startTime},
				"end_time":   bson.M{"$gt": startTime},
			},
			// Caso 2: La nueva reserva termina durante una reserva existente
			{
				"start_time": bson.M{"$lt": endTime},
				"end_time":   bson.M{"$gte": endTime},
			},
			// Caso 3: La nueva reserva contiene completamente una reserva existente
			{
				"start_time": bson.M{"$gte": startTime},
				"end_time":   bson.M{"$lte": endTime},
			},
		},
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	// Si count > 0, significa que hay conflicto (no está disponible)
	return count == 0, nil
}
