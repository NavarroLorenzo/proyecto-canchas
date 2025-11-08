package domain

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email     string    `gorm:"uniqueIndex;not null;size:100" json:"email"`
	Password  string    `gorm:"not null;size:255" json:"-"` // El "-" evita que se serialice en JSON
	FirstName string    `gorm:"size:50" json:"first_name"`
	LastName  string    `gorm:"size:50" json:"last_name"`
	Role      string    `gorm:"not null;size:20;default:'normal'" json:"role"` // "normal" o "admin"
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName especifica el nombre de la tabla
func (User) TableName() string {
	return "users"
}
