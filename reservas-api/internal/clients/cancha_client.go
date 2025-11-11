package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reservas-api/config"
	"time"
)

type CanchaClient interface {
	ValidateCancha(canchaID string) (bool, *CanchaResponse, error)
}

type canchaClient struct {
	httpClient *http.Client
	baseURL    string
}

type CanchaResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Location    string  `json:"location"`
	Address     string  `json:"address"`
	Price       float64 `json:"price"`
	Capacity    int     `json:"capacity"`
	Available   bool    `json:"available"`
	ImageURL    string  `json:"image_url"`
}

// NewCanchaClient crea una nueva instancia del cliente HTTP para canchas-api
func NewCanchaClient() CanchaClient {
	return &canchaClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: config.AppConfig.CanchasAPIURL,
	}
}

// ValidateCancha verifica si una cancha existe y retorna sus datos
func (c *canchaClient) ValidateCancha(canchaID string) (bool, *CanchaResponse, error) {
	url := fmt.Sprintf("%s/canchas/%s", c.baseURL, canchaID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return false, nil, fmt.Errorf("error calling canchas-api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil, errors.New("cancha not found")
	}

	if resp.StatusCode != http.StatusOK {
		return false, nil, fmt.Errorf("canchas-api returned status: %d", resp.StatusCode)
	}

	var cancha CanchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&cancha); err != nil {
		return false, nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Verificar que la cancha esté disponible
	if !cancha.Available {
		return false, nil, errors.New("cancha not available")
	}

	return true, &cancha, nil
}
