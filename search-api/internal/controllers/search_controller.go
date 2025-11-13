package controllers

import (
	"fmt"
	"net/http"
	"search-api/internal/dto"
	"search-api/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
)

type SearchController struct {
	service services.SearchService
}

func NewSearchController(service services.SearchService) *SearchController {
	return &SearchController{service: service}
}

func (ctrl *SearchController) Search(c *gin.Context) {
	var req dto.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid query parameters",
			Message: err.Error(),
		})
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// Construir query para Solr. Priorizamos búsqueda por número, luego por id, luego por query.
	var baseQ string
	if req.Number > 0 {
		baseQ = fmt.Sprintf("number:%d", req.Number)
	} else if req.Id != "" {
		baseQ = fmt.Sprintf("id:%s", req.Id)
	} else if req.Query != "" {
		// Para búsqueda de texto, usar la query directamente (edismax la manejará)
		// edismax buscará en los campos especificados en qf (name, description, location, address, number)
		baseQ = req.Query
	} else {
		baseQ = "*:*"
	}

	// Preparar filtros (fq) para Solr - estos son más eficientes que incluirlos en la query principal
	var fqFilters []string
	if typeFilter := strings.ToLower(strings.TrimSpace(req.Type)); typeFilter != "" {
		// type es un campo text_general; buscar el término exacto (tokenizado)
		fqFilters = append(fqFilters, fmt.Sprintf("type:%s", typeFilter))
	}
	if req.Available != "" {
		// available en Solr es booleano
		if strings.ToLower(req.Available) == "true" {
			fqFilters = append(fqFilters, "available:true")
		} else if strings.ToLower(req.Available) == "false" {
			fqFilters = append(fqFilters, "available:false")
		}
	}

	sortField := strings.ToLower(req.SortBy)
	sortDir := strings.ToLower(req.SortOrder)
	if sortDir != "desc" {
		sortDir = "asc"
	}

	sortParam := ""
	switch sortField {
	case "price", "capacity":
		sortParam = fmt.Sprintf("%s %s", sortField, sortDir)
	case "name":
		sortParam = fmt.Sprintf("name_sort %s", sortDir)
	}

	resp, err := ctrl.service.Search(baseQ, fqFilters, req.Page, req.PageSize, sortParam)
	if err != nil {
		// Si hay cualquier error con Solr (conexión, configuración, o errores de Solr),
		// devolver un resultado vacío en lugar de error 500 para evitar errores en el frontend
		if strings.Contains(err.Error(), "failed to query Solr") || 
		   strings.Contains(err.Error(), "SOLR_URL is not configured") ||
		   strings.Contains(err.Error(), "SOLR_CORE is not configured") ||
		   strings.Contains(err.Error(), "Solr error") {
			// Devolver respuesta vacía en lugar de error 500
			c.JSON(http.StatusOK, &dto.SearchResponse{
				Results:    []interface{}{},
				Total:      0,
				Page:       req.Page,
				PageSize:   req.PageSize,
				TotalPages: 0,
			})
			return
		}
		
		// Para otros errores inesperados, devolver 500
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Search failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
