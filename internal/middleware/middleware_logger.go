package middleware

import (
	"log"
	"receipt-wrangler/api/internal/logging"
)

var middleware_logger *log.Logger

func InitHandlerLogger() {
	middleware_logger = logging.GetLogger()
}
