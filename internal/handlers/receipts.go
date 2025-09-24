package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"receipt-wrangler/api/internal/wranglerasynq"

	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetPagedReceiptsForGroup(w http.ResponseWriter, r *http.Request) {
	groupId := chi.URLParam(r, "groupId")
	handler := structs.Handler{
		ErrorMessage: "Error getting receipts",
		Writer:       w,
		Request:      r,
		GroupId:      groupId,
		GroupRole:    models.VIEWER,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			pagedRequest := commands.ReceiptPagedRequestCommand{}
			err := pagedRequest.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			pagedData := structs.PagedData{}
			token := structs.GetClaims(r)

			var associations []string
			if pagedRequest.FullReceipts {
				associations = constants.FULL_RECEIPT_ASSOCIATIONS
			}

			receiptRepository := repositories.NewReceiptRepository(nil)
			receipts, count, err := receiptRepository.GetPagedReceiptsByGroupId(
				token.UserId,
				groupId,
				pagedRequest,
				associations,
			)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			anyData := make([]any, len(receipts))
			for i := 0; i < len(receipts); i++ {
				anyData[i] = receipts[i]
			}

			pagedData.Data = anyData
			pagedData.TotalCount = count

			bytes, err := utils.MarshalResponseData(pagedData)
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

func GetReceiptsForGroupIds(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting receipts",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			var err error
			var receipts []models.Receipt
			var groupIds []string

			token := structs.GetClaims(r)

			r.ParseForm()

			groupIds, ok := r.Form["groupIds"]
			if !ok {
				return http.StatusInternalServerError, err
			}

			if false {

			} else {
				groupMemberRepository := repositories.NewGroupMemberRepository(nil)
				userGroupIds, err := groupMemberRepository.GetGroupIdsByUserId(utils.UintToString(token.UserId))
				if err != nil {
					return http.StatusInternalServerError, err
				}
				var userGroupIdInterfaces = make([]interface{}, len(userGroupIds))
				for i := range userGroupIds {
					userGroupIdInterfaces[i] = userGroupIds[i]
				}

				// if !utils.Contains(userGroupIdInterfaces, groupIds) {
				// 	return http.StatusForbidden, errors.New("not allowed to access group")
				// }

				receiptRepository := repositories.NewReceiptRepository(nil)
				receipts, err = receiptRepository.GetReceiptsByGroupIds(groupIds, "*", clause.Associations)
				if err != nil {
					return http.StatusInternalServerError, err
				}
			}

			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(receipts)
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

func CreateReceipt(w http.ResponseWriter, r *http.Request) {
	errMessage := "Error creating receipt"
	token := structs.GetClaims(r)

	command := commands.UpsertReceiptCommand{}
	err := command.LoadDataFromRequest(w, r)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		utils.WriteCustomErrorResponse(w, errMessage, http.StatusInternalServerError)
		return
	}
	vErrs := command.Validate(token.UserId, true)
	if len(vErrs.Errors) > 0 {
		structs.WriteValidatorErrorResponse(w, vErrs, http.StatusInternalServerError)
		return
	}

	stringId := utils.UintToString(command.GroupId)

	// TODO: Clean up to make sure group id is not an all group, and remove middleware sets and checks
	handler := structs.Handler{
		ErrorMessage: errMessage,
		Writer:       w,
		Request:      r,
		GroupId:      stringId,
		GroupRole:    models.EDITOR,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			receiptRepository := repositories.NewReceiptRepository(nil)
			createdReceipt, err := receiptRepository.CreateReceipt(command, token.UserId, true)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := json.Marshal(createdReceipt)
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

func QuickScan(w http.ResponseWriter, r *http.Request) {
	errMsg := "Error processing quick scan."
	var quickScanCommand commands.QuickScanCommand

	vErr, err := quickScanCommand.LoadDataFromRequestAndValidate(w, r)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	var groupIds = make([]string, 0)
	for i := 0; i < len(quickScanCommand.GroupIds); i++ {
		groupIds = append(groupIds, utils.UintToString(quickScanCommand.GroupIds[i]))
	}

	handler := structs.Handler{
		ErrorMessage: errMsg,
		Writer:       w,
		Request:      r,
		GroupRole:    models.EDITOR,
		GroupIds:     groupIds,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusInternalServerError)
				return http.StatusInternalServerError, errors.New("validation error")
			}

			fileRepository := repositories.NewFileRepository(nil)

			token := structs.GetClaims(r)
			for i := 0; i < len(quickScanCommand.Files); i++ {
				fileBytes := make([]byte, quickScanCommand.FileHeaders[i].Size)

				_, err := quickScanCommand.Files[i].Read(fileBytes)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				tempPath, err := fileRepository.WriteTempFile(fileBytes)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				payload := wranglerasynq.QuickScanTaskPayload{
					Token:            token,
					PaidByUserId:     quickScanCommand.PaidByUserIds[i],
					GroupId:          quickScanCommand.GroupIds[i],
					Status:           quickScanCommand.Statuses[i],
					TempPath:         tempPath,
					OriginalFileName: quickScanCommand.FileHeaders[i].Filename,
				}

				payloadBytes, err := json.Marshal(payload)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				task := asynq.NewTask(wranglerasynq.QuickScan, payloadBytes)

				_, err = wranglerasynq.EnqueueTask(task, models.QuickScanQueue)
				if err != nil {
					return http.StatusInternalServerError, err
				}
			}

			w.WriteHeader(http.StatusOK)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

// TODO: move to repository call
func GetReceipt(w http.ResponseWriter, r *http.Request) {
	receiptId := chi.URLParam(r, "id")

	handler := structs.Handler{
		ErrorMessage: "Error retrieving receipt.",
		Writer:       w,
		Request:      r,
		ReceiptId:    receiptId,
		GroupRole:    models.VIEWER,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "id")
			receiptRepository := repositories.NewReceiptRepository(nil)

			receipt, err := receiptRepository.GetFullyLoadedReceiptById(id)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := json.Marshal(receipt)
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

func UpdateReceipt(w http.ResponseWriter, r *http.Request) {
	receiptId := chi.URLParam(r, "id")

	handler := structs.Handler{
		ErrorMessage: "Error updating receipt.",
		Writer:       w,
		Request:      r,
		ReceiptId:    receiptId,
		GroupRole:    models.EDITOR,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)
			command := commands.UpsertReceiptCommand{}
			receiptRepository := repositories.NewReceiptRepository(nil)
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			vErrs := command.Validate(token.UserId, false)
			if len(vErrs.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErrs, http.StatusInternalServerError)
				return 0, nil
			}

			updatedReceipt, err := receiptRepository.UpdateReceipt(receiptId, command, token.UserId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := json.Marshal(updatedReceipt)
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

func BulkReceiptStatusUpdate(w http.ResponseWriter, r *http.Request) {
	bulkCommand := commands.BulkStatusUpdateCommand{}
	err := bulkCommand.LoadDataFromRequest(w, r)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		utils.WriteCustomErrorResponse(w, "Error resolving receipts", http.StatusInternalServerError)
		return
	}

	receiptIdStrings := make([]string, len(bulkCommand.ReceiptIds))
	for i := 0; i < len(bulkCommand.ReceiptIds); i++ {
		receiptIdStrings[i] = utils.UintToString(bulkCommand.ReceiptIds[i])
	}

	handler := structs.Handler{
		ErrorMessage: "Error resolving receipts",
		Writer:       w,
		Request:      r,
		ReceiptIds:   receiptIdStrings,
		GroupRole:    models.EDITOR,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := repositories.GetDB()
			receiptRepository := repositories.NewReceiptRepository(nil)
			var receipts []models.Receipt

			if len(bulkCommand.Status) == 0 {
				return http.StatusBadRequest, errors.New("Status required")
			}

			if !utils.Contains(models.ReceiptStatuses(), bulkCommand.Status) {
				return http.StatusBadRequest, errors.New("Invalid status")
			}

			err := db.Transaction(func(tx *gorm.DB) error {
				receiptRepository.SetTransaction(tx)
				tErr := tx.Table("receipts").Where("id IN ?", bulkCommand.ReceiptIds).Select("id", "status", "resolved_date").Find(&receipts).Error
				if tErr != nil {
					return tErr
				}

				if len(receipts) > 0 {
					for i := 0; i < len(receipts); i++ {
						receipt := receipts[i]
						receipts[i].Status = bulkCommand.Status
						tErr = tx.Model(&receipt).Updates(map[string]interface{}{"status": bulkCommand.Status}).Error
						if tErr != nil {
							return tErr
						}
					}
				}

				if len(bulkCommand.Comment) > 0 {
					token := structs.GetClaims(r)
					comments := make([]models.Comment, len(bulkCommand.ReceiptIds))

					for i := 0; i < len(bulkCommand.ReceiptIds); i++ {
						comments[i] = models.Comment{
							ReceiptId: bulkCommand.ReceiptIds[i],
							Comment:   bulkCommand.Comment,
							UserId:    &token.UserId,
						}
					}

					tErr = tx.Create(&comments).Error
					if tErr != nil {
						return tErr
					}
				}

				for i := 0; i < len(receipts); i++ {
					err = receiptRepository.AfterReceiptUpdated(&receipts[i])
					if err != nil {
						return err
					}
				}

				receiptRepository.ClearTransaction()
				return nil
			})
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = db.Table("receipts").Where("id IN ?", bulkCommand.ReceiptIds).Select("id, resolved_date, status").Find(&receipts).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(&receipts)
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

func HasAccess(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Unable to access receipt",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)
			groupService := services.NewGroupService(nil)

			receiptId := r.URL.Query().Get("receiptId")
			if len(receiptId) == 0 {
				return http.StatusBadRequest, errors.New("receiptId required")
			}

			groupRole := r.URL.Query().Get("groupRole")
			if len(groupRole) == 0 {
				return http.StatusBadRequest, errors.New("groupRole required")
			}

			receiptRepository := repositories.NewReceiptRepository(nil)
			receipt, err := receiptRepository.GetReceiptById(receiptId)
			if err != nil {
				return http.StatusBadRequest, err
			}

			validatedGroupRole, err := models.GroupRole(groupRole).Value()
			if err != nil {
				return http.StatusBadRequest, err
			}

			err = groupService.ValidateGroupRole(
				models.GroupRole(validatedGroupRole),
				fmt.Sprint(receipt.GroupId),
				fmt.Sprint(token.UserId),
			)
			if err != nil {
				return http.StatusForbidden, err
			}

			w.WriteHeader(200)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func DeleteReceipt(w http.ResponseWriter, r *http.Request) {
	receiptId := chi.URLParam(r, "id")

	handler := structs.Handler{
		ErrorMessage: "Error deleting receipt.",
		Writer:       w,
		Request:      r,
		ReceiptId:    receiptId,
		GroupRole:    models.EDITOR,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "id")
			receiptService := services.NewReceiptService(nil)

			err := receiptService.DeleteReceipt(id)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			return 0, nil
		},
	}

	HandleRequest(handler)
}

func DuplicateReceipt(w http.ResponseWriter, r *http.Request) {
	receiptId := chi.URLParam(r, "id")

	handler := structs.Handler{
		ErrorMessage: "Error duplicating receipt",
		Writer:       w,
		Request:      r,
		ReceiptId:    receiptId,
		GroupRole:    models.EDITOR,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)

			receiptService := services.NewReceiptService(nil)
			newReceipt, err := receiptService.DuplicateReceipt(token.UserId, receiptId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			responseBytes, err := json.Marshal(newReceipt)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			w.Write(responseBytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
