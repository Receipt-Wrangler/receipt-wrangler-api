package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func GetAllTags(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error retrieving tags",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			tagsRepository := repositories.NewTagsRepository(nil)
			tags, err := tagsRepository.GetAllTags("*")
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(&tags)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func GetPagedTags(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error retrieving tags",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			pagedData := structs.PagedData{}
			pagedRequestCommand := commands.PagedRequestCommand{}
			err := pagedRequestCommand.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			tagsRepository := repositories.NewTagsRepository(nil)
			tags, err := tagsRepository.GetAllPagedTags(pagedRequestCommand)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			anyData := make([]any, len(tags))
			for i := 0; i < len(tags); i++ {
				anyData[i] = tags[i]
			}

			pagedData.Data = anyData
			pagedData.TotalCount = int64(len(anyData))

			bytes, err := utils.MarshalResponseData(pagedData)
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
