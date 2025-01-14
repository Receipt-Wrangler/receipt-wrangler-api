package commands

import (
	"encoding/json"
	"errors"
	"github.com/gabriel-vasile/mimetype"
	"io"
	"mime/multipart"
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/structs"
)

type ConfigImportCommand struct {
	File       multipart.File        `json:"file"`
	FileHeader *multipart.FileHeader `json:"fileHeader"`
	Config     structs.Config        `json:"config"`
}

func (command *ConfigImportCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseMultipartForm(constants.MultipartFormMaxSize)
	if err != nil {
		return err
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		return nil
	}

	mimeType, err := mimetype.DetectReader(file)
	if err != nil {
		return err
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	if mimeType.String() != constants.ApplicationJson {
		return errors.New("Invalid file type")
	}

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

	command.File = file
	command.FileHeader = fileHeader
	command.Config = config

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
