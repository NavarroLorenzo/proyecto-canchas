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

// mockCanchaRepository implementa el repositorio en memoria para probar el servicio sin Mongo.
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

// mockPublisher guarda eventos publicados para verificar que se emitan.
type mockPublisher struct {
	events []messaging.Event
}

func (m *mockPublisher) PublishEvent(e messaging.Event) error {
	m.events = append(m.events, e)
	return nil
}

func (m *mockPublisher) Close() error { return nil }

// mockReservaClient cumple la interfaz y permite extender tests sin llamar a reservas reales.
type mockReservaClient struct{}

func (m *mockReservaClient) DeleteByCanchaID(id string) error { return nil }

func TestCreateCancha_Success(t *testing.T) {
	// Caso feliz: crea cancha nueva y emite evento create
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
		Price:       10, // precio base
		Capacity:    5,  // capacidad baja aplica fee mínimo
		Available:   true,
	}

	resp, err := svc.Create(req)
	if err != nil {
		t.Fatalf("se esperaba sin error, llegó %v", err)
	}
	if resp.Name != req.Name || resp.Type != req.Type || resp.Number != req.Number {
		t.Fatalf("la respuesta no coincide con la solicitud: %+v", resp)
	}
	// Precio esperado: base 10 + 5% + fee 8 = 18.5
	if resp.Price != 18.5 {
		t.Fatalf("precio final incorrecto, got %v", resp.Price)
	}
	if len(pub.events) != 1 || pub.events[0].Type != "create" {
		t.Fatalf("se esperaba un evento create, llegaron %+v", pub.events)
	}
}

func TestCreateCancha_DuplicateNumberAndType(t *testing.T) {
	repo := newMockRepo()
	// Caso de duplicado por número+tipo
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
		t.Fatalf("se esperaba error de número/tipo duplicado, llegó nil")
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
		t.Fatalf("se esperaba error de nombre duplicado, llegó nil")
	}
}
