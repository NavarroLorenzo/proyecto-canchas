package controllers

import (
	"net/http"
	"reservas-api/internal/dto"
	"reservas-api/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ReservaController struct {
	service services.ReservaService
}

// NewReservaController crea una nueva instancia del controlador
func NewReservaController(service services.ReservaService) *ReservaController {
	return &ReservaController{service: service}
}

// Create maneja la creación de una reserva
// POST /reservas
func (ctrl *ReservaController) Create(c *gin.Context) {
	var req dto.CreateReservaRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	token := c.GetHeader("Authorization")
	reserva, err := ctrl.service.Create(&req, token)

	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user validation failed" ||
			err.Error() == "cancha validation failed" ||
			err.Error() == "cancha not available for the selected time slot" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to create reserva",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, reserva)
}

// GetByID obtiene una reserva por su ID
// GET /reservas/:id
func (ctrl *ReservaController) GetByID(c *gin.Context) {
	id := c.Param("id")

	reserva, err := ctrl.service.GetByID(id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "reserva not found" || err.Error() == "invalid ID format" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to get reserva",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, reserva)
}

// GetAll obtiene todas las reservas
// GET /reservas
func (ctrl *ReservaController) GetAll(c *gin.Context) {
	reservas, err := ctrl.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get reservas",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, reservas)
}

// GetByUserID obtiene todas las reservas de un usuario
// GET /reservas/user/:user_id
func (ctrl *ReservaController) GetByUserID(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user ID",
			Message: "User ID must be a valid number",
		})
		return
	}

	reservas, err := ctrl.service.GetByUserID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get reservas",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, reservas)
}

// GetByCanchaID obtiene todas las reservas de una cancha
// GET /reservas/cancha/:cancha_id
func (ctrl *ReservaController) GetByCanchaID(c *gin.Context) {
	canchaID := c.Param("cancha_id")

	reservas, err := ctrl.service.GetByCanchaID(canchaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get reservas",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, reservas)
}

// Update actualiza una reserva existente
// PUT /reservas/:id
func (ctrl *ReservaController) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateReservaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	reserva, err := ctrl.service.Update(id, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "reserva not found" || err.Error() == "invalid ID format" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "cannot update a cancelled reservation" ||
			err.Error() == "cancha not available for the selected time slot" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to update reserva",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, reserva)
}

// Cancel cancela una reserva
// DELETE /reservas/:id
func (ctrl *ReservaController) Cancel(c *gin.Context) {
	id := c.Param("id")

	if err := ctrl.service.Cancel(id); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "reserva not found" || err.Error() == "invalid ID format" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "reservation already cancelled" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to cancel reserva",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reserva cancelled successfully",
	})
}
