package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func HandleRequest(handler structs.Handler) {
	errCode, err := handler.HandlerFunction(handler.Writer, handler.Request)
	if len(handler.GroupRole) > 0 && len(handler.GroupId) > 0 {
		token := utils.GetJWT(handler.Request)
		err := services.ValidateGroupRole(models.GroupRole(handler.GroupRole), handler.GroupId, simpleutils.UintToString(token.UserId))
		if err != nil {
			handler_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(handler.Writer, "User is unauthorized to access entity", http.StatusForbidden)
			return
		}
	}
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(handler.Writer, handler.ErrorMessage, errCode)
		return
	}
}
