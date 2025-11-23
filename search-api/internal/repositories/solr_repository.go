package repositories

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SolrRepository interface {
	Search(params string) ([]map[string]interface{}, int, error)
	Add(doc map[string]interface{}) error
	DeleteByQuery(query string) error
	ClearAll() error
}

type solrRepository struct {
	baseURL    string
	core       string
	httpClient *http.Client
}

func NewSolrRepository(baseURL, core string, httpClient *http.Client) SolrRepository {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &solrRepository{
		baseURL:    baseURL,
		core:       core,
		httpClient: httpClient,
	}
}

// Search ejecuta una query completa en Solr; params ya debe venir url-encoded.
func (r *solrRepository) Search(params string) ([]map[string]interface{}, int, error) {
	url := fmt.Sprintf("%s/%s/select?%s", r.baseURL, r.core, params)

	resp, err := r.httpClient.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("Solr returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Response struct {
			Docs []map[string]interface{} `json:"docs"`
			Num  int                     `json:"numFound"`
		} `json:"response"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 0, err
	}

	return result.Response.Docs, result.Response.Num, nil
}

// Add agrega o actualiza un documento en el core.
func (r *solrRepository) Add(doc map[string]interface{}) error {
	payload := map[string]interface{}{
		"add": map[string]interface{}{
			"doc": doc,
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s/update?commit=true", r.baseURL, r.core)
	resp, err := r.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error al indexar cancha: %s", string(body))
	}
	return nil
}

// DeleteByQuery elimina documentos que cumplan la query (id:xxx, *:*, etc).
func (r *solrRepository) DeleteByQuery(query string) error {
	payload := map[string]interface{}{
		"delete": map[string]string{
			"query": query,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s/update?commit=true", r.baseURL, r.core)
	resp, err := r.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error al borrar cancha: %s", string(body))
	}
	return nil
}

// ClearAll elimina todos los documentos del core.
func (r *solrRepository) ClearAll() error {
	return r.DeleteByQuery("*:*")
}
