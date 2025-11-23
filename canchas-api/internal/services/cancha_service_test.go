package services

import (
	"errors"
	"testing"
	"time"

	"canchas-api/internal/domain"
	"canchas-api/internal/dto"
	"canchas-api/internal/messaging"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// mockCanchaRepository is an in-memory implementation of CanchaRepository for tests.
type mockCanchaRepository struct {
	canchas map[string]*domain.Cancha
}

func newMockRepo() *mockCanchaRepository {
	return &mockCanchaRepository{canchas: map[string]*domain.Cancha{}}
}

func (m *mockCanchaRepository) Create(c *domain.Cancha) error {
	id := primitive.NewObjectID()
	c.ID = id
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	m.canchas[id.Hex()] = c
	return nil
}

func (m *mockCanchaRepository) GetByID(id string) (*domain.Cancha, error) {
	c, ok := m.canchas[id]
	if !ok {
		return nil, errors.New("cancha not found")
	}
	return c, nil
}

func (m *mockCanchaRepository) GetAll() ([]domain.Cancha, error) { return nil, nil }

func (m *mockCanchaRepository) GetByNumberAndType(number int, tipo string) (*domain.Cancha, error) {
	for _, c := range m.canchas {
		if c.Number == number && c.Type == tipo {
			return c, nil
		}
	}
	return nil, nil
}

func (m *mockCanchaRepository) GetByName(name string) (*domain.Cancha, error) {
	for _, c := range m.canchas {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, nil
}

func (m *mockCanchaRepository) Update(id string, cancha *domain.Cancha) error { return nil }
func (m *mockCanchaRepository) Delete(id string) error                        { return nil }

// mockPublisher captures published events without hitting RabbitMQ.
type mockPublisher struct {
	events []messaging.Event
}

func (m *mockPublisher) PublishEvent(e messaging.Event) error {
	m.events = append(m.events, e)
	return nil
}

func (m *mockPublisher) Close() error { return nil }

// mockReservaClient is unused in current tests but satisfies the interface.
type mockReservaClient struct{}

func (m *mockReservaClient) DeleteByCanchaID(id string) error { return nil }

func TestCreateCancha_Success(t *testing.T) {
	repo := newMockRepo()
	pub := &mockPublisher{}
	svc := NewCanchaService(repo, pub, &mockReservaClient{})

	req := &dto.CreateCanchaRequest{
		Name:        "Cancha Uno",
		Type:        "futbol",
		Description: "desc",
		Location:    "loc",
		Address:     "addr",
		Number:      1,
		Price:       10,
		Capacity:    5,
		Available:   true,
	}

	resp, err := svc.Create(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Name != req.Name || resp.Type != req.Type || resp.Number != req.Number {
		t.Fatalf("response does not match request: %+v", resp)
	}
	if len(pub.events) != 1 || pub.events[0].Type != "create" {
		t.Fatalf("expected one create event, got %+v", pub.events)
	}
}

func TestCreateCancha_DuplicateNumberAndType(t *testing.T) {
	repo := newMockRepo()
	// Seed repository with one cancha to trigger duplicate validation.
	_ = repo.Create(&domain.Cancha{
		Name:     "Existente",
		Type:     "futbol",
		Number:   1,
		Price:    10,
		Capacity: 5,
	})
	svc := NewCanchaService(repo, &mockPublisher{}, &mockReservaClient{})

	req := &dto.CreateCanchaRequest{
		Name:        "Nueva",
		Type:        "futbol",
		Description: "desc",
		Number:      1,
		Price:       10,
		Capacity:    5,
	}

	if _, err := svc.Create(req); err == nil {
		t.Fatalf("expected duplicate number/type error, got nil")
	}
}

func TestCreateCancha_DuplicateName(t *testing.T) {
	repo := newMockRepo()
	_ = repo.Create(&domain.Cancha{
		Name:     "Repetida",
		Type:     "futbol",
		Number:   2,
		Price:    15,
		Capacity: 8,
	})
	svc := NewCanchaService(repo, &mockPublisher{}, &mockReservaClient{})

	req := &dto.CreateCanchaRequest{
		Name:        "Repetida",
        Type:        "tenis",
		Description: "desc",
		Number:      3,
		Price:       12,
		Capacity:    4,
	}

	if _, err := svc.Create(req); err == nil {
		t.Fatalf("expected duplicate name error, got nil")
	}
}
