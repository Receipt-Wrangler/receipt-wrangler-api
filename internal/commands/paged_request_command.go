package commands

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type SortDirection string

const (
	ASCENDING  SortDirection = "asc"
	DESCENDING SortDirection = "desc"
	DEFAULT    SortDirection = ""
)

func GetValidSortDirections() []any {
	return []any{ASCENDING, DESCENDING, DEFAULT}
}

func (sortDirection *SortDirection) Scan(value string) error {
	*sortDirection = SortDirection(value)
	return nil
}

func (sortDirection SortDirection) Value() (driver.Value, error) {
	if sortDirection != ASCENDING && sortDirection != DESCENDING && sortDirection != DEFAULT {
		return nil, errors.New("invalid sortDirection")
	}
	return string(sortDirection), nil
}

type PagedRequestCommand struct {
	Page          int           `json:"page"`
	PageSize      int           `json:"pageSize"`
	OrderBy       string        `json:"orderBy"`
	SortDirection SortDirection `json:"sortDirection"`
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

func (command *PagedRequestCommand) Validate() structs.ValidatorError {
	vErr := structs.ValidatorError{}
	errorMap := make(map[string]string)

	if command.Page < 1 {
		errorMap["page"] = "Page must be greater than or equal to 0"
	}

	if command.PageSize < 1 && command.PageSize != -1 {
		errorMap["pageSize"] = "PageSize must be greater than or equal to 1, or -1 for no limit"
	}

	if command.SortDirection != ASCENDING && command.SortDirection != DESCENDING && command.SortDirection != DEFAULT {
		errorMap["sortDirection"] = "Invalid sort direction"
	}

	vErr.Errors = errorMap
	return vErr
}

type ReceiptPagedRequestCommand struct {
	PagedRequestCommand
	Filter       ReceiptPagedRequestFilter `json:"filter"`
	FullReceipts bool                      `json:"fullReceipts"`
}

func (command *ReceiptPagedRequestCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &command)
	if err != nil {
		return err
	}

	if command.Filter.Amount.Value == nil || command.Filter.Amount.Value == "" {
		command.Filter.Amount.Value = float64(0)
	}

	if command.Filter.PaidBy.Value == nil || command.Filter.PaidBy.Value == "" {
		command.Filter.PaidBy.Value = make([]interface{}, 0)
	}

	if command.Filter.Categories.Value == nil || command.Filter.Categories.Value == "" {
		command.Filter.Categories.Value = make([]interface{}, 0)
	}

	if command.Filter.Tags.Value == nil || command.Filter.Tags.Value == "" {
		command.Filter.Tags.Value = make([]interface{}, 0)
	}

	if command.Filter.Status.Value == nil || command.Filter.Status.Value == "" {
		command.Filter.Status.Value = make([]interface{}, 0)
	}

	if command.Filter.CreatedAt.Value == nil {
		command.Filter.CreatedAt.Value = ""
	}

	if command.Filter.Date.Value == nil {
		command.Filter.Date.Value = ""
	}

	if command.Filter.ResolvedDate.Value == nil {
		command.Filter.ResolvedDate.Value = ""
	}

	return nil
}

type ReceiptPagedRequestFilter struct {
	Date         PagedRequestField `json:"date"`
	Amount       PagedRequestField `json:"amount"`
	Name         PagedRequestField `json:"name"`
	PaidBy       PagedRequestField `json:"paidBy"`
	Categories   PagedRequestField `json:"categories"`
	Tags         PagedRequestField `json:"Tags"`
	Status       PagedRequestField `json:"status"`
	ResolvedDate PagedRequestField `json:"resolvedDate"`
	CreatedAt    PagedRequestField `json:"createdAt"`
}

type PagedRequestField struct {
	Operation FilterOperation `json:"operation"`
	Value     interface{}
}

type FilterOperation string

const (
	CONTAINS             FilterOperation = "CONTAINS"
	EQUALS               FilterOperation = "EQUALS"
	GREATER_THAN         FilterOperation = "GREATER_THAN"
	LESS_THAN            FilterOperation = "LESS_THAN"
	BETWEEN              FilterOperation = "BETWEEN"
	WITHIN_CURRENT_MONTH FilterOperation = "WITHIN_CURRENT_MONTH"
)

func (self *FilterOperation) Scan(value string) error {
	*self = FilterOperation(value)
	return nil
}

func (self FilterOperation) Value() (driver.Value, error) {
	return string(self), nil
}
