package handlers

import (
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

func UploadReceiptImage(w http.ResponseWriter, r *http.Request) {
	// TODO: Validate size and file type
	db := db.GetDB()
	basePath, err := os.Getwd()
	errMsg := "Error uploading image."

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

	// Get initial group directory to see if it exists
	fileData := r.Context().Value("fileData").(models.FileData)
	filePath, err := BuildFilePath(utils.UintToString(fileData.ReceiptId), "", fileData.Name)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}
	groupDir, _ := filepath.Split(filePath)

	err = db.Model(models.FileData{}).Create(&fileData).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		os.Remove(filePath)
		return
	}

	// Check if group's path exists
	if _, err := os.Stat(groupDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(groupDir, os.ModePerm)
		if err != nil {
			handler_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return
		}
	}

	// Rebuild file path with correct file id
	filePath, err = BuildFilePath(utils.UintToString(fileData.ReceiptId), utils.UintToString(fileData.ID), fileData.Name)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	// TODO: Fix perms
	err = os.WriteFile(filePath, fileData.ImageData, 777)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
}

func GetReceiptImage(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	errMsg := "Error retrieving image."
	id := chi.URLParam(r, "id")
	var fileData models.FileData
	var receipt models.Receipt

	err := db.Model(models.FileData{}).Where("id = ?", id).First(&fileData).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	err = db.Model(models.Receipt{}).Where("id = ?", fileData.ReceiptId).Select("id").Find(&receipt).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	path, err := BuildFilePath(utils.UintToString(fileData.ReceiptId), id, fileData.Name)
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

func RemoveReceiptImage(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	errMsg := "Error retrieving image."

	id := chi.URLParam(r, "id")
	var fileData models.FileData

	err := db.Model(models.FileData{}).Where("id = ?", id).First(&fileData).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	err = db.Delete(fileData).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	path, err := BuildFilePath(utils.UintToString(fileData.ReceiptId), id, fileData.Name)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	err = os.Remove(path)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
}

func BuildFilePath(rid string, fId string, fname string) (string, error) {
	db := db.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", rid).Select("group_id").Find(&receipt).Error
	if err != nil {
		return "", err
	}

	basePath, err := os.Getwd()
	if err != nil {
		handler_logger.Print(err.Error())
		return "", err
	}

	group, err := repositories.GetGroupById(utils.UintToString(receipt.GroupId))
	if err != nil {
		return "", err
	}

	strGroupId := utils.UintToString(group.ID)

	fileName := rid + "-" + fId + "-" + fname
	groupPath := strGroupId + "-" + group.Name
	path := filepath.Join(basePath, "data", groupPath, fileName)

	return path, nil
}
