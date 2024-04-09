package commands

import (
	"receipt-wrangler/api/internal/structs"
)

type UpsertCommentCommand struct {
	Comment   string `json:"comment"`
	ReceiptId uint   `json:"receiptId"`
	UserId    uint   `json:"userId"`
}

func (comment *UpsertCommentCommand) Validate(userRequestId uint) structs.ValidatorError {
	errors := make(map[string]string)
	vErr := structs.ValidatorError{}

	if len(comment.Comment) == 0 {
		errors["comment"] = "Comment is required"
	}

	if comment.ReceiptId == 0 {
		errors["receiptId"] = "Receipt Id is required"
	}

	if comment.UserId == 0 {
		errors["userId"] = "User Id is required"
	}

	if comment.UserId != userRequestId {
		errors["userId"] = "Bad user id"
	}

	vErr.Errors = errors
	return vErr
}
