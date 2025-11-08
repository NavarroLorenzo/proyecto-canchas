package controllers

import (
	"canchas-api/internal/dto"
	"canchas-api/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CanchaController struct {
	service services.CanchaService
}

// NewCanchaController crea una nueva instancia del controlador
func NewCanchaController(service services.CanchaService) *CanchaController {
	return &CanchaController{service: service}
}

// Create maneja la creación de una cancha
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
		statusCode := http.StatusInternalServerError
		if err.Error() == "invalid owner: user does not exist" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.ErrorResponse{
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

// Update actualiza una cancha existente
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

// Delete elimina una cancha
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

// GetByOwnerID obtiene todas las canchas de un owner
// GET /canchas/owner/:owner_id
func (ctrl *CanchaController) GetByOwnerID(c *gin.Context) {
	ownerIDStr := c.Param("owner_id")
	ownerID, err := strconv.ParseUint(ownerIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid owner ID",
			Message: "Owner ID must be a valid number",
		})
		return
	}

	canchas, err := ctrl.service.GetByOwnerID(uint(ownerID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get canchas",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, canchas)
}
