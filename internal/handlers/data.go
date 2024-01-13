package handlers

import (
	"fmt"
	"net/http"
	"os"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"strings"

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

			ocrResults, err := services.ReadAllReceiptImagesForGroup(groupId, simpleutils.UintToString(token.UserId))
			if err != nil {
				return http.StatusInternalServerError, err
			}

			fileRepository := repositories.NewFileRepository(nil)

			for i, exportResults := range ocrResults {
				fileName := strings.Split(exportResults.Filename, ".")[0]
				tempPath := fileRepository.GetTempDirectoryPath() + "/" + fileName + "-" + fmt.Sprint(i) + ".txt"
				os.WriteFile(tempPath, []byte(exportResults.OcrText), 0600)
				if err != nil {
					return http.StatusInternalServerError, err
				}
			}

			// create files
			// add files to zip

			w.WriteHeader(200)
			// w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
