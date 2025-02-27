package handlers

import (
	"fmt"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

func GetAllTags(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error retrieving tags",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
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

func CreateTag(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error creating tag",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			tag := commands.UpsertTagCommand{}
			err := tag.LoadDataFromRequest(w, r)
			vErrs := tag.Validate()
			if len(vErrs.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErrs, http.StatusBadRequest)
				return 0, nil
			}
			if err != nil {
				return http.StatusInternalServerError, err
			}

			tagRepository := repositories.NewTagsRepository(nil)
			createdTag, err := tagRepository.CreateTag(tag)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(&createdTag)
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
		ResponseType: constants.ApplicationJson,
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

func UpdateTag(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error updating tag",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "tagId")

			tagUpsertCommand := commands.UpsertTagCommand{}
			err := tagUpsertCommand.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			tagRepository := repositories.NewTagsRepository(nil)
			updatedTag, err := tagRepository.UpdateTag(id, tagUpsertCommand)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(updatedTag)
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

func DeleteTag(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error deleting tag",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "tagId")

			uintId, err := utils.StringToUint(id)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			tagRepository := repositories.NewTagsRepository(nil)
			err = tagRepository.DeleteTag(uintId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(200)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func GetTagNameCount(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting tag count",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.TextPlain,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			tagRepository := repositories.NewTagsRepository(nil)
			tagName := chi.URLParam(r, "tagName")
			count, err := tagRepository.GetCount("tags", fmt.Sprintf("name = '%s'", tagName))
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(count)
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
