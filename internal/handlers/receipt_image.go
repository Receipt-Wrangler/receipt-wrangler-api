package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func UploadReceiptImage(w http.ResponseWriter, r *http.Request) {
	// TODO: Validate size and file type
	db := db.GetDB()
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

	fileData := r.Context().Value("fileData").(models.FileData)
	receiptId := fmt.Sprint(fileData.ReceiptId)
	fileName := userPath + "/" + receiptId + "-" + fileData.Name
	// TODO: Fix perms
	err = os.WriteFile(fileName, fileData.ImageData, 777)
	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	if db.Model(models.FileData{}).Create(&fileData).Error != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		os.Remove(fileName)
		return
	}

	w.WriteHeader(200)
}

func GetReceiptImage(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	basePath, err := os.Getwd()
	token := utils.GetJWT(r)
	errMsg := "Error retrieving image."

	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	id := chi.URLParam(r, "id")
	var fileData models.FileData
	var receipt models.Receipt

	if db.Model(models.Receipt{}).Where("id = ?", id).Select("owned_by_user_id").Find(&receipt).Error != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	if receipt.OwnedByUserID != token.UserId {
		utils.WriteCustomErrorResponse(w, errMsg, 403)
		return
	}

	err = db.Model(models.FileData{}).Where("receipt_id = ?", id).First(&fileData).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		w.WriteHeader(204)
		return
	}

	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	path := filepath.Join(basePath, "data", token.Username, id+"-"+fileData.Name)
	fmt.Println(path)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		utils.WriteCustomErrorResponse(w, errMsg, 404)
		return
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	imageData := "data:" + fileData.FileType + ";base64," + base64.StdEncoding.EncodeToString(bytes)

	w.WriteHeader(200)
	w.Write([]byte(imageData))
}
