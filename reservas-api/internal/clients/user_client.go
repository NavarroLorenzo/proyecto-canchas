package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reservas-api/config"
	"time"
)

type UserClient interface {
	ValidateUser(userID uint, token string) (bool, *UserResponse, error)
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

// ValidateUser verifica si un usuario existe y retorna sus datos
func (c *userClient) ValidateUser(userID uint, token string) (bool, *UserResponse, error) {
	url := fmt.Sprintf("%s/users/%d", c.baseURL, userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, nil, fmt.Errorf("error creating request: %w", err)
	}

	// 🟢 Agregar el header Authorization con el token Bearer
	if token != "" {
		req.Header.Add("Authorization", token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, nil, fmt.Errorf("error calling users-api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil, errors.New("user not found")
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return false, nil, errors.New("unauthorized: invalid or missing token")
	}

	if resp.StatusCode != http.StatusOK {
		return false, nil, fmt.Errorf("users-api returned status: %d", resp.StatusCode)
	}

	var user UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return false, nil, fmt.Errorf("error decoding response: %w", err)
	}

	return true, &user, nil
}
