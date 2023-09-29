package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/email"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

func GetGroupsForUser(w http.ResponseWriter, r *http.Request) {
	errMsg := "Error retrieving groups."
	token := structs.GetJWT(r)

	groups, err := services.GetGroupsForUser(simpleutils.UintToString(token.UserId))
	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	bytes, err := utils.MarshalResponseData(groups)
	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	utils.SetJSONResponseHeaders(w)
	w.WriteHeader(200)
	w.Write(bytes)
}

func GetGroupById(w http.ResponseWriter, r *http.Request) {
	errMsg := "Error retrieving group."
	id := chi.URLParam(r, "groupId")

	groupRepository := repositories.NewGroupRepository(nil)
	groups, err := groupRepository.GetGroupById(id, true)
	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	bytes, err := utils.MarshalResponseData(groups)
	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	utils.SetJSONResponseHeaders(w)
	w.WriteHeader(200)
	w.Write(bytes)
}

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error creating group",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetJWT(r)
			group := r.Context().Value("group").(models.Group)

			groupRepository := repositories.NewGroupRepository(nil)
			group, err := groupRepository.CreateGroup(group, token.UserId)

			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(group)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			fileRepository := repositories.NewFileRepository(nil)
			groupPath, err := fileRepository.BuildGroupPath(group.ID, "")
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = utils.MakeDirectory(groupPath)
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

func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	errMsg := "Error updating group."
	group := r.Context().Value("group").(models.Group)
	groupId := chi.URLParam(r, "groupId")

	groupRepository := repositories.NewGroupRepository(nil)
	updatedGroup, err := groupRepository.UpdateGroup(group, groupId)

	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	bytes, err := utils.MarshalResponseData(updatedGroup)
	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	utils.SetJSONResponseHeaders(w)
	w.WriteHeader(200)
	w.Write(bytes)
}

func UpdateGroupSettings(w http.ResponseWriter, r *http.Request) {
	groupId := chi.URLParam(r, "groupId")

	handler := structs.Handler{
		ErrorMessage: "Error updating group settings",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		GroupRole:    models.OWNER,
		GroupId:      groupId,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			command := commands.UpdateGroupSettingsCommand{}
			vErr, err := command.LoadDataFromRequestAndValidate(w, r)
			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusBadRequest)
				return 0, nil
			}
			if err != nil {
				return http.StatusInternalServerError, err
			}

			groupSettingsRepository := repositories.NewGroupSettingsRepository(nil)
			updatedGroupSettings, err := groupSettingsRepository.UpdateGroupSettings(groupId, command)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(updatedGroupSettings)
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

func PollGroupEmail(w http.ResponseWriter, r *http.Request) {
	groupId := chi.URLParam(r, "groupId")

	handler := structs.Handler{
		ErrorMessage: "Error polling emails",
		Writer:       w,
		Request:      r,
		GroupRole:    models.VIEWER,
		GroupId:      groupId,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			err := email.CallClient(groupId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			return 0, nil
		},
	}

	HandleRequest(handler)
}

func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	errMsg := "Error deleting group."
	id := chi.URLParam(r, "groupId")

	err := services.DeleteGroup(id)
	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
