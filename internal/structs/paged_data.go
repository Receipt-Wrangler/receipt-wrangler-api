package structs

type PagedData struct {
	Data       []any `json:"data"`
	TotalCount int64 `json:"totalCount"`
}
