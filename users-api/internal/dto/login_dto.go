package dto

// LoginRequest - DTO para login
type LoginRequest struct {
	Login    string `json:"login" binding:"required"` // puede ser username o email
	Password string `json:"password" binding:"required"`
}

// LoginResponse - DTO para respuesta de login
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
