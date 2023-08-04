package handlers

import (
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"receipt-wrangler/api/internal/constants"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
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
	err = utils.DirectoryExists(basePath+"/data", true)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	// Get initial group directory to see if it exists
	fileData := r.Context().Value("fileData").(models.FileData)
	filePath, err := utils.BuildFilePath(simpleutils.UintToString(fileData.ReceiptId), "", fileData.Name)
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
	err = utils.DirectoryExists(groupDir, true)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	// Rebuild file path with correct file id
	filePath, err = utils.BuildFilePath(simpleutils.UintToString(fileData.ReceiptId), simpleutils.UintToString(fileData.ID), fileData.Name)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	err = utils.WriteFile(filePath, fileData.ImageData)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	fileData.ImageData = make([]byte, 0)
	bytes, err := utils.MarshalResponseData(fileData)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func GetReceiptImage(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error retrieving image.",
		Writer:       w,
		Request:      r,
		ResponseType: "",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := db.GetDB()
			id := chi.URLParam(r, "id")
			var fileData models.FileData
			var receipt models.Receipt
			result := make(map[string]string)

			err := db.Model(models.FileData{}).Where("id = ?", id).First(&fileData).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = db.Model(models.Receipt{}).Where("id = ?", fileData.ReceiptId).Select("id").Find(&receipt).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.GetBytesForFileData(fileData)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			imageData := "data:" + fileData.FileType + ";base64," + base64.StdEncoding.EncodeToString(bytes)
			result["encodedImage"] = imageData

			resultBytes, err := utils.MarshalResponseData(result)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(200)
			w.Write([]byte(resultBytes))

			return 0, nil
		},
	}

	HandleRequest(handler)
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

	path, err := utils.BuildFilePath(simpleutils.UintToString(fileData.ReceiptId), id, fileData.Name)
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

func MagicFillFromImage(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error performing magic fill.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			receiptImageId := r.URL.Query().Get("receiptImageId")
			errCode, err := validateReceiptImageAccess(r, models.VIEWER, receiptImageId)
			if err != nil {
				return errCode, err
			}

			filledReceipt, err := services.ReadReceiptImage(receiptImageId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(filledReceipt)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(200)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func validateReceiptImageAccess(r *http.Request, groupRole models.GroupRole, receiptImageId string) (int, error) {
	token := utils.GetJWT(r)
	receiptImageIdUint, err := simpleutils.StringToUint(receiptImageId)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	receiptImageRepository := repositories.NewReceiptImageRepository(nil)

	receiptImage, err := receiptImageRepository.GetReceiptImageById(receiptImageIdUint)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	err = services.ValidateGroupRole(groupRole, simpleutils.UintToString(receiptImage.Receipt.GroupId), simpleutils.UintToString(token.UserId))
	if err != nil {
		return http.StatusForbidden, err
	}

	return 0, nil
}
