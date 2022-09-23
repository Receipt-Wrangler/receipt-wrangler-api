package handlers

import (
	"encoding/json"
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func GetAllReceipts(w http.ResponseWriter, r *http.Request) {
	// re add user id in claim??
	db := db.GetDB()
	token := utils.GetJWT(r)
	errMsg := "Error retrieving receipts."
	var receipts []models.Receipt

	err := db.Model(models.Receipt{}).Where("owned_by_user_id = ?", token.UserId).Find(&receipts).Error
	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	bytes, err := json.Marshal(receipts)
	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func CreateReceipt(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	token := utils.GetJWT(r)

	errMsg := "Error creating receipts."
	bodyData := r.Context().Value("receipt").(models.Receipt)
	bodyData.OwnedByUserID = token.UserId

	vErr := validateReceipt(bodyData)
	if len(vErr.Errors) > 0 {
		utils.WriteValidatorErrorResponse(w, vErr, 400)
		return
	}

	err := db.Model(models.Receipt{}).Create(&bodyData).Error
	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	bytes, err := json.Marshal(bodyData)
	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func validateReceipt(r models.Receipt) structs.ValidatorError {
	err := structs.ValidatorError{
		Errors: make(map[string]string),
	}

	if len(r.Name) == 0 {
		err.Errors["name"] = "Name is required"
	}

	if r.Amount <= 0 {
		err.Errors["amount"] = "Amount must be greater than zero"
	}

	if r.Date.IsZero() {
		err.Errors["date"] = "Date is required"
	}

	return err
}
