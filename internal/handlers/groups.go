package handlers

import (
	"errors"
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
	handler := structs.Handler{
		ErrorMessage: "Error retrieving groups.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetJWT(r)

			db := repositories.GetDB()
			db.Where("user_id = ?", token.UserId).Find(&models.Group{})

			groups, err := services.GetGroupsForUser(simpleutils.UintToString(token.UserId))
			if err != nil {
				return http.StatusInternalServerError, err
			}

			// TODO: 12/1/2023 This is a way for ensure all users have an all group. We can remove this in a few months
			hasAllGroup := false
			for i := 0; i < len(groups); i++ {
				if groups[i].IsAllGroup {
					hasAllGroup = true
				}
			}

			if !hasAllGroup {
				groupsRepository := repositories.NewGroupRepository(nil)
				_, err := groupsRepository.CreateAllGroup(token.UserId)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				groups, err = services.GetGroupsForUser(simpleutils.UintToString(token.UserId))
				if err != nil {
					return http.StatusInternalServerError, err
				}
			}

			bytes, err := utils.MarshalResponseData(groups)
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

func GetGroupById(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error retrieving group.",
		Writer:       w,
		Request:      r,
		GroupId:      chi.URLParam(r, "groupId"),
		GroupRole:    models.VIEWER,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "groupId")

			groupRepository := repositories.NewGroupRepository(nil)
			groups, err := groupRepository.GetGroupById(id, true)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(groups)
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

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error creating group",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetJWT(r)
			group := r.Context().Value("group").(models.Group)
			group.IsAllGroup = false

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

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error updating group.",
		Writer:       w,
		Request:      r,
		GroupId:      chi.URLParam(r, "groupId"),
		GroupRole:    models.OWNER,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			group := r.Context().Value("group").(models.Group)
			groupId := chi.URLParam(r, "groupId")

			groupRepository := repositories.NewGroupRepository(nil)
			uintGroupId, err := simpleutils.StringToUint(groupId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			isAllGroup, err := groupRepository.IsAllGroup(uintGroupId)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			if isAllGroup {
				return http.StatusBadRequest, errors.New("cannot update all group")
			}

			updatedGroup, err := groupRepository.UpdateGroup(group, groupId)

			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(updatedGroup)
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
	errMessage := "Error polling email(s), please review your email integration settings"

	handler := structs.Handler{
		ErrorMessage:  errMessage,
		Writer:        w,
		Request:       r,
		GroupRole:     models.VIEWER,
		GroupId:       groupId,
		AllowAllGroup: true,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			groupSettingsRepository := repositories.NewGroupSettingsRepository(nil)
			groupIdsToPoll := []string{}

			if groupId == "all" {
				token := structs.GetJWT(r)
				disabledEmailIntegrationCnt := 0
				groups, err := services.GetGroupsForUser(simpleutils.UintToString(token.UserId))
				if err != nil {
					return http.StatusInternalServerError, err
				}

				for _, group := range groups {
					if group.GroupSettings.EmailIntegrationEnabled {
						groupIdsToPoll = append(groupIdsToPoll, simpleutils.UintToString(group.ID))
					} else {
						disabledEmailIntegrationCnt += 1
					}
				}

				if disabledEmailIntegrationCnt == len(groups) {
					return http.StatusBadRequest, errors.New("email integration is not enabled for any of your groups")
				}

			} else {
				groupSettings, err := groupSettingsRepository.GetAllGroupSettings("group_id = ?", groupId)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				for _, groupSetting := range groupSettings {
					if !groupSetting.EmailIntegrationEnabled {
						return http.StatusBadRequest, errors.New("email integration is not enabled for this group")
					}
					groupIdsToPoll = append(groupIdsToPoll, simpleutils.UintToString(groupSetting.GroupId))
				}
			}

			err := email.CallClient(false, groupIdsToPoll)
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

	handler := structs.Handler{
		ErrorMessage: "Error deleting group.",
		Writer:       w,
		Request:      r,
		GroupId:      chi.URLParam(r, "groupId"),
		GroupRole:    models.OWNER,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "groupId")
			uintGroupId, err := simpleutils.StringToUint(id)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			groupRepository := repositories.NewGroupRepository(nil)
			isAllGroup, err := groupRepository.IsAllGroup(uintGroupId)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			if isAllGroup {
				return http.StatusBadRequest, errors.New("cannot delete all group")
			}

			err = services.DeleteGroup(id)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
