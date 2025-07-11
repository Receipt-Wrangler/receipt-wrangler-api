package commands

import (
	"github.com/shopspring/decimal"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
)

type UpsertShareCommand struct {
	Amount          decimal.Decimal         `json:"amount"`
	ChargedToUserId uint                    `json:"chargedToUserId"`
	IsTaxed         bool                    `json:"isTaxed"`
	Name            string                  `json:"name"`
	ReceiptId       uint                    `json:"receiptId"`
	Status          models.ShareStatus      `json:"status"`
	Categories      []UpsertCategoryCommand `json:"categories"`
	Tags            []UpsertTagCommand      `json:"tags"`
}

func (share *UpsertShareCommand) Validate(receiptAmount decimal.Decimal, isCreate bool) structs.ValidatorError {
	errors := make(map[string]string)
	vErr := structs.ValidatorError{}

	if share.Amount.IsZero() {
		errors["amount"] = "Amount is required"
	}

	if share.Amount.GreaterThan(receiptAmount) {
		errors["amount"] = "Amount cannot be greater than receipt amount"
	}

	if share.Amount.LessThanOrEqual(decimal.Zero) {
		errors["amount"] = "Amount must be greater than zero"
	}

	if len(share.Name) == 0 {
		errors["name"] = "Name is required"
	}

	if !isCreate {
		if share.ReceiptId == 0 {
			errors["receiptId"] = "Receipt Id is required"
		}
	}

	if share.ChargedToUserId == 0 {
		errors["chargedToUserId"] = "Charged To User Id is required"
	}

	if len(share.Status) == 0 {
		errors["status"] = "Status is required"
	}

	vErr.Errors = errors
	return vErr
}
