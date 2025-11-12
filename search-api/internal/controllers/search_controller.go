package controllers

import (
	"net/http"
	"search-api/internal/dto"
	"search-api/internal/services"

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

	resp, err := ctrl.service.Search(req.Query, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Search failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
