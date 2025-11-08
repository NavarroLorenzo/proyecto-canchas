package services

import (
	"canchas-api/internal/clients"
	"canchas-api/internal/domain"
	"canchas-api/internal/dto"
	"canchas-api/internal/messaging"
	"canchas-api/internal/repositories"
	"errors"
	"time"
)

type CanchaService interface {
	Create(req *dto.CreateCanchaRequest) (*dto.CanchaResponse, error)
	GetByID(id string) (*dto.CanchaResponse, error)
	GetAll() (*dto.CanchasListResponse, error)
	Update(id string, req *dto.UpdateCanchaRequest) (*dto.CanchaResponse, error)
	Delete(id string) error
	GetByOwnerID(ownerID uint) (*dto.CanchasListResponse, error)
}

type canchaService struct {
	repo       repositories.CanchaRepository
	userClient clients.UserClient
	publisher  messaging.RabbitMQPublisher
}

// NewCanchaService crea una nueva instancia del servicio
func NewCanchaService(
	repo repositories.CanchaRepository,
	userClient clients.UserClient,
	publisher messaging.RabbitMQPublisher,
) CanchaService {
	return &canchaService{
		repo:       repo,
		userClient: userClient,
		publisher:  publisher,
	}
}

// Create crea una nueva cancha
func (s *canchaService) Create(req *dto.CreateCanchaRequest) (*dto.CanchaResponse, error) {
	// Validar que el owner existe en users-api
	exists, err := s.userClient.ValidateUser(req.OwnerID)
	if err != nil || !exists {
		return nil, errors.New("invalid owner: user does not exist")
	}

	// Crear la cancha
	cancha := &domain.Cancha{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Location:    req.Location,
		Address:     req.Address,
		Price:       req.Price,
		Capacity:    req.Capacity,
		Available:   req.Available,
		ImageURL:    req.ImageURL,
		OwnerID:     req.OwnerID,
	}

	if err := s.repo.Create(cancha); err != nil {
		return nil, err
	}

	// Publicar evento a RabbitMQ
	event := messaging.Event{
		Type:      "create",
		Entity:    "cancha",
		EntityID:  cancha.ID.Hex(),
		Data:      cancha,
		Timestamp: time.Now().Unix(),
	}
	if err := s.publisher.PublishEvent(event); err != nil {
		// Log error pero no fallar la operación
		// En producción podrías usar un sistema de retry
		println("Warning: failed to publish event:", err.Error())
	}

	return s.domainToResponse(cancha), nil
}

// GetByID obtiene una cancha por su ID
func (s *canchaService) GetByID(id string) (*dto.CanchaResponse, error) {
	cancha, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.domainToResponse(cancha), nil
}

// GetAll obtiene todas las canchas
func (s *canchaService) GetAll() (*dto.CanchasListResponse, error) {
	canchas, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.CanchaResponse, len(canchas))
	for i, cancha := range canchas {
		responses[i] = *s.domainToResponse(&cancha)
	}

	return &dto.CanchasListResponse{
		Canchas: responses,
		Total:   int64(len(canchas)),
	}, nil
}

// Update actualiza una cancha existente
func (s *canchaService) Update(id string, req *dto.UpdateCanchaRequest) (*dto.CanchaResponse, error) {
	// Obtener la cancha existente
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Actualizar solo los campos proporcionados
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Type != "" {
		existing.Type = req.Type
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Location != "" {
		existing.Location = req.Location
	}
	if req.Address != "" {
		existing.Address = req.Address
	}
	if req.Price > 0 {
		existing.Price = req.Price
	}
	if req.Capacity > 0 {
		existing.Capacity = req.Capacity
	}
	if req.Available != nil {
		existing.Available = *req.Available
	}
	if req.ImageURL != "" {
		existing.ImageURL = req.ImageURL
	}

	if err := s.repo.Update(id, existing); err != nil {
		return nil, err
	}

	// Publicar evento a RabbitMQ
	event := messaging.Event{
		Type:      "update",
		Entity:    "cancha",
		EntityID:  id,
		Data:      existing,
		Timestamp: time.Now().Unix(),
	}
	if err := s.publisher.PublishEvent(event); err != nil {
		println("Warning: failed to publish event:", err.Error())
	}

	return s.domainToResponse(existing), nil
}

// Delete elimina una cancha
func (s *canchaService) Delete(id string) error {
	// Verificar que la cancha existe
	cancha, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	// Publicar evento a RabbitMQ
	event := messaging.Event{
		Type:      "delete",
		Entity:    "cancha",
		EntityID:  id,
		Data:      cancha,
		Timestamp: time.Now().Unix(),
	}
	if err := s.publisher.PublishEvent(event); err != nil {
		println("Warning: failed to publish event:", err.Error())
	}

	return nil
}

// GetByOwnerID obtiene todas las canchas de un owner
func (s *canchaService) GetByOwnerID(ownerID uint) (*dto.CanchasListResponse, error) {
	canchas, err := s.repo.GetByOwnerID(ownerID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.CanchaResponse, len(canchas))
	for i, cancha := range canchas {
		responses[i] = *s.domainToResponse(&cancha)
	}

	return &dto.CanchasListResponse{
		Canchas: responses,
		Total:   int64(len(canchas)),
	}, nil
}

// domainToResponse convierte una Cancha del dominio a CanchaResponse DTO
func (s *canchaService) domainToResponse(cancha *domain.Cancha) *dto.CanchaResponse {
	return &dto.CanchaResponse{
		ID:          cancha.ID.Hex(),
		Name:        cancha.Name,
		Type:        cancha.Type,
		Description: cancha.Description,
		Location:    cancha.Location,
		Address:     cancha.Address,
		Price:       cancha.Price,
		Capacity:    cancha.Capacity,
		Available:   cancha.Available,
		ImageURL:    cancha.ImageURL,
		OwnerID:     cancha.OwnerID,
		CreatedAt:   cancha.CreatedAt,
		UpdatedAt:   cancha.UpdatedAt,
	}
}
