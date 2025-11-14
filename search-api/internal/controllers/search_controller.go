package controllers

import (
	"fmt"
	"net/http"
	"search-api/internal/dto"
	"search-api/internal/services"
	"search-api/internal/utils"
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
		baseQ = req.Query
	} else {
		baseQ = "*:*"
	}

	// Añadir filtros adicionales (type, available) como condiciones AND
	var fqParts []string
	if req.Type != "" {
		// Normalizar tipo (minusculas y sin acentos) para que coincida con el indexado
		t := utils.NormalizeString(req.Type)
		fqParts = append(fqParts, fmt.Sprintf("type:%s", t))
	}
	if req.Available != "" {
		// available en Solr es booleano
		if strings.ToLower(req.Available) == "true" {
			fqParts = append(fqParts, "available:true")
		} else if strings.ToLower(req.Available) == "false" {
			fqParts = append(fqParts, "available:false")
		}
	}

	finalQ := baseQ
	if len(fqParts) > 0 {
		finalQ = fmt.Sprintf("(%s) AND %s", baseQ, strings.Join(fqParts, " AND "))
	}

	// Construir string de ordenamiento para Solr
	var sortStr string
	if req.SortBy != "" {
		// Mapear campos del frontend a campos de Solr
		solrField := req.SortBy
		switch req.SortBy {
		case "name":
			// Usar name_sort que es el campo optimizado para ordenamiento
			solrField = "name_sort"
		case "price":
			solrField = "price"
		case "capacity":
			solrField = "capacity"
		}

		// Determinar dirección de ordenamiento
		order := "asc"
		if req.SortOrder != "" {
			order = strings.ToLower(req.SortOrder)
		}
		if order != "asc" && order != "desc" {
			order = "asc"
		}

		sortStr = fmt.Sprintf("%s %s", solrField, order)
	}

	resp, err := ctrl.service.Search(finalQ, fqParts, req.Page, req.PageSize, sortStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Search failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
