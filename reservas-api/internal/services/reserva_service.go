package services

import (
	"errors"
	"fmt"
	"reservas-api/internal/clients"
	"reservas-api/internal/domain"
	"reservas-api/internal/dto"
	"reservas-api/internal/messaging"
	"reservas-api/internal/repositories"
	"reservas-api/internal/utils"
	"time"
)

type ReservaService interface {
	Create(req *dto.CreateReservaRequest) (*dto.ReservaResponse, error)
	GetByID(id string) (*dto.ReservaResponse, error)
	GetAll() (*dto.ReservasListResponse, error)
	GetByUserID(userID uint) (*dto.ReservasListResponse, error)
	GetByCanchaID(canchaID string) (*dto.ReservasListResponse, error)
	Update(id string, req *dto.UpdateReservaRequest) (*dto.ReservaResponse, error)
	Cancel(id string) error
}

type reservaService struct {
	repo         repositories.ReservaRepository
	userClient   clients.UserClient
	canchaClient clients.CanchaClient
	publisher    messaging.RabbitMQPublisher
}

// NewReservaService crea una nueva instancia del servicio
func NewReservaService(
	repo repositories.ReservaRepository,
	userClient clients.UserClient,
	canchaClient clients.CanchaClient,
	publisher messaging.RabbitMQPublisher,
) ReservaService {
	return &reservaService{
		repo:         repo,
		userClient:   userClient,
		canchaClient: canchaClient,
		publisher:    publisher,
	}
}

// Create crea una nueva reserva con validación concurrente
func (s *reservaService) Create(req *dto.CreateReservaRequest) (*dto.ReservaResponse, error) {
	// Variables para almacenar resultados de las validaciones
	var userData *clients.UserResponse
	var canchaData *clients.CanchaResponse
	var duration int
	var totalPrice float64
	var date time.Time

	// 🚀 CÁLCULO CONCURRENTE: Preparar validaciones
	validations := []utils.ConcurrentValidation{
		// Validación 1: Usuario existe
		{
			Name: "user_validation",
			Function: func() dto.ValidationResult {
				valid, user, err := s.userClient.ValidateUser(req.UserID)
				if err != nil || !valid {
					return dto.ValidationResult{
						Valid:   false,
						Message: fmt.Sprintf("user validation failed: %v", err),
					}
				}
				userData = user
				return dto.ValidationResult{Valid: true, Data: user}
			},
		},
		// Validación 2: Cancha existe y está disponible
		{
			Name: "cancha_validation",
			Function: func() dto.ValidationResult {
				valid, cancha, err := s.canchaClient.ValidateCancha(req.CanchaID)
				if err != nil || !valid {
					return dto.ValidationResult{
						Valid:   false,
						Message: fmt.Sprintf("cancha validation failed: %v", err),
					}
				}
				canchaData = cancha
				return dto.ValidationResult{Valid: true, Data: cancha}
			},
		},
		// Validación 3: Calcular duración
		{
			Name: "duration_calculation",
			Function: func() dto.ValidationResult {
				dur, err := utils.CalculateDuration(req.StartTime, req.EndTime)
				if err != nil {
					return dto.ValidationResult{
						Valid:   false,
						Message: fmt.Sprintf("duration calculation failed: %v", err),
					}
				}
				duration = dur
				return dto.ValidationResult{Valid: true, Data: dur}
			},
		},
		// Validación 4: Parsear fecha
		{
			Name: "date_parsing",
			Function: func() dto.ValidationResult {
				parsedDate, err := utils.ParseDate(req.Date)
				if err != nil {
					return dto.ValidationResult{
						Valid:   false,
						Message: fmt.Sprintf("date parsing failed: %v", err),
					}
				}
				// Verificar que la fecha no sea en el pasado
				if parsedDate.Before(time.Now().Truncate(24 * time.Hour)) {
					return dto.ValidationResult{
						Valid:   false,
						Message: "cannot make reservations for past dates",
					}
				}
				date = parsedDate
				return dto.ValidationResult{Valid: true, Data: parsedDate}
			},
		},
	}

	// 🔥 EJECUTAR VALIDACIONES CONCURRENTEMENTE (GoRoutines + Channels + WaitGroup)
	allValid, validationErrors := utils.ExecuteConcurrentValidations(validations)

	if !allValid {
		return nil, errors.New(fmt.Sprintf("validation failed: %v", validationErrors))
	}

	// Calcular precio (después de que todas las validaciones pasaron)
	totalPrice = utils.CalculatePrice(canchaData.Price, duration)

	// Verificar disponibilidad (esto debe ser secuencial para evitar condiciones de carrera)
	available, err := s.repo.CheckAvailability(req.CanchaID, date, req.StartTime, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("error checking availability: %w", err)
	}
	if !available {
		return nil, errors.New("cancha not available for the selected time slot")
	}

	// Crear la reserva
	reserva := &domain.Reserva{
		CanchaID:   req.CanchaID,
		UserID:     req.UserID,
		Date:       date,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Duration:   duration,
		Status:     "confirmed",
		TotalPrice: totalPrice,
		CanchaName: canchaData.Name,
		UserName:   fmt.Sprintf("%s %s", userData.FirstName, userData.LastName),
	}

	if err := s.repo.Create(reserva); err != nil {
		return nil, err
	}

	// Publicar evento a RabbitMQ
	event := messaging.Event{
		Type:      "create",
		Entity:    "reserva",
		EntityID:  reserva.ID.Hex(),
		Data:      reserva,
		Timestamp: time.Now().Unix(),
	}
	if err := s.publisher.PublishEvent(event); err != nil {
		println("Warning: failed to publish event:", err.Error())
	}

	return s.domainToResponse(reserva), nil
}

// GetByID obtiene una reserva por su ID
func (s *reservaService) GetByID(id string) (*dto.ReservaResponse, error) {
	reserva, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.domainToResponse(reserva), nil
}

// GetAll obtiene todas las reservas
func (s *reservaService) GetAll() (*dto.ReservasListResponse, error) {
	reservas, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ReservaResponse, len(reservas))
	for i, reserva := range reservas {
		responses[i] = *s.domainToResponse(&reserva)
	}

	return &dto.ReservasListResponse{
		Reservas: responses,
		Total:    int64(len(reservas)),
	}, nil
}

// GetByUserID obtiene todas las reservas de un usuario
func (s *reservaService) GetByUserID(userID uint) (*dto.ReservasListResponse, error) {
	reservas, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ReservaResponse, len(reservas))
	for i, reserva := range reservas {
		responses[i] = *s.domainToResponse(&reserva)
	}

	return &dto.ReservasListResponse{
		Reservas: responses,
		Total:    int64(len(reservas)),
	}, nil
}

// GetByCanchaID obtiene todas las reservas de una cancha
func (s *reservaService) GetByCanchaID(canchaID string) (*dto.ReservasListResponse, error) {
	reservas, err := s.repo.GetByCanchaID(canchaID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ReservaResponse, len(reservas))
	for i, reserva := range reservas {
		responses[i] = *s.domainToResponse(&reserva)
	}

	return &dto.ReservasListResponse{
		Reservas: responses,
		Total:    int64(len(reservas)),
	}, nil
}

// Update actualiza una reserva existente
func (s *reservaService) Update(id string, req *dto.UpdateReservaRequest) (*dto.ReservaResponse, error) {
	// Obtener la reserva existente
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// No permitir actualizar reservas canceladas
	if existing.Status == "cancelled" {
		return nil, errors.New("cannot update a cancelled reservation")
	}

	// Actualizar campos si se proporcionan
	if req.Date != "" {
		date, err := utils.ParseDate(req.Date)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
		existing.Date = date
	}

	if req.StartTime != "" {
		existing.StartTime = req.StartTime
	}

	if req.EndTime != "" {
		existing.EndTime = req.EndTime
	}

	if req.Status != "" {
		existing.Status = req.Status
	}

	// Recalcular duración y precio si cambiaron las horas
	if req.StartTime != "" || req.EndTime != "" {
		duration, err := utils.CalculateDuration(existing.StartTime, existing.EndTime)
		if err != nil {
			return nil, err
		}
		existing.Duration = duration

		// Obtener precio de la cancha
		_, cancha, err := s.canchaClient.ValidateCancha(existing.CanchaID)
		if err != nil {
			return nil, err
		}
		existing.TotalPrice = utils.CalculatePrice(cancha.Price, duration)
	}

	// Verificar disponibilidad si cambió la fecha o las horas
	if req.Date != "" || req.StartTime != "" || req.EndTime != "" {
		available, err := s.repo.CheckAvailability(existing.CanchaID, existing.Date, existing.StartTime, existing.EndTime)
		if err != nil {
			return nil, err
		}
		if !available {
			return nil, errors.New("cancha not available for the selected time slot")
		}
	}

	if err := s.repo.Update(id, existing); err != nil {
		return nil, err
	}

	// Publicar evento
	event := messaging.Event{
		Type:      "update",
		Entity:    "reserva",
		EntityID:  id,
		Data:      existing,
		Timestamp: time.Now().Unix(),
	}
	if err := s.publisher.PublishEvent(event); err != nil {
		println("Warning: failed to publish event:", err.Error())
	}

	return s.domainToResponse(existing), nil
}

// Cancel cancela una reserva
func (s *reservaService) Cancel(id string) error {
	// Verificar que la reserva existe
	reserva, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if reserva.Status == "cancelled" {
		return errors.New("reservation already cancelled")
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	// Publicar evento
	event := messaging.Event{
		Type:      "cancel",
		Entity:    "reserva",
		EntityID:  id,
		Data:      reserva,
		Timestamp: time.Now().Unix(),
	}
	if err := s.publisher.PublishEvent(event); err != nil {
		println("Warning: failed to publish event:", err.Error())
	}

	return nil
}

// domainToResponse convierte una Reserva del dominio a ReservaResponse DTO
func (s *reservaService) domainToResponse(reserva *domain.Reserva) *dto.ReservaResponse {
	return &dto.ReservaResponse{
		ID:         reserva.ID.Hex(),
		CanchaID:   reserva.CanchaID,
		CanchaName: reserva.CanchaName,
		UserID:     reserva.UserID,
		UserName:   reserva.UserName,
		Date:       utils.FormatDate(reserva.Date),
		StartTime:  reserva.StartTime,
		EndTime:    reserva.EndTime,
		Duration:   reserva.Duration,
		Status:     reserva.Status,
		TotalPrice: reserva.TotalPrice,
		CreatedAt:  reserva.CreatedAt,
		UpdatedAt:  reserva.UpdatedAt,
	}
}
