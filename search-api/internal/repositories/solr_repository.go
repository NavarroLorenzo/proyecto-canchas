package repositories

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"search-api/internal/domain"
)

type SolrRepository interface {
	Search(query string, filters map[string]string, sort string, start, rows int) ([]domain.CanchaSearch, int64, error)
	AddOrUpdate(cancha domain.CanchaSearch) error
	Delete(id string) error
}

type solrRepository struct {
	baseURL string
	core    string
}

func NewSolrRepository(baseURL, core string) SolrRepository {
	return &solrRepository{
		baseURL: baseURL,
		core:    core,
	}
}

// 🔹 Ejecutar búsqueda en Solr
func (r *solrRepository) Search(query string, filters map[string]string, sort string, start, rows int) ([]domain.CanchaSearch, int64, error) {
	url := fmt.Sprintf("%s/%s/select?q=%s&start=%d&rows=%d&sort=%s&wt=json",
		r.baseURL, r.core, query, start, rows, sort)

	for k, v := range filters {
		url += fmt.Sprintf("&fq=%s:%s", k, v)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result struct {
		Response struct {
			Docs []domain.CanchaSearch `json:"docs"`
			Num  int64                 `json:"numFound"`
		} `json:"response"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 0, err
	}

	return result.Response.Docs, result.Response.Num, nil
}

// 🔹 Agregar o actualizar una cancha
func (r *solrRepository) AddOrUpdate(cancha domain.CanchaSearch) error {
	data := []domain.CanchaSearch{cancha}
	jsonData, _ := json.Marshal(data)

	url := fmt.Sprintf("%s/%s/update?commit=true", r.baseURL, r.core)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("error al indexar cancha: %s", string(body))
	}
	return nil
}

// 🔹 Borrar cancha por ID
func (r *solrRepository) Delete(id string) error {
	deleteCmd := fmt.Sprintf(`{"delete":{"id":"%s"}}`, id)
	url := fmt.Sprintf("%s/%s/update?commit=true", r.baseURL, r.core)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(deleteCmd)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("error al borrar cancha: %s", string(body))
	}
	return nil
}
