package handlers

import (
	"log"
	"receipt-wrangler/api/internal/logging"
)

var handler_logger *log.Logger

func InitHandlerLogger() {
	handler_logger = logging.GetLogger()
}
