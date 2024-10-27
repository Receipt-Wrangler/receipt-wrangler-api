package handlers

import (
	"net/http"
	"os"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func GetAboutData(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting about data",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			envVersion := os.Getenv("VERSION")
			envBuildDate := os.Getenv("BUILD_DATE")

			version := "latest"

			if envVersion != "" {
				version = envVersion
			}

			about := structs.About{
				Version:   version,
				BuildDate: envBuildDate,
			}

			aboutBytes, err := utils.MarshalResponseData(&about)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			w.Write(aboutBytes)
			return 0, nil
		},
	}

	HandleRequest(handler)
}
