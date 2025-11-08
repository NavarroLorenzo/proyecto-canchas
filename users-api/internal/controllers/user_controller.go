package controllers

import (
	"net/http"
	"strconv"
	"users-api/internal/dto"
	"users-api/internal/services"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service services.UserService
}

// NewUserController crea una nueva instancia del controlador
func NewUserController(service services.UserService) *UserController {
	return &UserController{service: service}
}

// Register maneja el registro de usuarios normales
// POST /users/register
func (ctrl *UserController) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	user, err := ctrl.service.Register(&req, "normal")
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Registration failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// RegisterAdmin maneja el registro de usuarios administradores
// POST /users/admin
func (ctrl *UserController) RegisterAdmin(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	user, err := ctrl.service.Register(&req, "admin")
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Registration failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login maneja el login de usuarios
// POST /users/login
func (ctrl *UserController) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	response, err := ctrl.service.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Login failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetByID obtiene un usuario por su ID
// GET /users/:id
func (ctrl *UserController) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid number",
		})
		return
	}

	user, err := ctrl.service.GetByID(uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to get user",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetAll obtiene todos los usuarios
// GET /users
func (ctrl *UserController) GetAll(c *gin.Context) {
	users, err := ctrl.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get users",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

// Update actualiza un usuario
// PUT /users/:id
func (ctrl *UserController) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid number",
		})
		return
	}

	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	user, err := ctrl.service.Update(uint(id), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to update user",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Delete elimina un usuario
// DELETE /users/:id
func (ctrl *UserController) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid number",
		})
		return
	}

	if err := ctrl.service.Delete(uint(id)); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Failed to delete user",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}
