package handlers

import (
	"fmt"
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/tesseract"

	"github.com/go-chi/chi/v5"
)

func GetOcrTextForGroup(w http.ResponseWriter, r *http.Request) {
	groupId := chi.URLParam(r, "groupId")
	handler := structs.Handler{
		ErrorMessage: "Error getting ocr text.",
		Writer:       w,
		Request:      r,
		GroupId:      groupId,
		GroupRole:    models.OWNER,
		UserRole:     models.ADMIN,
		ResponseType: constants.APPLICATION_ZIP,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetJWT(r)

			fileDataResults, err := services.GetReceiptImagesForGroup(groupId, simpleutils.UintToString(token.UserId))
			if err != nil {
				return http.StatusInternalServerError, err
			}

			fileRepository := repositories.NewFileRepository(nil)
			ocrTextResults := make([]string, 0)
			for _, fileData := range fileDataResults {
				filePath, err := fileRepository.BuildFilePath(simpleutils.UintToString(fileData.ReceiptId), simpleutils.UintToString(fileData.ID), fileData.Name)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				ocrText, err := tesseract.ReadImage(filePath)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				ocrTextResults = append(ocrTextResults, ocrText)
			}

			fmt.Println(ocrTextResults)

			w.WriteHeader(200)
			// w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
