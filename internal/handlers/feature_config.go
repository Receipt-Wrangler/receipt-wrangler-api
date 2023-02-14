package handlers

import (
	"net/http"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/utils"
)

func GetFeatureConfig(w http.ResponseWriter, r *http.Request) {
	errMsg := "Error getting feature config."
	featureConfig := config.GetFeatureConfig()

	bytes, err := utils.MarshalResponseData(featureConfig)
	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		handler_logger.Print(err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}
