package handlers

import (
	"encoding/base64"
	"encoding/json"
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
	// TODO: Validate size
	handler := structs.Handler{
		ErrorMessage: "Error retrieving image.",
		Writer:       w,
		Request:      r,
		ResponseType: "",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			fileImageRepository := repositories.NewReceiptImageRepository(nil)
			fileData := r.Context().Value("fileData").(models.FileData)

			createdFile, err := fileImageRepository.CreateReceiptImage(fileData)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(createdFile)
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
			result := make(map[string]string)

			err := db.Model(models.FileData{}).Where("id = ?", id).First(&fileData).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = db.Model(models.Receipt{}).Where("id = ?", fileData.ReceiptId).Select("id").Find(&receipt).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			fileRepository := repositories.NewFileRepository(nil)
			if fileData.FileType == constants.ANY_IMAGE {
				bytes, err = fileRepository.GetBytesForFileData(fileData)
				if err != nil {
					return http.StatusInternalServerError, err
				}
				fileType = fileData.FileType
			} else if fileData.FileType == constants.APPLICATION_PDF {
				fileRepository := repositories.NewFileRepository(nil)
				filePath, err := fileRepository.BuildFilePath(simpleutils.UintToString(fileData.ReceiptId), simpleutils.UintToString(fileData.ID), fileData.Name)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				bytes, err = fileRepository.ConvertPdfToImage(filePath)
				if err != nil {
					return http.StatusInternalServerError, err
				}
				fileType = "image/jpeg"
			}

			imageData := "data:" + fileType + ";base64," + base64.StdEncoding.EncodeToString(bytes)
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

			body, err := utils.GetBodyData(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if len(receiptImageId) > 0 {
				errCode, err := validateReceiptImageAccess(r, models.VIEWER, receiptImageId)
				if err != nil {
					return errCode, err
				}

				filledReceipt, err = services.ReadReceiptImage(receiptImageId)
				if err != nil {
					return http.StatusInternalServerError, err
				}
			} else if len(body) > 0 {
				var magicFillCommand commands.MagicFillCommand
				err := json.Unmarshal(body, &magicFillCommand)
				if err != nil {
					return http.StatusInternalServerError, err
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
