package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetPagedReceiptsForGroup(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting receipts",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			groupId := chi.URLParam(r, "groupId")
			pagedRequest := r.Context().Value("pagedRequest").(commands.ReceiptPagedRequestCommand)
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

			w.WriteHeader(200)
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
	handler := structs.Handler{
		ErrorMessage: "Error creating receipt.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetJWT(r)
			receiptRepository := repositories.NewReceiptRepository(nil)

			bodyData := r.Context().Value("receipt").(models.Receipt)
			createdReceipt, err := receiptRepository.CreateReceipt(bodyData, token.UserId)
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
	groupId := ""
	handler := structs.Handler{
		ErrorMessage: "Error processing quick scan.",
		Writer:       w,
		Request:      r,
		GroupRole:    models.EDITOR,
		GroupId:      groupId,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			var quickScanCommand commands.QuickScanCommand
			var createdReceipt models.Receipt
			db := repositories.GetDB()
			fileRepository := repositories.NewFileRepository(nil)

			vErr, err := quickScanCommand.LoadDataFromRequestAndValidate(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			_, err = fileRepository.ValidateFileType(models.FileData{ImageData: quickScanCommand.ImageData})
			if err != nil {
				return http.StatusInternalServerError, err
			}

			groupId = simpleutils.UintToString(quickScanCommand.GroupId)

			if len(vErr.Errors) > 0 {
				structs.WriteValidatorErrorResponse(w, vErr, http.StatusInternalServerError)
				return http.StatusInternalServerError, errors.New("validation error")
			}

			magicFillCommand := commands.MagicFillCommand{
				ImageData: quickScanCommand.ImageData,
				Filename:  quickScanCommand.Name,
			}

			token := structs.GetJWT(r)
			receiptRepository := repositories.NewReceiptRepository(nil)
			receiptImageRepository := repositories.NewReceiptImageRepository(nil)

			receipt, err := services.MagicFillFromImage(magicFillCommand)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			receipt.PaidByUserID = quickScanCommand.PaidByUserId
			receipt.Status = models.ReceiptStatus(quickScanCommand.Status)
			receipt.GroupId = quickScanCommand.GroupId

			db.Transaction(func(tx *gorm.DB) error {
				receiptRepository.SetTransaction(tx)
				receiptImageRepository.SetTransaction(tx)

				createdReceipt, err = receiptRepository.CreateReceipt(receipt, token.UserId)
				if err != nil {
					return err
				}

				quickScanCommand.FileData.ReceiptId = createdReceipt.ID
				_, err := receiptImageRepository.CreateReceiptImage(quickScanCommand.FileData)
				if err != nil {
					return err
				}

				return nil
			})

			bytes, err := utils.MarshalResponseData(createdReceipt)
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

func GetReceipt(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error retrieving receipt.",
		Writer:       w,
		Request:      r,
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
	handler := structs.Handler{
		ErrorMessage: "Error updating receipt.",
		Writer:       w,
		Request:      r,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := repositories.GetDB()

			id := chi.URLParam(r, "id")
			u64Id, err := strconv.ParseUint(id, 10, 32)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bodyData := r.Context().Value("receipt").(models.Receipt)
			bodyData.ID = uint(u64Id)

			err = db.Transaction(func(tx *gorm.DB) error {
				txErr := tx.Session(&gorm.Session{FullSaveAssociations: true}).Model(&bodyData).Select("*").Omit("ID", "created_by", "updated_at", "created_at").Where("id = ?", uint(u64Id)).Save(bodyData).Error
				if txErr != nil {
					handler_logger.Print(txErr.Error())
					return txErr
				}

				txErr = tx.Model(&bodyData).Association("Tags").Replace(bodyData.Tags)
				if txErr != nil {
					handler_logger.Print(txErr.Error())
					return txErr
				}

				txErr = tx.Model(&bodyData).Association("Categories").Replace(bodyData.Categories)
				if txErr != nil {
					handler_logger.Print(txErr.Error())
					return txErr
				}

				txErr = tx.Model(&bodyData).Association("ReceiptItems").Replace(bodyData.ReceiptItems)
				if txErr != nil {
					handler_logger.Print(txErr.Error())
					return txErr
				}

				return nil
			})

			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			return 0, nil
		},
	}

	HandleRequest(handler)
}

func BulkReceiptStatusUpdate(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error resolving receipts",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := repositories.GetDB()
			bulkResolve := r.Context().Value("BulkStatusUpdateCommand").(commands.BulkStatusUpdateCommand)
			var receipts []models.Receipt

			if len(bulkResolve.Status) == 0 {
				return http.StatusBadRequest, errors.New("Status required")
			}

			if !utils.Contains(constants.ReceiptStatuses(), bulkResolve.Status) {
				return http.StatusBadRequest, errors.New("Invalid status")
			}

			err := db.Transaction(func(tx *gorm.DB) error {
				tErr := tx.Table("receipts").Where("id IN ?", bulkResolve.ReceiptIds).Select("id", "status", "resolved_date").Find(&receipts).Error
				if tErr != nil {
					return tErr
				}

				if len(receipts) > 0 {
					for i := 0; i < len(receipts); i++ {
						receipt := receipts[i]
						tErr = tx.Model(&receipt).Updates(map[string]interface{}{"status": bulkResolve.Status}).Error
						if tErr != nil {
							return tErr
						}
					}
				}

				if len(bulkResolve.Comment) > 0 {
					token := structs.GetJWT(r)
					comments := make([]models.Comment, len(bulkResolve.ReceiptIds))

					for i := 0; i < len(bulkResolve.ReceiptIds); i++ {
						comments[i] = models.Comment{ReceiptId: bulkResolve.ReceiptIds[i], Comment: bulkResolve.Comment, UserId: &token.UserId}
					}

					tErr = tx.Create(&comments).Error
					if tErr != nil {
						return tErr
					}
				}

				return nil
			})
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = db.Table("receipts").Where("id IN ?", bulkResolve.ReceiptIds).Select("id, resolved_date, status").Find(&receipts).Error
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

			err = services.ValidateGroupRole(models.GroupRole(validatedGroupRole), fmt.Sprint(receipt.GroupId), fmt.Sprint(token.UserId))
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
	handler := structs.Handler{
		ErrorMessage: "Error deleting receipt.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "id")

			err := services.DeleteReceipt(id)
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
	handler := structs.Handler{
		ErrorMessage: "Error duplicating receipt",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := repositories.GetDB()
			newReceipt := models.Receipt{}

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

				dstPath, err := fileRepository.BuildFilePath(simpleutils.UintToString(newReceipt.ID), simpleutils.UintToString(fileData.ID), fileData.Name)
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
