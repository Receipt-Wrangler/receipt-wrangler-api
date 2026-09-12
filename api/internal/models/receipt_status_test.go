package models

import (
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func TestReceiptStatus_Value(t *testing.T) {
	valid := []ReceiptStatus{OPEN, NEEDS_ATTENTION, RESOLVED, DRAFT, DECLINED}
	for _, v := range valid {
		assertValuerValid(t, string(v), v, string(v))
	}

	// An empty value is accepted and normalized to "".
	assertValuerValid(t, "empty", ReceiptStatus(""), "")
}

func TestReceiptStatus_Value_Invalid(t *testing.T) {
	assertValuerInvalid(t, "bogus", ReceiptStatus("bogus"))
}

func TestReceiptStatus_Scan(t *testing.T) {
	var status ReceiptStatus
	err := status.Scan("OPEN")
	if err != nil {
		utils.PrintTestError(t, err, nil)
	}
	if status != OPEN {
		utils.PrintTestError(t, status, OPEN)
	}
}

func TestReceiptStatuses(t *testing.T) {
	statuses := ReceiptStatuses()

	expected := []interface{}{OPEN, NEEDS_ATTENTION, RESOLVED, DRAFT, DECLINED}
	if len(statuses) != len(expected) {
		utils.PrintTestError(t, len(statuses), len(expected))
	}

	for i, v := range expected {
		if statuses[i] != v {
			utils.PrintTestError(t, statuses[i], v)
		}
	}
}
