package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func HandleRequest(handler structs.Handler) {
	if len(handler.ResponseType) > 0 {
		handler.Writer.Header().Set("Content-Type", handler.ResponseType)
	}

	if len(handler.ReceiptId) > 0 {
		var receipt models.Receipt
		db := repositories.GetDB()
		err := db.Model(models.Receipt{}).Where("id = ?", handler.ReceiptId).Select("group_id").First(&receipt).Error
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
			utils.WriteCustomErrorResponse(handler.Writer, "User is unauthorized to access entity", http.StatusForbidden)
			return
		}

		handler.GroupId = utils.UintToString(receipt.GroupId)
	}

	if len(handler.ReceiptIds) > 0 {
		var receipts []models.Receipt
		db := repositories.GetDB()
		err := db.Model(models.Receipt{}).Where("id IN (?)", handler.ReceiptIds).Select("group_id").Find(&receipts).Error
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
			utils.WriteCustomErrorResponse(handler.Writer, "User is unauthorized to access entity", http.StatusForbidden)
			return
		}

		for _, receipt := range receipts {
			handler.GroupIds = append(handler.GroupIds, utils.UintToString(receipt.GroupId))
		}
	}

	if len(handler.GroupRole) > 0 && len(handler.GroupId) == 0 && len(handler.GroupIds) == 0 {
		utils.WriteCustomErrorResponse(handler.Writer, "Group ID is required to validate group role", http.StatusForbidden)
		return
	}

	if len(handler.GroupRole) > 0 && len(handler.GroupId) > 0 {
		groupService := services.NewGroupService(nil)
		token := structs.GetClaims(handler.Request)
		err := groupService.ValidateGroupRole(models.GroupRole(handler.GroupRole), handler.GroupId, utils.UintToString(token.UserId))
		hasOrUserRole := false

		if len(handler.OrUserRole) > 0 {
			hasOrUserRole = models.HasRole(handler.OrUserRole, token.UserRole)
		}

		if err != nil && !hasOrUserRole {
			logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
			utils.WriteCustomErrorResponse(handler.Writer, "User is unauthorized to access entity", http.StatusForbidden)
			return
		}
	}

	if len(handler.GroupRole) > 0 && len(handler.GroupIds) == 0 && len(handler.GroupId) == 0 {
		utils.WriteCustomErrorResponse(handler.Writer, "Group IDs are required to validate group role", http.StatusForbidden)
		return
	}

	if len(handler.GroupRole) > 0 && len(handler.GroupIds) > 0 {
		groupService := services.NewGroupService(nil)
		token := structs.GetClaims(handler.Request)

		for _, groupId := range handler.GroupIds {
			err := groupService.ValidateGroupRole(models.GroupRole(handler.GroupRole), groupId, utils.UintToString(token.UserId))
			if err != nil {
				logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
				utils.WriteCustomErrorResponse(handler.Writer, "User is unauthorized to access entity", http.StatusForbidden)
				return
			}
		}
	}

	if len(handler.UserRole) > 0 {
		token := structs.GetClaims(handler.Request)
		hasUserRole := models.HasRole(handler.UserRole, token.UserRole)
		if !hasUserRole {
			logging.LogStd(logging.LOG_LEVEL_ERROR, "User is unauthorized to perform this action.")
			utils.WriteCustomErrorResponse(handler.Writer, "User is unauthorized to perform this action.", http.StatusForbidden)
			return
		}
	}

	errCode, err := handler.HandlerFunction(handler.Writer, handler.Request)

	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		utils.WriteCustomErrorResponse(handler.Writer, handler.ErrorMessage, errCode)
		return
	}
}
