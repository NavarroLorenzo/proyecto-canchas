package services

import (
	"errors"
	"users-api/internal/domain"
	"users-api/internal/dto"
	"users-api/internal/repositories"
	"users-api/utils"
)

type UserService interface {
	Register(req *dto.RegisterRequest, role string) (*dto.UserResponse, error)
	Login(req *dto.LoginRequest) (*dto.LoginResponse, error)
	GetByID(id uint) (*dto.UserResponse, error)
	GetAll() (*dto.UsersListResponse, error)
	Update(id uint, req *dto.RegisterRequest) (*dto.UserResponse, error)
	Delete(id uint) error
}

type userService struct {
	repo repositories.UserRepository
}

// NewUserService crea una nueva instancia del servicio
func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

// Register registra un nuevo usuario
func (s *userService) Register(req *dto.RegisterRequest, role string) (*dto.UserResponse, error) {
	// Validar si el username ya existe
	exists, err := s.repo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	// Validar si el email ya existe
	exists, err = s.repo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	// Hashear la contraseña
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("error hashing password")
	}

	// Crear el usuario
	user := &domain.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      role,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return s.domainToResponse(user), nil
}

// Login valida las credenciales y retorna un token
func (s *userService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// Buscar usuario por username o email
	user, err := s.repo.GetByUsernameOrEmail(req.Login)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Verificar la contraseña
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	// Generar token JWT
	token, err := utils.GenerateToken(user)
	if err != nil {
		return nil, errors.New("error generating token")
	}

	return &dto.LoginResponse{
		Token: token,
		User:  *s.domainToResponse(user),
	}, nil
}

// GetByID obtiene un usuario por su ID
func (s *userService) GetByID(id uint) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.domainToResponse(user), nil
}

// GetAll obtiene todos los usuarios
func (s *userService) GetAll() (*dto.UsersListResponse, error) {
	users, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *s.domainToResponse(&user)
	}

	return &dto.UsersListResponse{
		Users: userResponses,
		Total: int64(len(users)),
	}, nil
}

// Update actualiza un usuario
func (s *userService) Update(id uint, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Actualizar campos
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Email = req.Email

	// Si se proporciona una nueva contraseña, hashearla
	if req.Password != "" {
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			return nil, errors.New("error hashing password")
		}
		user.Password = hashedPassword
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	return s.domainToResponse(user), nil
}

// Delete elimina un usuario
func (s *userService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// domainToResponse convierte un User del dominio a UserResponse DTO
func (s *userService) domainToResponse(user *domain.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}
