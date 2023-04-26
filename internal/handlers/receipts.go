package handlers

import (
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/constants"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
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
			pagedRequest := r.Context().Value("pagedRequest").(structs.PagedRequest)
			pagedData := structs.PagedData{}

			receipts, err := repositories.GetPagedReceiptsByGroupId(groupId, pagedRequest)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			anyData := make([]any, len(receipts))
			for i := 0; i < len(receipts); i++ {
				anyData[i] = receipts[i]
			}

			pagedData.Data = anyData

			count, err := repositories.GetGroupReceiptCount(groupId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

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

func CreateReceipt(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()

	errMsg := "Error creating receipts."
	bodyData := r.Context().Value("receipt").(models.Receipt)

	err := db.Model(models.Receipt{}).Select("*").Create(&bodyData).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	bytes, err := json.Marshal(bodyData)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func GetReceipt(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	var receipt models.Receipt
	errMsg := "Error retrieving receipt."

	id := chi.URLParam(r, "id")

	err := db.Model(models.Receipt{}).Where("id = ?", id).Preload(clause.Associations).Preload("Comments.Replies").Find(&receipt).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 404)
		return
	}

	bytes, err := json.Marshal(receipt)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func UpdateReceipt(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()

	errMsg := "Error updating receipt."
	id := chi.URLParam(r, "id")
	u64Id, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	bodyData := r.Context().Value("receipt").(models.Receipt)
	bodyData.ID = uint(u64Id)

	err = db.Transaction(func(tx *gorm.DB) error {
		txErr := db.Session(&gorm.Session{FullSaveAssociations: true}).Model(&bodyData).Select("*").Omit("ID").Where("id = ?", uint(u64Id)).Save(bodyData).Error
		if txErr != nil {
			handler_logger.Print(txErr.Error())
			return txErr
		}

		txErr = db.Model(&bodyData).Association("Tags").Replace(bodyData.Tags)
		if txErr != nil {
			handler_logger.Print(txErr.Error())
			return txErr
		}

		txErr = db.Model(&bodyData).Association("Categories").Replace(bodyData.Categories)
		if txErr != nil {
			handler_logger.Print(txErr.Error())
			return txErr
		}

		txErr = db.Model(&bodyData).Association("ReceiptItems").Replace(bodyData.ReceiptItems)
		if txErr != nil {
			handler_logger.Print(txErr.Error())
			return txErr
		}

		// return nil will commit the whole transaction
		return nil
	})

	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
}

func BulkResolveReceipts(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error resolving receipts",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := db.GetDB()
			bulkResolve := r.Context().Value("bulkResolve").(structs.BulkResolve)
			var receipts []models.Receipt

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
					token := utils.GetJWT(r)
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

func DeleteReceipt(w http.ResponseWriter, r *http.Request) {
	errMsg := "Error deleting receipt."
	id := chi.URLParam(r, "id")

	err := services.DeleteReceipt(id)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
}

func DuplicateReceipt(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error duplicating receipt",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := db.GetDB()
			newReceipt := models.Receipt{}

			receiptId := chi.URLParam(r, "id")
			receipt, err := repositories.GetFullyLoadedReceiptById(receiptId)

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
			for i, fileData := range newReceipt.ImageFiles {
				srcFileData := receipt.ImageFiles[i]
				srcImageBytes, err := utils.GetBytesForFileData(srcFileData)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				dstPath, err := utils.BuildFilePath(utils.UintToString(newReceipt.ID), utils.UintToString(fileData.ID), fileData.Name)
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
