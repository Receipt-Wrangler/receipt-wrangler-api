package middleware

import (
	"log"
	"receipt-wrangler/api/internal/logging"
)

var middleware_logger *log.Logger

func InitMiddlewareLogger() {
	middleware_logger = logging.GetLogger()
}
