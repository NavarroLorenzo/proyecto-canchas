package controllers

import (
	"canchas-api/internal/dto"
	"canchas-api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CanchaController struct {
	service services.CanchaService
}

func NewCanchaController(service services.CanchaService) *CanchaController {
	return &CanchaController{service: service}
}

// Create maneja la creación de una cancha (SOLO ADMIN)
// POST /canchas
func (ctrl *CanchaController) Create(c *gin.Context) {
	var req dto.CreateCanchaRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	cancha, err := ctrl.service.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to create cancha",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, cancha)
}

// GetByID obtiene una cancha por su ID
// GET /canchas/:id
func (ctrl *CanchaController) GetByID(c *gin.Context) {
	id := c.Param("id")

	cancha, err := ctrl.service.GetByID(id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "cancha not found" || err.Error() == "invalid ID format" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to get cancha",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, cancha)
}

// GetAll obtiene todas las canchas
// GET /canchas
func (ctrl *CanchaController) GetAll(c *gin.Context) {
	canchas, err := ctrl.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get canchas",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, canchas)
}

// Update actualiza una cancha existente (SOLO ADMIN)
// PUT /canchas/:id
func (ctrl *CanchaController) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateCanchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	cancha, err := ctrl.service.Update(id, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "cancha not found" || err.Error() == "invalid ID format" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to update cancha",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, cancha)
}

// Delete elimina una cancha (SOLO ADMIN)
// DELETE /canchas/:id
func (ctrl *CanchaController) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := ctrl.service.Delete(id); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "cancha not found" || err.Error() == "invalid ID format" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to delete cancha",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cancha deleted successfully",
	})
}

// ❌ ELIMINAR método GetByOwnerID
