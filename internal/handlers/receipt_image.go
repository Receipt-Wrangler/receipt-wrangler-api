package handlers

import (
	"errors"
	"net/http"
	"os"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func UploadReceiptImage(w http.ResponseWriter, r *http.Request) {
	// TODO: Validate size and file type
	basePath, err := os.Getwd()
	errMsg := "Error uploading image."
	token := utils.GetJWT(r)
	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	// Check if data path exists
	if _, err := os.Stat(basePath + "/data"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(basePath+"/data", os.ModePerm)
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return
		}
	}

	userPath := basePath + "/data/" + token.Username

	// Check if user's path exists
	if _, err := os.Stat(userPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(basePath+"/data/"+token.Username, os.ModePerm)
		if err != nil {
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return
		}
	}

	fileData := r.Context().Value("fileData").(structs.FileData)
	// TODO: Fix perms
	err = os.WriteFile(userPath+"/"+fileData.Name, fileData.ImageData, 777)
	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
}
