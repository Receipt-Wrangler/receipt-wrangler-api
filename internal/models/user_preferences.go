package models

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/utils"
)

type UserPrefernces struct {
	BaseModel
	QuickScanDefaultGroupId  *uint         `json:"quickScanDefaultGroup"`
	QuickScanDefaultGroup    *Group        `json:"-"`
	QuickScanDefaultPaidById *uint         `json:"quickScanDefaultPaidBy"`
	QuickScanDefaultPaidBy   *User         `json:"-"`
	QuickScanDefaultStatus   ReceiptStatus `json:"quickScanDefaultStatus"`
}

func (userPrefernces *UserPrefernces) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &userPrefernces)
	if err != nil {
		return err
	}

	return nil
}
