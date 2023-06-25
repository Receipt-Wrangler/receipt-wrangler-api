package structs

import "database/sql/driver"

type PagedRequest struct {
	Page          int                `json:"page"`
	PageSize      int                `json:"pageSize"`
	OrderBy       string             `json:"orderBy"`
	SortDirection string             `json:"sortDirection"`
	Filter        PagedRequestFilter `json:"filter"`
}

type PagedRequestFilter struct {
	Date         PagedRequestField `json:"date"`
	Amount       PagedRequestField `json:"amount"`
	Name         PagedRequestField `json:"name"`
	PaidBy       PagedRequestField `json:"paidBy"`
	Categories   PagedRequestField `json:"categories"`
	Tags         PagedRequestField `json:"Tags"`
	Status       PagedRequestField `json:"status"`
	ResolvedDate PagedRequestField `json:"ResolvedDate"`
}

type PagedRequestField struct {
	Operation FilterOperation `json:"operation"`
	Value     interface{}
}

type FilterOperation string

const (
	CONTAINS     FilterOperation = "CONTAINS"
	EQUALS       FilterOperation = "EQUALS"
	GREATER_THAN FilterOperation = "GREATER_THAN"
	LESS_THAN    FilterOperation = "LESS_THAN"
)

func (self *FilterOperation) Scan(value string) error {
	*self = FilterOperation(value)
	return nil
}

func (self FilterOperation) Value() (driver.Value, error) {
	return string(self), nil
}
