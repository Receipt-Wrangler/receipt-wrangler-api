package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
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
			zipFilename := "results.zip"

			ocrResults, err := services.ReadAllReceiptImagesForGroup(groupId, simpleutils.UintToString(token.UserId))
			if err != nil {
				return http.StatusInternalServerError, err
			}

			fileRepository := repositories.NewFileRepository(nil)
			tempFilenames := []string{}

			for i, exportResults := range ocrResults {
				filename := strings.Split(exportResults.Filename, ".")[0] + "-" + fmt.Sprint(i) + ".txt"
				tempPath := filepath.Join(fileRepository.GetTempDirectoryPath(), filename)
				err := os.WriteFile(tempPath, []byte(exportResults.OcrText), 0600)
				defer os.Remove(tempPath)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				tempFilenames = append(tempFilenames, filename)
			}

			zipPath, err := fileRepository.CreateZipFromTempFiles(zipFilename, tempFilenames)
			defer os.Remove(zipPath)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.ReadFile(zipPath)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.Header().Set("Content-Disposition", "attachment; filename="+zipFilename)

			w.WriteHeader(200)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
