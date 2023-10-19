package commands

import (
	"mime/multipart"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
)

type QuickScanCommand struct {
	File         multipart.File        `json:"file"`
	FileHeader   *multipart.FileHeader `json:"fileHeader"`
	PaidByUser   models.User           `json:"paidByUser"`
	PaidByUserId uint                  `json:"paidByUserId"`
	Group        models.Group          `json:"-"`
	GroupId      uint                  `json:"groupId"`
	Status       models.ReceiptStatus  `json:"status"`
}

func (command *QuickScanCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	var status models.ReceiptStatus

	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		return err
	}

	groupId, err := simpleutils.StringToUint(r.Form.Get("groupId"))
	if err != nil {
		return err
	}

	paidByUserId, err := simpleutils.StringToUint(r.Form.Get("paidByUserId"))
	if err != nil {
		return err
	}

	err = status.Scan(r.Form.Get("status"))
	if err != nil {
		return err
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		return err
	}
	defer file.Close()

	command.File = file
	command.FileHeader = fileHeader
	command.GroupId = groupId
	command.Status = status
	command.PaidByUserId = paidByUserId

	return nil
}

func (command QuickScanCommand) Validate() structs.ValidatorError {
	vErr := structs.ValidatorError{
		Errors: make(map[string]string),
	}

	if command.GroupId == 0 {
		vErr.Errors["groupId"] = "Group Id is required."
	}

	if len(command.Status) == 0 {
		vErr.Errors["status"] = "Status is required."
	}

	if command.FileHeader == nil {
		vErr.Errors["file"] = "File is required."
	}

	return vErr
}

func (command *QuickScanCommand) LoadDataFromRequestAndValidate(w http.ResponseWriter, r *http.Request) (structs.ValidatorError, error) {
	err := command.LoadDataFromRequest(w, r)
	if err != nil {
		return structs.ValidatorError{}, err
	}

	vErr := command.Validate()
	if len(vErr.Errors) > 0 {
		return vErr, nil
	}

	return structs.ValidatorError{}, nil
}
