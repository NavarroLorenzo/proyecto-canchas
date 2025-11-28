package services

import (
	"errors"
	"testing"
	"time"

	"users-api/config"
	"users-api/internal/domain"
	"users-api/internal/dto"
	"users-api/utils"
)

// mockUserRepository implementa UserRepository en memoria para probar el servicio.
type mockUserRepository struct {
	users map[uint]*domain.User
}

// newMockUserRepo inicializa el repositorio en memoria.
func newMockUserRepo() *mockUserRepository {
	return &mockUserRepository{users: make(map[uint]*domain.User)}
}

func (m *mockUserRepository) nextID() uint {
	return uint(len(m.users) + 1)
}

// Create simula un insert y asigna ID/fechas.
func (m *mockUserRepository) Create(user *domain.User) error {
	user.ID = m.nextID()
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) GetByID(id uint) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) GetByUsername(username string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) GetByEmail(email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) GetByUsernameOrEmail(login string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Username == login || u.Email == login {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) GetAll() ([]domain.User, error) { return nil, nil }

func (m *mockUserRepository) Update(user *domain.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return errors.New("user not found")
	}
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) Delete(id uint) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepository) ExistsByUsername(username string) (bool, error) {
	for _, u := range m.users {
		if u.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockUserRepository) ExistsByEmail(email string) (bool, error) {
	for _, u := range m.users {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func TestRegisterSuccess(t *testing.T) {
	// Registro exitoso: hashea password y respeta rol
	repo := newMockUserRepo()
	config.AppConfig = &config.Config{JWTSecret: "test-secret"}
	svc := NewUserService(repo)

	req := &dto.RegisterRequest{
		Username:  "alice",
		Email:     "alice@example.com",
		Password:  "pass123",
		FirstName: "Alice",
		LastName:  "Doe",
	}

	resp, err := svc.Register(req, "admin")
	if err != nil {
		t.Fatalf("se esperaba registro sin error, llegó: %v", err)
	}
	if resp.Username != req.Username || resp.Email != req.Email || resp.Role != "admin" {
		t.Fatalf("la respuesta no coincide con la solicitud: %+v", resp)
	}
	stored, _ := repo.GetByID(resp.ID)
	if stored == nil || stored.Password == req.Password {
		t.Fatalf("la contraseña debería haberse hasheado")
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	// Debe fallar si el username ya existe
	repo := newMockUserRepo()
	_ = repo.Create(&domain.User{
		Username: "bob",
		Email:    "bob@example.com",
		Password: "hashed",
	})
	svc := NewUserService(repo)

	req := &dto.RegisterRequest{
		Username:  "bob",
		Email:     "another@example.com",
		Password:  "pass",
		FirstName: "Bob",
		LastName:  "Smith",
	}

	if _, err := svc.Register(req, "normal"); err == nil {
		t.Fatalf("se esperaba error por username duplicado, llegó nil")
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	// Debe fallar si el email ya existe
	repo := newMockUserRepo()
	_ = repo.Create(&domain.User{
		Username: "charlie",
		Email:    "charlie@example.com",
		Password: "hashed",
	})
	svc := NewUserService(repo)

	req := &dto.RegisterRequest{
		Username:  "other",
		Email:     "charlie@example.com",
		Password:  "pass",
		FirstName: "Other",
		LastName:  "User",
	}

	if _, err := svc.Register(req, "normal"); err == nil {
		t.Fatalf("se esperaba error por email duplicado, llegó nil")
	}
}

func TestLoginSuccess(t *testing.T) {
	// Login correcto debe devolver token y user
	repo := newMockUserRepo()
	config.AppConfig = &config.Config{JWTSecret: "test-secret"}
	hash, _ := utils.HashPassword("pass123")
	_ = repo.Create(&domain.User{
		Username: "david",
		Email:    "david@example.com",
		Password: hash,
		Role:     "normal",
	})
	svc := NewUserService(repo)

	resp, err := svc.Login(&dto.LoginRequest{
		Login:    "david",
		Password: "pass123",
	})
	if err != nil {
		t.Fatalf("login debería ser exitoso, error: %v", err)
	}
	if resp.Token == "" || resp.User.Username != "david" {
		t.Fatalf("respuesta de login inválida: %+v", resp)
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	// Login debe fallar con password incorrecta
	repo := newMockUserRepo()
	config.AppConfig = &config.Config{JWTSecret: "test-secret"}
	hash, _ := utils.HashPassword("pass123")
	_ = repo.Create(&domain.User{
		Username: "eric",
		Email:    "eric@example.com",
		Password: hash,
		Role:     "normal",
	})
	svc := NewUserService(repo)

	if _, err := svc.Login(&dto.LoginRequest{Login: "eric", Password: "wrong"}); err == nil {
		t.Fatalf("se esperaba error por contraseña inválida, llegó nil")
	}
}
