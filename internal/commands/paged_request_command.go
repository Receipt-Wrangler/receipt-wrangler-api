package commands

import (
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/utils"
)

type PagedRequestCommand struct {
	Page          int    `json:"page"`
	PageSize      int    `json:"pageSize"`
	OrderBy       string `json:"orderBy"`
	SortDirection string `json:"sortDirection"`
}

func (command *PagedRequestCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	pagedRequestCommand := PagedRequestCommand{}

	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &pagedRequestCommand)
	if err != nil {
		return err
	}

	command.Page = pagedRequestCommand.Page
	command.PageSize = pagedRequestCommand.PageSize
	command.OrderBy = pagedRequestCommand.OrderBy
	command.SortDirection = pagedRequestCommand.SortDirection

	return nil
}

type ReceiptPagedRequestCommand struct {
	PagedRequestCommand
	Filter ReceiptPagedRequestFilter `json:"filter"`
}

type ReceiptPagedRequestFilter struct {
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
