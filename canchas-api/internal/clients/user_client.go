package clients

import (
	"canchas-api/config"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type UserClient interface {
	ValidateUser(userID uint) (bool, error)
	GetUser(userID uint) (*UserResponse, error)
}

type userClient struct {
	httpClient *http.Client
	baseURL    string
}

type UserResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// NewUserClient crea una nueva instancia del cliente HTTP para users-api
func NewUserClient() UserClient {
	return &userClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: config.AppConfig.UsersAPIURL,
	}
}

// ValidateUser verifica si un usuario existe
func (c *userClient) ValidateUser(userID uint) (bool, error) {
	_, err := c.GetUser(userID)
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetUser obtiene la información de un usuario por ID
func (c *userClient) GetUser(userID uint) (*UserResponse, error) {
	url := fmt.Sprintf("%s/users/%d", c.baseURL, userID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error calling users-api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("user not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("users-api returned status: %d", resp.StatusCode)
	}

	var user UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &user, nil
}
