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
		baseQ = req.Query
	} else {
		baseQ = "*:*"
	}

	// Añadir filtros adicionales (type, available) como condiciones AND
	var fqParts []string
	if req.Type != "" {
		// type es un campo text_general; buscamos el token exacto (Solr normaliza a minúsculas)
		fqParts = append(fqParts, fmt.Sprintf("type:%s", req.Type))
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

	resp, err := ctrl.service.Search(finalQ, fqParts, req.Page, req.PageSize, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Search failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
