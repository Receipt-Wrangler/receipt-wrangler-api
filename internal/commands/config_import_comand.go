package commands

import (
	"encoding/json"
	"errors"
	"github.com/gabriel-vasile/mimetype"
	"mime/multipart"
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/structs"
)

type ConfigImportCommand struct {
	Files      multipart.File        `json:"file"`
	FileHeader *multipart.FileHeader `json:"fileHeader"`
	Config     structs.Config        `json:"config"`
}

func (command *ConfigImportCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseMultipartForm(constants.MULTIPART_FORM_MAX_SIZE)
	if err != nil {
		return err
	}

	files := make([]multipart.File, 0)
	fileHeaders := make([]*multipart.FileHeader, 0)

	for _, fileHeader := range r.MultipartForm.File["file"] {
		fileHeaders = append(fileHeaders, fileHeader)
		file, err := fileHeader.Open()
		if err != nil {
			return err
		}

		mimeType, err := mimetype.DetectReader(file)
		if err != nil {
			return err
		}

		if mimeType.String() != constants.APPLICATION_JSON {
			return errors.New("Invalid file type")
		}
		defer file.Close()
		files = append(files, file)

		config := structs.Config{}
		fileBytes := make([]byte, fileHeader.Size)

		_, err = file.Read(fileBytes)
		if err != nil {
			return err
		}

		err = json.Unmarshal(fileBytes, &config)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command *ConfigImportCommand) Validate() structs.ValidatorError {
	vErr := structs.ValidatorError{
		Errors: make(map[string]string),
	}

	if command.FileHeader == nil {
		vErr.Errors["file"] = "File cannot be empty"
	}

	return vErr
}
