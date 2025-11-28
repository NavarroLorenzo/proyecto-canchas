package services

import (
	"errors"
	"testing"
	"time"

	"reservas-api/config"
	"reservas-api/internal/clients"
	"reservas-api/internal/domain"
	"reservas-api/internal/dto"
	"reservas-api/internal/messaging"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Mocks para aislar el servicio
type mockReservaRepository struct {
	created        *domain.Reserva
	availabilityOk bool
}

func (m *mockReservaRepository) Create(reserva *domain.Reserva) error {
	m.created = reserva
	// simula persistencia asignando ID y fechas
	reserva.ID = primitive.NewObjectID()
	reserva.CreatedAt = time.Now()
	reserva.UpdatedAt = time.Now()
	return nil
}
func (m *mockReservaRepository) GetByID(id string) (*domain.Reserva, error) {
	return nil, errors.New("not implemented")
}
func (m *mockReservaRepository) GetAll() ([]domain.Reserva, error) { return nil, nil }
func (m *mockReservaRepository) GetByUserID(userID uint) ([]domain.Reserva, error) {
	return nil, nil
}
func (m *mockReservaRepository) GetByCanchaID(canchaID string) ([]domain.Reserva, error) {
	return nil, nil
}
func (m *mockReservaRepository) DeleteByCanchaID(canchaID string) (int64, error) { return 0, nil }
func (m *mockReservaRepository) Update(id string, reserva *domain.Reserva) error { return nil }
func (m *mockReservaRepository) Delete(id string) error                          { return nil }
func (m *mockReservaRepository) CheckAvailability(canchaID string, date time.Time, startTime, endTime string) (bool, error) {
	return m.availabilityOk, nil
}

type mockUserClient struct {
	valid bool
	data  *clients.UserResponse
	err   error
}

func (m *mockUserClient) ValidateUser(userID uint, token string) (bool, *clients.UserResponse, error) {
	return m.valid, m.data, m.err
}

type mockCanchaClient struct {
	valid bool
	data  *clients.CanchaResponse
	err   error
}

func (m *mockCanchaClient) ValidateCancha(canchaID string) (bool, *clients.CanchaResponse, error) {
	return m.valid, m.data, m.err
}

type mockPublisher struct {
	events []messaging.Event
}

func (m *mockPublisher) PublishEvent(e messaging.Event) error {
	m.events = append(m.events, e)
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}

// Tests
func TestCreateReservaSuccess(t *testing.T) {
	repo := &mockReservaRepository{availabilityOk: true}
	userCli := &mockUserClient{
		valid: true,
		data: &clients.UserResponse{
			ID:        1,
			FirstName: "Alice",
			LastName:  "Doe",
		},
	}
	canchaCli := &mockCanchaClient{
		valid: true,
		data: &clients.CanchaResponse{
			ID:        "c1",
			Name:      "Cancha Uno",
			Type:      "futbol",
			Price:     100,
			Available: true,
		},
	}
	pub := &mockPublisher{}
	config.AppConfig = &config.Config{}

	svc := NewReservaService(repo, userCli, canchaCli, pub)

	req := &dto.CreateReservaRequest{
		CanchaID:  "c1",
		UserID:    1,
		Date:      time.Now().Add(24 * time.Hour).Format("2006-01-02"),
		StartTime: "10:00",
		EndTime:   "11:00",
	}

	resp, err := svc.Create(req, "token")
	if err != nil {
		t.Fatalf("se esperaba reserva creada sin error, llego: %v", err)
	}
	if resp.CanchaID != "c1" || resp.UserID != 1 {
		t.Fatalf("respuesta inesperada: %+v", resp)
	}
	if len(pub.events) != 1 || pub.events[0].Type != "create" {
		t.Fatalf("debe publicarse un evento create, eventos: %+v", pub.events)
	}
}

func TestCreateReservaUnavailable(t *testing.T) {
	repo := &mockReservaRepository{availabilityOk: false}
	userCli := &mockUserClient{valid: true, data: &clients.UserResponse{ID: 1, FirstName: "Test", LastName: "User"}}
	canchaCli := &mockCanchaClient{valid: true, data: &clients.CanchaResponse{ID: "c1", Name: "Cancha", Type: "futbol", Price: 100, Available: true}}
	pub := &mockPublisher{}
	config.AppConfig = &config.Config{}

	svc := NewReservaService(repo, userCli, canchaCli, pub)

	req := &dto.CreateReservaRequest{
		CanchaID:  "c1",
		UserID:    1,
		Date:      time.Now().Add(24 * time.Hour).Format("2006-01-02"),
		StartTime: "10:00",
		EndTime:   "11:00",
	}

	if _, err := svc.Create(req, "token"); err == nil {
		t.Fatalf("se esperaba error por disponibilidad, llegó nil")
	}
}
