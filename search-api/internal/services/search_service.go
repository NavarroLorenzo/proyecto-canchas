package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type SearchService interface {
	// Indexa una cancha en Solr (usado por el consumer RabbitMQ)
	IndexCancha(data interface{}) error

	// Ejecuta una búsqueda en Solr (usado por el endpoint /search)
	Search(q string, page, pageSize int) (map[string]interface{}, error)
}

type searchService struct {
	solrURL  string
	coreName string
}

// Constructor del servicio
func NewSearchService() SearchService {
	return &searchService{
		solrURL:  os.Getenv("SOLR_URL"),
		coreName: os.Getenv("SOLR_CORE"),
	}
}

// ✅ Indexa una cancha en Solr (cuando llega un evento desde RabbitMQ)
func (s *searchService) IndexCancha(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal cancha data: %v", err)
	}

	url := fmt.Sprintf("%s/%s/update?commit=true", s.solrURL, s.coreName)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send data to Solr: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Solr returned %d: %s", resp.StatusCode, string(body))
	}

	fmt.Println("[Solr] Cancha indexada correctamente en Solr ✅")
	return nil
}

// ✅ Permite buscar canchas en Solr desde el endpoint /search
func (s *searchService) Search(q string, page, pageSize int) (map[string]interface{}, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	url := fmt.Sprintf("%s/%s/select?q=%s&start=%d&rows=%d&wt=json",
		s.solrURL, s.coreName, q, start, pageSize)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to query Solr: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Solr error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Solr response: %v", err)
	}

	return result, nil
}
