package dto

type SearchRequest struct {
	Query    string `form:"q" json:"q"`
	Id       string `form:"id" json:"id"`
	Number   int    `form:"number" json:"number"`
	Type     string `form:"type" json:"type"`
	Available string `form:"available" json:"available"`
	SortBy   string `form:"sort_by" json:"sort_by"`
	SortOrder string `form:"sort_order" json:"sort_order"`
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
