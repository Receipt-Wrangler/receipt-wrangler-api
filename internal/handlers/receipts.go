package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
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
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	bytes, err := json.Marshal(receipts)
	if err != nil {
		handler_logger.Print(err.Error())
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

	err := db.Model(models.Receipt{}).Select("*").Create(&bodyData).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	bytes, err := json.Marshal(bodyData)
	if err != nil {
		handler_logger.Print(err.Error())
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

	err := db.Model(models.Receipt{}).Where("id = ?", id).Preload(clause.Associations).Find(&receipt).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 404)
		return
	}

	if receipt.OwnedByUserID != token.UserId {
		handler_logger.Print("Unauthorized")
		utils.WriteCustomErrorResponse(w, errMsg, 403)
		return
	}

	bytes, err := json.Marshal(receipt)
	if err != nil {
		handler_logger.Print(err.Error())
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
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	bodyData := r.Context().Value("receipt").(models.Receipt)
	bodyData.ID = uint(u64Id)

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&bodyData).Select("*").Omit("ID, OwnedByUserID").Where("id = ?", uint(u64Id)).Save(bodyData).Error
		if txErr != nil {
			handler_logger.Print(txErr.Error())
			return txErr
		}

		txErr = db.Model(&bodyData).Association("Tags").Replace(bodyData.Tags)
		if txErr != nil {
			handler_logger.Print(txErr.Error())
			return txErr
		}

		txErr = db.Model(&bodyData).Association("Categories").Replace(bodyData.Categories)
		if txErr != nil {
			handler_logger.Print(txErr.Error())
			return txErr
		}

		txErr = db.Model(&bodyData).Association("ReceiptItems").Replace(bodyData.ReceiptItems)
		if txErr != nil {
			handler_logger.Print(txErr.Error())
			return txErr
		}

		// return nil will commit the whole transaction
		return nil
	})

	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
}

func ToggleIsResolved(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	var err error
	var receipt models.Receipt

	errMsg := "Error toggling is resolved receipt."
	id := chi.URLParam(r, "id")

	err = db.Model(models.Receipt{}).Select("id, is_resolved").Where("id = ?", id).Find(&receipt).Error

	if err != nil {
		handler_logger.Print(err)
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}
	err = db.Model(&receipt).Update("is_resolved", !receipt.IsResolved).Error
	if err != nil {
		handler_logger.Print(err)
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
}

func DeleteReceipt(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	var err error
	var receipt models.Receipt
	errMsg := "Error deleting receipt."
	token := utils.GetJWT(r)

	id := chi.URLParam(r, "id")

	err = db.Model(models.Receipt{}).Where("id = ?", id).Preload("ImageFiles").Find(&receipt).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		err = tx.Select(clause.Associations).Delete(&receipt).Error
		if err != nil {
			handler_logger.Print(err.Error())
			return err
		}

		err = tx.Delete(&receipt).Error
		if err != nil {
			handler_logger.Print(err.Error())
			return err
		}

		for _, f := range receipt.ImageFiles {
			path, _ := BuildFilePath(token.Username, id, f.Name)
			os.Remove(path)
		}

		return nil
	})

	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
}
