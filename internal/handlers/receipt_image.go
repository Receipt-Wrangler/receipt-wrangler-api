package handlers

import (
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

func UploadReceiptImage(w http.ResponseWriter, r *http.Request) {
	errMessage := "Error uploading image."
	fileRepository := repositories.NewFileRepository(nil)

	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMessage, http.StatusInternalServerError)
	}

	receiptId, err := simpleutils.StringToUint(r.Form.Get("receiptId"))
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMessage, http.StatusInternalServerError)
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMessage, http.StatusInternalServerError)
	}
	defer file.Close()

	// TODO: Validate size
	handler := structs.Handler{
		ErrorMessage: errMessage,
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
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

			fileDataView := models.FileDataView{
				Id:           createdFile.ID,
				EncodedImage: encodedImage,
			}

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
	handler := structs.Handler{
		ErrorMessage: "Error retrieving image.",
		Writer:       w,
		Request:      r,
		ResponseType: "",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := repositories.GetDB()
			id := chi.URLParam(r, "id")
			var fileData models.FileData
			var receipt models.Receipt
			var bytes []byte
			var fileType string
			result := models.FileDataView{}

			err := db.Model(models.FileData{}).Where("id = ?", id).First(&fileData).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = db.Model(models.Receipt{}).Where("id = ?", fileData.ReceiptId).Select("id").Find(&receipt).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			fileRepository := repositories.NewFileRepository(nil)
			bytes, err = fileRepository.GetBytesForFileData(fileData)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			fileType, err = fileRepository.GetFileType(bytes)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			imageData := "data:" + fileType + ";base64," + base64.StdEncoding.EncodeToString(bytes)
			result.EncodedImage = imageData
			result.Id = fileData.ID

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
	handler := structs.Handler{
		ErrorMessage: "Error deleting image.",
		Writer:       w,
		Request:      r,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := repositories.GetDB()

			id := chi.URLParam(r, "id")
			var fileData models.FileData

			err := db.Model(models.FileData{}).Where("id = ?", id).First(&fileData).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = db.Delete(fileData).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			fileRepository := repositories.NewFileRepository(nil)
			path, err := fileRepository.BuildFilePath(simpleutils.UintToString(fileData.ReceiptId), id, fileData.Name)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = os.Remove(path)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(200)

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
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			receiptImageId := r.URL.Query().Get("receiptImageId")
			filledReceipt := models.Receipt{}

			if len(receiptImageId) > 0 {
				errCode, err := validateReceiptImageAccess(r, models.VIEWER, receiptImageId)
				if err != nil {
					return errCode, err
				}

				filledReceipt, err = services.ReadReceiptImage(receiptImageId)
				if err != nil {
					return http.StatusInternalServerError, err
				}
			} else {
				err := r.ParseMultipartForm(50 << 20)
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

				filledReceipt, err = services.MagicFillFromImage(magicFillCommand)
				if err != nil {
					return http.StatusInternalServerError, err
				}
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

func ConvertToJpg(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error converting receipt.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
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

			validatedFileType, err := fileRepository.ValidateFileType(fileBytes)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if validatedFileType != constants.APPLICATION_PDF {
				return http.StatusBadRequest, errors.New("file must be a PDF")
			}

			jpgBytes, err := fileRepository.ConvertPdfToJpg(fileBytes)
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
	token := structs.GetJWT(r)
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
