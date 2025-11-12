package dto

// ErrorResponse - DTO para errores
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
