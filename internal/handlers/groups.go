package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"receipt-wrangler/api/internal/wranglerasynq"
	"strings"

	"github.com/hibiken/asynq"

	"github.com/go-chi/chi/v5"
)

func GetPagedGroups(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error retrieving groups.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			command := commands.PagedGroupRequestCommand{}
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErrs := command.Validate(r)
			if len(vErrs.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErrs, http.StatusBadRequest)
				return 0, nil
			}

			token := structs.GetClaims(r)
			userIdString := utils.UintToString(token.UserId)
			groupRepository := repositories.NewGroupRepository(nil)

			groups, count, err := groupRepository.GetPagedGroups(command, userIdString)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			anyData := make([]any, len(groups))
			for i := 0; i < len(groups); i++ {
				anyData[i] = groups[i]
			}

			bytes, err := utils.MarshalResponseData(structs.PagedData{
				TotalCount: count,
				Data:       anyData,
			})
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

func GetGroupsForUser(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error retrieving groups.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)
			groupService := services.NewGroupService(nil)

			groups, err := groupService.GetGroupsForUser(utils.UintToString(token.UserId))
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
				groupService := services.NewGroupService(nil)

				_, err := groupsRepository.CreateAllGroup(token.UserId)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				groups, err = groupService.GetGroupsForUser(utils.UintToString(token.UserId))
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
		OrUserRole:   models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "groupId")

			groupRepository := repositories.NewGroupRepository(nil)
			groups, err := groupRepository.GetGroupById(id, true, true, true)
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
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			command := commands.UpsertGroupCommand{}
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErrs := command.Validate(true)
			if len(vErrs.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErrs, http.StatusBadRequest)
				return 0, nil
			}

			token := structs.GetClaims(r)

			command.IsAllGroup = false
			groupRepository := repositories.NewGroupRepository(nil)
			group, err := groupRepository.CreateGroup(command, token.UserId)

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

// TODO: move hooks, and update swagger to take command
func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error updating group.",
		Writer:       w,
		Request:      r,
		GroupId:      chi.URLParam(r, "groupId"),
		GroupRole:    models.OWNER,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			command := commands.UpsertGroupCommand{}
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErrs := command.Validate(true)
			if len(vErrs.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErrs, http.StatusBadRequest)
				return 0, nil
			}
			groupId := chi.URLParam(r, "groupId")

			groupRepository := repositories.NewGroupRepository(nil)
			uintGroupId, err := utils.StringToUint(groupId)
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

			updatedGroup, err := groupRepository.UpdateGroup(command, groupId)

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
		ResponseType: constants.ApplicationJson,
		UserRole:     models.ADMIN,
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

func UpdateGroupReceiptSettings(w http.ResponseWriter, r *http.Request) {
	groupId := chi.URLParam(r, "groupId")

	handler := structs.Handler{
		ErrorMessage: "Error updating group receipt settings",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		GroupId:      groupId,
		GroupRole:    models.OWNER,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			command := commands.UpdateGroupReceiptSettingsCommand{}
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			groupReceiptSettingsRepository := repositories.NewGroupReceiptSettingsRepository(nil)
			updatedGroupReceiptSettings, err := groupReceiptSettingsRepository.UpdateGroupReceiptSettings(groupId, command)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(updatedGroupReceiptSettings)
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

func PollGroupEmail(w http.ResponseWriter, r *http.Request) {
	groupId := chi.URLParam(r, "groupId")
	errMessage := "Error polling email(s), please review your email integration settings"

	handler := structs.Handler{
		ErrorMessage: errMessage,
		Writer:       w,
		Request:      r,
		GroupRole:    models.VIEWER,
		GroupId:      groupId,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			groupSettingsRepository := repositories.NewGroupSettingsRepository(nil)
			groupRepository := repositories.NewGroupRepository(nil)

			groupIdsToPoll := []string{}

			uintGroupId, err := utils.StringToUint(groupId)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			isAllGroup, err := groupRepository.IsAllGroup(uintGroupId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if isAllGroup {
				token := structs.GetClaims(r)
				disabledEmailIntegrationCnt := 0
				groupService := services.NewGroupService(nil)

				groups, err := groupService.GetGroupsForUser(utils.UintToString(token.UserId))
				if err != nil {
					return http.StatusInternalServerError, err
				}

				for _, group := range groups {
					if group.GroupSettings.EmailIntegrationEnabled {
						groupIdsToPoll = append(groupIdsToPoll, utils.UintToString(group.ID))
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
					groupIdsToPoll = append(groupIdsToPoll, utils.UintToString(groupSetting.GroupId))
				}
			}

			// TODO: Enqueue instead
			taskPayload := wranglerasynq.EmailPollTaskPayload{
				GroupIds: groupIdsToPoll,
			}

			payloadBytes, err := json.Marshal(taskPayload)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			task := asynq.NewTask(wranglerasynq.EmailPoll, payloadBytes)
			_, err = wranglerasynq.EnqueueTask(task, models.EmailPollingQueue)
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
			groupService := services.NewGroupService(nil)

			uintGroupId, err := utils.StringToUint(id)
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

			err = groupService.DeleteGroup(id, false)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func GetOcrTextForGroup(w http.ResponseWriter, r *http.Request) {
	groupId := chi.URLParam(r, "groupId")
	handler := structs.Handler{
		ErrorMessage: "Error getting ocr text.",
		Writer:       w,
		Request:      r,
		GroupId:      groupId,
		GroupRole:    models.OWNER,
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationZip,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)
			zipFilename := "results.zip"

			ocrResults, err := services.ReadAllReceiptImagesForGroup(groupId, utils.UintToString(token.UserId))
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

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
