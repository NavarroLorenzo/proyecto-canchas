package dto

type SearchRequest struct {
	Query    string `form:"q" json:"q"`
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
}

type SearchResponse struct {
	Results    interface{} `json:"results"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}
