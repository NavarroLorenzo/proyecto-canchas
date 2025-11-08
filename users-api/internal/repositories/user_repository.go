package repositories

import (
	"errors"
	"users-api/internal/domain"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id uint) (*domain.User, error)
	GetByUsername(username string) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	GetByUsernameOrEmail(login string) (*domain.User, error)
	GetAll() ([]domain.User, error)
	Update(user *domain.User) error
	Delete(id uint) error
	ExistsByUsername(username string) (bool, error)
	ExistsByEmail(email string) (bool, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository crea una nueva instancia del repositorio
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create crea un nuevo usuario en la base de datos
func (r *userRepository) Create(user *domain.User) error {
	result := r.db.Create(user)
	return result.Error
}

// GetByID obtiene un usuario por su ID
func (r *userRepository) GetByID(id uint) (*domain.User, error) {
	var user domain.User
	result := r.db.First(&user, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}

	return &user, nil
}

// GetByUsername obtiene un usuario por su username
func (r *userRepository) GetByUsername(username string) (*domain.User, error) {
	var user domain.User
	result := r.db.Where("username = ?", username).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}

	return &user, nil
}

// GetByEmail obtiene un usuario por su email
func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	result := r.db.Where("email = ?", email).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}

	return &user, nil
}

// GetByUsernameOrEmail obtiene un usuario por username o email
func (r *userRepository) GetByUsernameOrEmail(login string) (*domain.User, error) {
	var user domain.User
	result := r.db.Where("username = ? OR email = ?", login, login).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}

	return &user, nil
}

// GetAll obtiene todos los usuarios
func (r *userRepository) GetAll() ([]domain.User, error) {
	var users []domain.User
	result := r.db.Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

// Update actualiza un usuario
func (r *userRepository) Update(user *domain.User) error {
	result := r.db.Save(user)
	return result.Error
}

// Delete elimina un usuario por su ID
func (r *userRepository) Delete(id uint) error {
	result := r.db.Delete(&domain.User{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// ExistsByUsername verifica si existe un usuario con ese username
func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	result := r.db.Model(&domain.User{}).Where("username = ?", username).Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

// ExistsByEmail verifica si existe un usuario con ese email
func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	result := r.db.Model(&domain.User{}).Where("email = ?", email).Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}
