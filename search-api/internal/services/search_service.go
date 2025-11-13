package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"

	"search-api/internal/dto"
)

type SearchService interface {
	IndexCancha(data interface{}) error
	DeleteCancha(id string) error
	Search(q string, page, pageSize int, sort string) (*dto.SearchResponse, error)
}

type searchService struct {
	solrURL  string
	coreName string
}

func NewSearchService() SearchService {
	return &searchService{
		solrURL:  os.Getenv("SOLR_URL"),
		coreName: os.Getenv("SOLR_CORE"),
	}
}

func (s *searchService) IndexCancha(data interface{}) error {
	doc, ok := data.(map[string]interface{})
	if !ok {
		raw, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal cancha data: %v", err)
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return fmt.Errorf("failed to normalize cancha data: %v", err)
		}
	}

	payload := map[string]interface{}{
		"add": map[string]interface{}{
			"doc": doc,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Solr payload: %v", err)
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

	fmt.Println("[Solr] Cancha indexada correctamente en Solr.")
	return nil
}

func (s *searchService) DeleteCancha(id string) error {
	payload := map[string]interface{}{
		"delete": map[string]string{
			"query": fmt.Sprintf("id:%s", id),
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal delete payload: %v", err)
	}

	url := fmt.Sprintf("%s/%s/update?commit=true", s.solrURL, s.coreName)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send delete request to Solr: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Solr delete returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *searchService) Search(q string, page, pageSize int, sort string) (*dto.SearchResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	params := url.Values{}
	params.Set("q", q)
	params.Set("start", fmt.Sprintf("%d", start))
	params.Set("rows", fmt.Sprintf("%d", pageSize))
	params.Set("wt", "json")
	params.Set("defType", "edismax")
	params.Set("qf", "name description location address number")
	if sort != "" {
		params.Set("sort", sort)
	}

	endpoint := fmt.Sprintf("%s/%s/select?%s", s.solrURL, s.coreName, params.Encode())

	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to query Solr: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Solr error %d: %s", resp.StatusCode, string(body))
	}

	var solrResp struct {
		Response struct {
			NumFound int                      `json:"numFound"`
			Start    int                      `json:"start"`
			Docs     []map[string]interface{} `json:"docs"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&solrResp); err != nil {
		return nil, fmt.Errorf("failed to decode Solr response: %v", err)
	}

	totalPages := int(math.Ceil(float64(solrResp.Response.NumFound) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &dto.SearchResponse{
		Results:    solrResp.Response.Docs,
		Total:      solrResp.Response.NumFound,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
