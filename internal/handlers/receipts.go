package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"net/http"
	"os"
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

func UploadReceiptImage(w http.ResponseWriter, r *http.Request) {
	basePath, err := os.Getwd()
	errMsg := "Error uploading image."
	token := utils.GetJWT(r)
	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	// Check if data path exists
	if _, err := os.Stat(basePath + "/data"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(basePath+"/data", os.ModePerm)
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return
		}
	}

	userPath := basePath + "/data/" + token.Username

	// Check if user's path exists
	if _, err := os.Stat(userPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(basePath+"/data/"+token.Username, os.ModePerm)
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return
		}
	}

	bodyData, err := utils.GetBodyData(w, r)
	var fileData structs.FileData

	err = json.Unmarshal(bodyData, &fileData)
	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	// Somewhere in the same package
	f, err := os.Create("outimage.jpg")
	if err != nil {
		// Handle error
	}
	defer f.Close()

	// Specify the quality, between 0-100
	// Higher is better
	opt := jpeg.Options{
		Quality: 90,
	}
	err = jpeg.Encode(f, img, &opt)
	if err != nil {
		// Handle error
	}

	// err = os.WriteFile(userPath+"/"+fileData.Name, []byte(fileData.ImageData), 777)
	// if err != nil {
	// 	utils.WriteCustomErrorResponse(w, errMsg, 500)
	// 	return
	// }

	w.WriteHeader(200)
	fmt.Println(basePath)
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
