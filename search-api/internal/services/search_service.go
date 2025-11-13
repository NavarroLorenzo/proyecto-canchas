package services

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"

	"search-api/config"
	"search-api/internal/cache"
	"search-api/internal/dto"
)

type SearchService interface {
	IndexCancha(data interface{}) error
	DeleteCancha(id string) error
	Search(q string, fqFilters []string, page, pageSize int, sort string) (*dto.SearchResponse, error)
}

type searchService struct {
	solrURL  string
	coreName string
	cache    *cache.Manager
}

func NewSearchService(cacheManager *cache.Manager) SearchService {
	solrURL := config.AppConfig.SolrURL
	coreName := config.AppConfig.SolrCore
	
	if solrURL == "" {
		solrURL = "http://localhost:8983/solr"
		log.Println("[SearchService] SOLR_URL not set, using default:", solrURL)
	}
	if coreName == "" {
		coreName = "canchas"
		log.Println("[SearchService] SOLR_CORE not set, using default:", coreName)
	}
	
	return &searchService{
		solrURL:  solrURL,
		coreName: coreName,
		cache:    cacheManager,
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

	if name, ok := doc["name"].(string); ok && name != "" {
		doc["name_sort"] = strings.ToLower(name)
	}
	
	// Normalizar el campo type a minúsculas para búsquedas consistentes
	if typeVal, ok := doc["type"].(string); ok && typeVal != "" {
		doc["type"] = strings.ToLower(strings.TrimSpace(typeVal))
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

	log.Println("[Solr] Cancha indexada correctamente en Solr.")
	if s.cache != nil {
		s.cache.InvalidateAll()
	}
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

	if s.cache != nil {
		s.cache.InvalidateAll()
	}
	return nil
}

func (s *searchService) Search(q string, fqFilters []string, page, pageSize int, sort string) (*dto.SearchResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	params := url.Values{}
	
	// Para búsquedas de texto, usar wildcards para permitir coincidencias parciales
	searchQuery := q
	if q != "*:*" && !strings.Contains(q, ":") {
		// Si es una búsqueda de texto libre (no es una query específica de campo),
		// agregar wildcards para permitir coincidencias parciales
		searchQuery = fmt.Sprintf("*%s*", q)
	}
	
	params.Set("q", searchQuery)
	params.Set("start", fmt.Sprintf("%d", start))
	params.Set("rows", fmt.Sprintf("%d", pageSize))
	params.Set("wt", "json")
	params.Set("defType", "edismax")
	params.Set("qf", "name description location address number")
	params.Set("mm", "1") // Minimum match: al menos 1 término debe coincidir
	
	// Agregar filtros fq (filter queries) - más eficientes que incluirlos en la query principal
	for _, fq := range fqFilters {
		params.Add("fq", fq)
	}
	
	if sort != "" {
		params.Set("sort", sort)
	}

	cacheKey := buildCacheKey(params.Encode())
	if s.cache != nil {
		if cached, ok := s.cache.GetSearchResponse(cacheKey); ok {
			return cached, nil
		}
	}

	if s.solrURL == "" {
		return nil, fmt.Errorf("SOLR_URL is not configured")
	}
	if s.coreName == "" {
		return nil, fmt.Errorf("SOLR_CORE is not configured")
	}

	endpoint := fmt.Sprintf("%s/%s/select?%s", s.solrURL, s.coreName, params.Encode())

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Printf("[SearchService] Error connecting to Solr at %s: %v", s.solrURL, err)
		return nil, fmt.Errorf("failed to query Solr at %s: %v", s.solrURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		log.Printf("[SearchService] Solr returned error %d: %s", resp.StatusCode, bodyStr)
		
		// Si el error es por ordenamiento (especialmente name_sort), intentar sin ordenamiento
		if strings.Contains(bodyStr, "can not sort") && sort != "" {
			log.Printf("[SearchService] Retrying without sort parameter due to sort error")
			// Reintentar sin el parámetro de ordenamiento
			params.Del("sort")
			endpoint = fmt.Sprintf("%s/%s/select?%s", s.solrURL, s.coreName, params.Encode())
			
			retryResp, retryErr := http.Get(endpoint)
			if retryErr != nil {
				return nil, fmt.Errorf("failed to query Solr at %s: %v", s.solrURL, retryErr)
			}
			defer retryResp.Body.Close()
			
			if retryResp.StatusCode == http.StatusOK {
				resp = retryResp
			} else {
				return nil, fmt.Errorf("Solr error %d: %s", resp.StatusCode, bodyStr)
			}
		} else {
			return nil, fmt.Errorf("Solr error %d: %s", resp.StatusCode, bodyStr)
		}
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

	result := &dto.SearchResponse{
		Results:    solrResp.Response.Docs,
		Total:      solrResp.Response.NumFound,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	if s.cache != nil {
		s.cache.SaveSearchResponse(cacheKey, result)
	}

	return result, nil
}

func buildCacheKey(signature string) string {
	hash := md5.Sum([]byte(signature))
	return hex.EncodeToString(hash[:])
}
