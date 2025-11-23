package services

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"search-api/config"
	"search-api/internal/cache"
	"search-api/internal/dto"
	"search-api/internal/repositories"
	"search-api/internal/utils"
)

type SearchService interface {
	IndexCancha(data interface{}) error
	DeleteCancha(id string) error
	Search(q string, fqFilters []string, page, pageSize int, sort string) (*dto.SearchResponse, error)
	ReindexAllCanchas() error
}

type searchService struct {
	cache         *cache.Manager
	canchasAPIURL string
	solrRepo      repositories.SolrRepository
	httpClient    *http.Client
}

func NewSearchService(cacheManager *cache.Manager, solrRepo repositories.SolrRepository, canchasAPIURL string) SearchService {
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

	if solrRepo == nil {
		solrRepo = repositories.NewSolrRepository(solrURL, coreName, nil)
	}

	return &searchService{
		cache:         cacheManager,
		canchasAPIURL: canchasAPIURL,
		solrRepo:      solrRepo,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
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

	// Normalizar el campo type para búsquedas consistentes (minusculas + sin acentos)
	if typeVal, ok := doc["type"].(string); ok && typeVal != "" {
		doc["type"] = utils.NormalizeString(typeVal)
	}

	if err := s.solrRepo.Add(doc); err != nil {
		return fmt.Errorf("failed to send data to Solr: %v", err)
	}

	log.Println("[Solr] Cancha indexada correctamente en Solr.")
	if s.cache != nil {
		s.cache.InvalidateAll()
	}
	return nil
}

func (s *searchService) DeleteCancha(id string) error {
	if err := s.solrRepo.DeleteByQuery(fmt.Sprintf("id:%s", id)); err != nil {
		return fmt.Errorf("failed to send delete request to Solr: %v", err)
	}

	if s.cache != nil {
		s.cache.InvalidateAll()
	}
	return nil
}

// ReindexAllCanchas vuelve a poblar Solr leyendo todas las canchas desde canchas-api.
func (s *searchService) ReindexAllCanchas() error {
	if s.canchasAPIURL == "" {
		return errors.New("CANCHAS_API_URL is not configured")
	}

	apiURL := strings.TrimRight(s.canchasAPIURL, "/")
	reqURL := fmt.Sprintf("%s/canchas", apiURL)

	log.Printf("[Reindex] Fetching canchas from %s", reqURL)
	resp, err := s.httpClient.Get(reqURL)
	if err != nil {
		return fmt.Errorf("failed to fetch canchas for reindex: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("canchas API returned %d: %s", resp.StatusCode, string(body))
	}

	var payload struct {
		Canchas []map[string]interface{} `json:"canchas"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("failed to decode canchas response: %w", err)
	}

	log.Printf("[Reindex] Clearing Solr core before reindexing")
	if err := s.solrRepo.ClearAll(); err != nil {
		return fmt.Errorf("failed to clear Solr index: %w", err)
	}

	var failed int
	for _, cancha := range payload.Canchas {
		if err := s.IndexCancha(cancha); err != nil {
			log.Printf("[Reindex] Failed to index cancha %+v: %v", cancha["id"], err)
			failed++
		}
	}

	if failed > 0 {
		return fmt.Errorf("reindex completed with %d failures", failed)
	}

	log.Printf("[Reindex] Reindex completed. Total canchas indexed: %d", len(payload.Canchas))
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
		// Si es texto libre (no es query por campo), agregar wildcards para coincidencias parciales
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

	docs, numFound, err := s.solrRepo.Search(params.Encode())
	if err != nil {
		// Si el error es por ordenamiento (especialmente name_sort), intentar sin ordenamiento
		if strings.Contains(err.Error(), "can not sort") && sort != "" {
			log.Printf("[SearchService] Retrying without sort parameter due to sort error")
			params.Del("sort")
			docs, numFound, err = s.solrRepo.Search(params.Encode())
		}
		if err != nil {
			return nil, err
		}
	}

	totalPages := int(math.Ceil(float64(numFound) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	result := &dto.SearchResponse{
		Results:    docs,
		Total:      numFound,
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
