package clients

import (
	"canchas-api/config"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ReservaClient interface {
	DeleteByCanchaID(canchaID string) error
}

type reservaClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewReservaClient() ReservaClient {
	return &reservaClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    config.AppConfig.ReservasAPIURL,
	}
}

func (c *reservaClient) DeleteByCanchaID(canchaID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/reservas/cancha/%s", c.baseURL, canchaID), nil)
	if err != nil {
		return fmt.Errorf("failed to build request to reservas-api: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call reservas-api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("reservas-api returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
