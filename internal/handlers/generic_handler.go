package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/simpleutils"
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
			handler_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(handler.Writer, "User is unauthorized to access entity", http.StatusForbidden)
			return
		}

		handler.GroupId = simpleutils.UintToString(receipt.GroupId)
	}

	if len(handler.GroupRole) > 0 && len(handler.GroupId) > 0 {
		var checkGroupRole bool
		if handler.GroupId == "all" && handler.AllowAllGroup {
			checkGroupRole = false
		} else {
			checkGroupRole = true
		}

		if checkGroupRole {
			token := structs.GetJWT(handler.Request)
			err := services.ValidateGroupRole(models.GroupRole(handler.GroupRole), handler.GroupId, simpleutils.UintToString(token.UserId))
			if err != nil {
				handler_logger.Print(err.Error())
				utils.WriteCustomErrorResponse(handler.Writer, "User is unauthorized to access entity", http.StatusForbidden)
				return
			}
		}
	}

	if len(handler.UserRole) > 0 {
		token := structs.GetJWT(handler.Request)
		hasUserRole := models.HasRole(handler.UserRole, token.UserRole)
		if !hasUserRole {
			handler_logger.Print("User is unauthorized to perform this action.")
			utils.WriteCustomErrorResponse(handler.Writer, "User is unauthorized to perform this action.", http.StatusForbidden)
			return
		}
	}

	errCode, err := handler.HandlerFunction(handler.Writer, handler.Request)

	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(handler.Writer, handler.ErrorMessage, errCode)
		return
	}
}
