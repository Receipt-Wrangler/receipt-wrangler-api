package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetAllReceipts(w http.ResponseWriter, r *http.Request) {
	// re add user id in claim??
	db := db.GetDB()
	token := utils.GetJWT(r)
	errMsg := "Error retrieving receipts."
	var receipts []models.Receipt

	err := db.Model(models.Receipt{}).Where("owned_by_user_id = ?", token.UserId).Preload("Tags").Preload("Categories").Find(&receipts).Error
	if err != nil {
		log.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	bytes, err := json.Marshal(receipts)
	if err != nil {
		log.Print(err.Error())
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
		log.Print(vErr.Errors)
		utils.WriteValidatorErrorResponse(w, vErr, 400)
		return
	}

	err := db.Model(models.Receipt{}).Create(&bodyData).Error
	if err != nil {
		log.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	bytes, err := json.Marshal(bodyData)
	if err != nil {
		log.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func GetReceipt(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	var receipt models.Receipt
	errMsg := "Error retrieving receipt."
	token := utils.GetJWT(r)

	id := chi.URLParam(r, "id")

	if db.Model(models.Receipt{}).Where("id = ?", id).Preload(clause.Associations).Find(&receipt).Error != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 404)
		return
	}

	if receipt.OwnedByUserID != token.UserId {
		log.Print("Unauthorized")
		utils.WriteCustomErrorResponse(w, errMsg, 403)
		return
	}

	bytes, err := json.Marshal(receipt)
	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func UpdateReceipt(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()

	errMsg := "Error updating receipt."
	id := chi.URLParam(r, "id")
	u64Id, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	bodyData := r.Context().Value("receipt").(models.Receipt)
	bodyData.ID = uint(u64Id)

	_, err, code := validateAccess(r, id)
	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, int(code))
		return
	}

	vErr := validateReceipt(bodyData)
	if len(vErr.Errors) > 0 {
		utils.WriteValidatorErrorResponse(w, vErr, 500)
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&bodyData).Omit("ID, OwnedByUserID").Where("id = ?", uint(u64Id)).Save(bodyData).Error
		if txErr != nil {
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return txErr
		}

		txErr = db.Model(&bodyData).Association("Tags").Replace(bodyData.Tags)
		if txErr != nil {
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return txErr
		}

		err = db.Model(&bodyData).Association("Categories").Replace(bodyData.Categories)
		if txErr != nil {
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return txErr
		}

		// return nil will commit the whole transaction
		return nil
	})

	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
}

func DeleteReceipt(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	var receipt models.Receipt
	errMsg := "Error deleting receipt."
	var fileData models.FileData
	token := utils.GetJWT(r)

	id := chi.URLParam(r, "id")

	receipt, err, errCode := validateAccess(r, id)
	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, int(errCode))
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(models.FileData{}).Where("receipt_id = ?", receipt.ID).Find(&fileData).Error
		if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
			return err
		} else {
			err = tx.Delete(&models.FileData{}, fileData.ID).Error
			if err != nil {
				return err
			}
		}

		err = tx.Delete(&models.Receipt{}, id).Error
		if err != nil {
			return err
		}

		if fmt.Sprint(fileData.ReceiptId) == id {
			path, _ := BuildFilePath(token.Username, id, fileData.Name)
			os.Remove(path)
		}

		// return nil will commit the whole transaction
		return nil
	})

	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
}

func validateAccess(r *http.Request, rid string) (models.Receipt, error, uint) {
	db := db.GetDB()
	token := utils.GetJWT(r)
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", rid).Find(&receipt).Error
	if err != nil {
		return receipt, err, 404
	}

	if receipt.OwnedByUserID != token.UserId {
		return receipt, err, 403
	}

	return receipt, nil, 200
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
