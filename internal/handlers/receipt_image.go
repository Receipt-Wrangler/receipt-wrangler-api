package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
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
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	// Check if data path exists
	if _, err := os.Stat(basePath + "/data"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(basePath+"/data", os.ModePerm)
		if err != nil {
			handler_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return
		}
	}

	userPath := filepath.Join(basePath, "data", token.Username)

	// Check if user's path exists
	if _, err := os.Stat(userPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(userPath, os.ModePerm)
		if err != nil {
			handler_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return
		}
	}

	fileData := r.Context().Value("fileData").(models.FileData)
	receiptId := fmt.Sprint(fileData.ReceiptId)
	path, err := BuildFilePath(token.Username, receiptId, fileData.Name)

	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	// TODO: Fix perms
	err = os.WriteFile(path, fileData.ImageData, 777)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	err = db.Model(models.FileData{}).Create(&fileData).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		os.Remove(path)
		return
	}

	w.WriteHeader(200)
}

func GetReceiptImage(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	token := utils.GetJWT(r)
	errMsg := "Error retrieving image."

	id := chi.URLParam(r, "id")
	var fileData models.FileData
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", id).Select("owned_by_user_id").Find(&receipt).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	if receipt.OwnedByUserID != token.UserId {
		handler_logger.Print("Unauthorized access")
		utils.WriteCustomErrorResponse(w, errMsg, 403)
		return
	}

	err = db.Model(models.FileData{}).Where("receipt_id = ?", id).First(&fileData).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		handler_logger.Print(err.Error())
		w.WriteHeader(204)
		return
	}

	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	path, err := BuildFilePath(token.Username, id, fileData.Name)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 404)
		return
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 404)
		return
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	imageData := "data:" + fileData.FileType + ";base64," + base64.StdEncoding.EncodeToString(bytes)

	w.WriteHeader(200)
	w.Write([]byte(imageData))
}

func BuildFilePath(uname string, rid string, fname string) (string, error) {
	basePath, err := os.Getwd()
	if err != nil {
		handler_logger.Print(err.Error())
		return "", err
	}

	fileName := rid + "-" + fname
	path := filepath.Join(basePath, "data", uname, fileName)

	return path, nil
}
