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
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jinzhu/copier"
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
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			pagedRequest := commands.ReceiptPagedRequestCommand{}
			err := pagedRequest.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			pagedData := structs.PagedData{}
			token := structs.GetJWT(r)

			receiptRepository := repositories.NewReceiptRepository(nil)
			receipts, count, err := receiptRepository.GetPagedReceiptsByGroupId(token.UserId, groupId, pagedRequest)
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
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			var err error
			var receipts []models.Receipt
			var groupIds []string

			token := structs.GetJWT(r)

			r.ParseForm()

			groupIds, ok := r.Form["groupIds"]
			if !ok {
				return http.StatusInternalServerError, err
			}

			if false {

			} else {
				groupMemberRepository := repositories.NewGroupMemberRepository(nil)
				userGroupIds, err := groupMemberRepository.GetGroupIdsByUserId(simpleutils.UintToString(token.UserId))
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
	token := structs.GetJWT(r)

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

	stringId := simpleutils.UintToString(command.GroupId)

	// TODO: Clean up to make sure group id is not an all group, and remove middleware sets and checks
	handler := structs.Handler{
		ErrorMessage: errMessage,
		Writer:       w,
		Request:      r,
		GroupId:      stringId,
		GroupRole:    models.EDITOR,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			receiptRepository := repositories.NewReceiptRepository(nil)
			createdReceipt, err := receiptRepository.CreateReceipt(command, token.UserId)
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
		groupIds = append(groupIds, simpleutils.UintToString(quickScanCommand.GroupIds[i]))
	}

	handler := structs.Handler{
		ErrorMessage: errMsg,
		Writer:       w,
		Request:      r,
		GroupRole:    models.EDITOR,
		GroupIds:     groupIds,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusInternalServerError)
				return http.StatusInternalServerError, errors.New("validation error")
			}

			systemSettingsRepository := repositories.NewSystemSettingsRepository(nil)
			systemSettings, err := systemSettingsRepository.GetSystemSettings()
			if err != nil {
				return http.StatusInternalServerError, err
			}

			var wg sync.WaitGroup

			semaphore := make(chan int, systemSettings.NumWorkers)
			results := make(chan models.Receipt, len(quickScanCommand.Files))
			createdReceipts := make([]models.Receipt, 0)
			receiptService := services.NewReceiptService(nil)

			token := structs.GetJWT(r)
			for i := 0; i < len(quickScanCommand.Files); i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					semaphore <- index

					createdReceipt, err := receiptService.QuickScan(
						token,
						quickScanCommand.Files[index],
						quickScanCommand.FileHeaders[index],
						quickScanCommand.PaidByUserIds[index],
						quickScanCommand.GroupIds[index],
						quickScanCommand.Statuses[index],
					)
					results <- createdReceipt
					if err != nil {
						results <- models.Receipt{}
					}

					<-semaphore
				}(i)
			}

			go func() {
				wg.Wait()
				close(results)
			}()

			for r := range results {
				createdReceipts = append(createdReceipts, r)
			}

			bytes, err := utils.MarshalResponseData(createdReceipts)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)

			_, err = w.Write(bytes)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func GetReceipt(w http.ResponseWriter, r *http.Request) {
	receiptId := chi.URLParam(r, "id")

	handler := structs.Handler{
		ErrorMessage: "Error retrieving receipt.",
		Writer:       w,
		Request:      r,
		ReceiptId:    receiptId,
		GroupRole:    models.VIEWER,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := repositories.GetDB()
			var receipt models.Receipt
			id := chi.URLParam(r, "id")

			err := db.Model(models.Receipt{}).Where("id = ?", id).Preload(clause.Associations).Preload("Comments.Replies").Find(&receipt).Error
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
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetJWT(r)
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

			updatedReceipt, err := receiptRepository.UpdateReceipt(receiptId, command)
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
		receiptIdStrings[i] = simpleutils.UintToString(bulkCommand.ReceiptIds[i])
	}

	handler := structs.Handler{
		ErrorMessage: "Error resolving receipts",
		Writer:       w,
		Request:      r,
		ReceiptIds:   receiptIdStrings,
		GroupRole:    models.EDITOR,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := repositories.GetDB()
			receiptRepository := repositories.NewReceiptRepository(nil)
			var receipts []models.Receipt

			if len(bulkCommand.Status) == 0 {
				return http.StatusBadRequest, errors.New("Status required")
			}

			if !utils.Contains(constants.ReceiptStatuses(), bulkCommand.Status) {
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
					token := structs.GetJWT(r)
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
		ErrorMessage: "Insufficient permissions.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetJWT(r)
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
		ResponseType: constants.APPLICATION_JSON,
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
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := repositories.GetDB()
			newReceipt := models.Receipt{}
			userId := structs.GetJWT(r).UserId

			receiptId := chi.URLParam(r, "id")
			receiptRepository := repositories.NewReceiptRepository(nil)
			receipt, err := receiptRepository.GetFullyLoadedReceiptById(receiptId)

			if err != nil {
				return http.StatusInternalServerError, err
			}

			copier.Copy(&newReceipt, receipt)

			newReceipt.ID = 0
			newReceipt.Name = newReceipt.Name + " duplicate"
			newReceipt.ImageFiles = make([]models.FileData, 0)
			newReceipt.ReceiptItems = make([]models.Item, 0)
			newReceipt.Comments = make([]models.Comment, 0)
			newReceipt.CreatedAt = time.Now()
			newReceipt.UpdatedAt = time.Now()
			newReceipt.CreatedBy = &userId

			// Remove fks from any related data
			for _, fileData := range receipt.ImageFiles {
				var newFileData models.FileData
				copier.Copy(&newFileData, fileData)

				newFileData.ID = 0
				newFileData.ReceiptId = 0
				newFileData.Receipt = models.Receipt{}
				newReceipt.ImageFiles = append(newReceipt.ImageFiles, newFileData)
			}

			// Copy items
			for _, item := range receipt.ReceiptItems {
				var newItem models.Item
				copier.Copy(&newItem, item)

				newItem.ID = 0
				newItem.ReceiptId = 0
				newItem.Receipt = models.Receipt{}
				newReceipt.ReceiptItems = append(newReceipt.ReceiptItems, newItem)
			}

			// Copy comments
			for _, comment := range receipt.Comments {
				var newComment models.Comment
				copier.Copy(&newComment, comment)

				newComment.ID = 0
				newComment.ReceiptId = 0
				newComment.Receipt = models.Receipt{}
				newReceipt.Comments = append(newReceipt.Comments, newComment)
			}

			err = db.Create(&newReceipt).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			// Copy receipt images
			fileRepository := repositories.NewFileRepository(nil)
			for i, fileData := range newReceipt.ImageFiles {
				srcFileData := receipt.ImageFiles[i]
				srcImageBytes, err := fileRepository.GetBytesForFileData(srcFileData)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				dstPath, err := fileRepository.BuildFilePath(
					simpleutils.UintToString(newReceipt.ID),
					simpleutils.UintToString(fileData.ID),
					fileData.Name,
				)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				err = utils.WriteFile(dstPath, srcImageBytes)
				if err != nil {
					return http.StatusInternalServerError, err
				}
			}

			responseBytes, err := utils.MarshalResponseData(newReceipt)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			// before we can create again we need to clear out all the fks

			w.WriteHeader(http.StatusOK)
			w.Write(responseBytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
