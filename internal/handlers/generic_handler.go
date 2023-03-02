package handlers

import (
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

func HandleRequest(handler structs.Handler) {
	errCode, err := handler.HandlerFunction(handler.Writer, handler.Request)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(handler.Writer, handler.ErrorMessage, errCode)
		return
	}
}
