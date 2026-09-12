package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

// SystemTaskPagedRequestFilter mirrors ReceiptPagedRequestFilter for the system
// task table. Every field is the same {operation, value} pair, so the desktop
// reuses the receipt filter's row UI and the FilterOperation enum verbatim.
//
// Unlike the receipt filter there is no init*Values pass seeding non-nil
// defaults: SystemTaskRepository.buildSystemTaskFilterQuery uses comma-ok type
// assertions throughout and treats a wrong-typed value as "field not set",
// so a malformed body is a query no-op rather than a panic.
type SystemTaskPagedRequestFilter struct {
	Type      PagedRequestField `json:"type"`
	RanBy     PagedRequestField `json:"ranBy"`
	StartedAt PagedRequestField `json:"startedAt"`
	EndedAt   PagedRequestField `json:"endedAt"`
}

type GetSystemTaskCommand struct {
	PagedRequestCommand
	AssociatedEntityId   uint                         `json:"associatedEntityId"`
	AssociatedEntityType models.AssociatedEntityType  `json:"associatedEntityType"`
	Filter               SystemTaskPagedRequestFilter `json:"filter"`
}

func (command *GetSystemTaskCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &command)
	if err != nil {
		return err
	}

	return nil
}

func (command *GetSystemTaskCommand) Validate() structs.ValidatorError {
	vErrs := command.PagedRequestCommand.Validate()

	return vErrs
}
