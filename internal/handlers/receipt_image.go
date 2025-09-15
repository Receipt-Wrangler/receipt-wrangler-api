package handlers

import (
	"net/http"
	"os"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"time"

	"github.com/go-chi/chi/v5"
)

func UploadReceiptImage(w http.ResponseWriter, r *http.Request) {
	errMessage := "Error uploading image."
	fileRepository := repositories.NewFileRepository(nil)

	err := r.ParseMultipartForm(constants.MultipartFormMaxSize)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		utils.WriteCustomErrorResponse(w, errMessage, http.StatusInternalServerError)
	}

	receiptId, err := utils.StringToUint(r.Form.Get("receiptId"))
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		utils.WriteCustomErrorResponse(w, errMessage, http.StatusInternalServerError)
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		utils.WriteCustomErrorResponse(w, errMessage, http.StatusInternalServerError)
	}
	defer file.Close()

	// TODO: Validate size
	handler := structs.Handler{
		ErrorMessage: errMessage,
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		ReceiptId:    r.Form.Get("receiptId"),
		GroupRole:    models.EDITOR,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			fileBytes := make([]byte, fileHeader.Size)

			_, err = file.Read(fileBytes)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			_, err = fileRepository.ValidateFileType(fileBytes)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			fileImageRepository := repositories.NewReceiptImageRepository(nil)
			fileData := models.FileData{
				Name:      fileHeader.Filename,
				Size:      uint(fileHeader.Size),
				ReceiptId: receiptId,
			}

			createdFile, err := fileImageRepository.CreateReceiptImage(fileData, fileBytes)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			fileBytes, err = fileRepository.GetBytesForFileData(createdFile)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			encodedImage, err := fileRepository.BuildEncodedImageString(fileBytes)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			fileDataView := models.FileDataView{}.FromFileData(createdFile)
			fileDataView.EncodedImage = encodedImage

			bytes, err := utils.MarshalResponseData(fileDataView)
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

func GetReceiptImage(w http.ResponseWriter, r *http.Request) {
	db := repositories.GetDB()
	errorMessage := "Error retrieving image."
	var fileData models.FileData
	id := chi.URLParam(r, "id")

	err := db.Model(models.FileData{}).Where("id = ?", id).First(&fileData).Error
	if err != nil {
		utils.WriteCustomErrorResponse(w, errorMessage, http.StatusInternalServerError)
		return
	}
	stringReceiptId := utils.UintToString(fileData.ReceiptId)

	handler := structs.Handler{
		ErrorMessage: "Error retrieving image.",
		ReceiptId:    stringReceiptId,
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			var receipt models.Receipt
			var bytes []byte
			result := models.FileDataView{}

			err = db.Model(models.Receipt{}).Where("id = ?", fileData.ReceiptId).Select("id").Find(&receipt).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			fileRepository := repositories.NewFileRepository(nil)
			bytes, err = fileRepository.GetBytesForFileData(fileData)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			imageData, err := fileRepository.BuildEncodedImageString(bytes)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			result = models.FileDataView{}.FromFileData(fileData)
			result.EncodedImage = imageData

			resultBytes, err := utils.MarshalResponseData(result)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			w.Write(resultBytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func DownloadReceiptImage(w http.ResponseWriter, r *http.Request) {
	db := repositories.GetDB()
	errorMessage := "Error retrieving image."
	var fileData models.FileData
	id := chi.URLParam(r, "id")

	err := db.Model(models.FileData{}).Where("id = ?", id).First(&fileData).Error
	if err != nil {
		utils.WriteCustomErrorResponse(w, errorMessage, http.StatusInternalServerError)
		return
	}
	stringReceiptId := utils.UintToString(fileData.ReceiptId)

	handler := structs.Handler{
		ErrorMessage: "Error downloading image.",
		ReceiptId:    stringReceiptId,
		GroupRole:    models.VIEWER,
		Writer:       w,
		Request:      r,
		ResponseType: "",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			fileRepository := repositories.NewFileRepository(nil)

			path, err := fileRepository.BuildFilePath(utils.UintToString(fileData.ReceiptId), utils.UintToString(fileData.ID), fileData.Name)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.Header().Set("Content-Disposition", "attachment; filename="+fileData.Name)

			http.ServeFile(w, r, path)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func RemoveReceiptImage(w http.ResponseWriter, r *http.Request) {
	db := repositories.GetDB()
	errorMessage := "Error deleting image."

	id := chi.URLParam(r, "id")
	var fileData models.FileData

	err := db.Model(models.FileData{}).Where("id = ?", id).First(&fileData).Error
	if err != nil {
		utils.WriteCustomErrorResponse(w, errorMessage, http.StatusInternalServerError)
	}
	stringReceiptId := utils.UintToString(fileData.ReceiptId)

	handler := structs.Handler{
		ErrorMessage: "Error deleting image.",
		ReceiptId:    stringReceiptId,
		Writer:       w,
		Request:      r,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			err = db.Delete(fileData).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			fileRepository := repositories.NewFileRepository(nil)
			path, err := fileRepository.BuildFilePath(utils.UintToString(fileData.ReceiptId), id, fileData.Name)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = os.Remove(path)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			return 0, nil
		},
	}

	HandleRequest(handler)
}

func MagicFillFromImage(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error performing magic fill.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			receiptImageId := r.URL.Query().Get("receiptImageId")
			receiptCommand := commands.UpsertReceiptCommand{}
			systemTaskService := services.NewSystemTaskService(nil)
			token := structs.GetClaims(r)

			var startTimer time.Time
			var endTimer time.Time

			if len(receiptImageId) > 0 {
				errCode, err := validateReceiptImageAccess(r, models.VIEWER, receiptImageId)
				if err != nil {
					return errCode, err
				}

				startTimer = time.Now()
				command, metadata, err := services.ReadReceiptImage(receiptImageId)
				endTimer = time.Now()

				_, taskErr := systemTaskService.CreateSystemTasksFromMetadata(
					metadata,
					startTimer,
					endTimer,
					models.MAGIC_FILL,
					&token.UserId,
					nil,
					"",
					nil,
				)
				if taskErr != nil {
					return http.StatusInternalServerError, taskErr
				}

				if err != nil {
					return http.StatusInternalServerError, err
				}

				receiptCommand = command
			} else {
				err := r.ParseMultipartForm(constants.MultipartFormMaxSize)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				file, fileHeader, err := r.FormFile("file")
				if err != nil {
					return http.StatusInternalServerError, err
				}
				defer file.Close()

				fileBytes := make([]byte, fileHeader.Size)
				_, err = file.Read(fileBytes)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				magicFillCommand := commands.MagicFillCommand{
					ImageData: fileBytes,
					Filename:  fileHeader.Filename,
				}

				startTimer = time.Now()
				command, metadata, err := services.MagicFillFromImage(magicFillCommand, "")
				endTimer = time.Now()

				_, taskErr := systemTaskService.CreateSystemTasksFromMetadata(
					metadata,
					startTimer,
					endTimer,
					models.MAGIC_FILL,
					&token.UserId,
					nil,
					"",
					nil,
				)
				if taskErr != nil {
					return http.StatusInternalServerError, taskErr
				}

				if err != nil {
					return http.StatusInternalServerError, err
				}

				receiptCommand = command
			}

			bytes, err := utils.MarshalResponseData(receiptCommand)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func ConvertToJpg(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error converting image.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			fileRepository := repositories.NewFileRepository(nil)
			result := make(map[string]string)

			err := r.ParseMultipartForm(50 << 20)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			file, fileHeader, err := r.FormFile("file")
			if err != nil {
				return http.StatusInternalServerError, err
			}

			fileBytes := make([]byte, fileHeader.Size)
			_, err = file.Read(fileBytes)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			_, err = fileRepository.ValidateFileType(fileBytes)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			jpgBytes, err := fileRepository.GetBytesFromImageBytes(fileBytes)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			encodedString, err := fileRepository.BuildEncodedImageString(jpgBytes)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			result["encodedImage"] = encodedString
			bytes, err := utils.MarshalResponseData(result)
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
	token := structs.GetClaims(r)
	groupService := services.NewGroupService(nil)

	receiptImageIdUint, err := utils.StringToUint(receiptImageId)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	receiptImageRepository := repositories.NewReceiptImageRepository(nil)

	receiptImage, err := receiptImageRepository.GetReceiptImageById(receiptImageIdUint)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	err = groupService.ValidateGroupRole(groupRole, utils.UintToString(receiptImage.Receipt.GroupId), utils.UintToString(token.UserId))
	if err != nil {
		return http.StatusForbidden, err
	}

	return 0, nil
}
