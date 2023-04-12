package structs

type PagedRequest struct {
	Page          int    `json:"page"`
	PageSize      int    `json:"pageSize"`
	OrderBy       string `json:"orderBy"`
	SortDirection string `json:"sortDirection"`
}
