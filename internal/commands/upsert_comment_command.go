package commands

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

type UpsertCommentCommand struct {
	Comment   string `json:"comment"`
	ReceiptId uint   `json:"receiptId"`
	UserId    *uint  `json:"userId"`
}

func (comment *UpsertCommentCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request, isCreate bool) error {
	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &comment)
	if err != nil {
		return err
	}

	if isCreate {
		token := structs.GetClaims(r)
		comment.UserId = &token.UserId
	}

	return nil
}

func (comment *UpsertCommentCommand) Validate(userRequestId uint, isCreate bool) structs.ValidatorError {
	errors := make(map[string]string)
	vErr := structs.ValidatorError{}

	if len(comment.Comment) == 0 {
		errors["comment"] = "Comment is required"
	}

	if !isCreate {
		if comment.ReceiptId == 0 {
			errors["receiptId"] = "Receipt Id is required"
		}
	}

	if comment.UserId == nil {
		errors["userId"] = "User Id is required"
	}

	if !isCreate {
		if comment.UserId != nil && *comment.UserId != userRequestId {
			errors["userId"] = "Bad user id"
		}

	}

	vErr.Errors = errors
	return vErr
}
