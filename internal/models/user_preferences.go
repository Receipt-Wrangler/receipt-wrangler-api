package models

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/utils"
)

type UserPrefernces struct {
	BaseModel
	UserId                   uint           `gorm:"not null; uniqueIndex" json:"userId"`
	User                     *User          `json:"-"`
	ShowLargeImagePreviews   bool           `json:"showLargeImagePreviews"`
	QuickScanDefaultGroupId  *uint          `json:"quickScanDefaultGroupId"`
	QuickScanDefaultGroup    *Group         `json:"-"`
	QuickScanDefaultPaidById *uint          `json:"quickScanDefaultPaidById"`
	QuickScanDefaultPaidBy   *User          `json:"-"`
	QuickScanDefaultStatus   ReceiptStatus  `json:"quickScanDefaultStatus"`
	UserShortcuts            []UserShortcut `json:"userShortcuts"`
}

func (userPreferences *UserPrefernces) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &userPreferences)
	if err != nil {
		return err
	}

	return nil
}
