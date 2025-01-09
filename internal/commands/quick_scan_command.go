package commands

import (
	"mime/multipart"
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
)

type QuickScanCommand struct {
	Files         []multipart.File        `json:"file"`
	FileHeaders   []*multipart.FileHeader `json:"fileHeader"`
	PaidByUserIds []uint                  `json:"paidByUserId"`
	GroupIds      []uint                  `json:"groupId"`
	Statuses      []models.ReceiptStatus  `json:"status"`
}

func (command *QuickScanCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseMultipartForm(constants.MultipartFormMaxSize)

	var form = r.Form

	var files = make([]multipart.File, 0)
	var fileHeaders = make([]*multipart.FileHeader, 0)
	var paidByUserIds = make([]uint, 0)
	var groupIds = make([]uint, 0)
	var statuses = make([]models.ReceiptStatus, 0)

	var formPaidByUserIds = form["paidByUserIds"]
	var formGroupIds = form["groupIds"]
	var formStatuses = form["statuses"]

	if err != nil {
		return err
	}

	for _, fileHeader := range r.MultipartForm.File["files"] {
		fileHeaders = append(fileHeaders, fileHeader)
		file, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer file.Close()
		files = append(files, file)
	}

	for _, userId := range formPaidByUserIds {
		formattedUserId, err := utils.StringToUint(userId)
		if err != nil {
			return err
		}

		paidByUserIds = append(paidByUserIds, formattedUserId)
	}

	for _, groupId := range formGroupIds {
		formattedGroupId, err := utils.StringToUint(groupId)
		if err != nil {
			return err
		}

		groupIds = append(groupIds, formattedGroupId)
	}

	for _, status := range formStatuses {
		var formattedStatus models.ReceiptStatus
		err = formattedStatus.Scan(strings.TrimSpace(status))
		if err != nil {
			return err
		}

		statuses = append(statuses, formattedStatus)
	}

	command.Files = files
	command.FileHeaders = fileHeaders
	command.PaidByUserIds = paidByUserIds
	command.GroupIds = groupIds
	command.Statuses = statuses

	return nil
}

func (command QuickScanCommand) Validate() structs.ValidatorError {
	vErr := structs.ValidatorError{
		Errors: make(map[string]string),
	}

	var filesLength = len(command.Files)

	if filesLength == 0 {
		vErr.Errors["files"] = "At least one file is required."
	}

	if len(command.PaidByUserIds) != filesLength {
		vErr.Errors["paidByUserId"] = "Paid By User Ids must match the number of files."
	}

	if len(command.GroupIds) != filesLength {
		vErr.Errors["groupIds"] = "Group Ids must match the number of files."
	}

	if len(command.Statuses) != filesLength {
		vErr.Errors["statuses"] = "Statuses must match the number of files."
	}

	if len(command.PaidByUserIds) == 0 {
		vErr.Errors["paidByUserId"] = "Paid By User Id is required."
	}

	if len(command.GroupIds) == 0 {
		vErr.Errors["groupId"] = "Group Id is required."
	}

	if len(command.Statuses) == 0 {
		vErr.Errors["status"] = "Status is required."
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
