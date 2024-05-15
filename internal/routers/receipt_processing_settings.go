package routers

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi/v5"
	"receipt-wrangler/api/internal/handlers"
	"receipt-wrangler/api/internal/middleware"
)

func BuildReceiptProcessingSettingsRouter(tokenValidator *jwtmiddleware.JWTMiddleware) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.MoveJWTCookieToHeader, tokenValidator.CheckJWT)
	router.Get("/{id}", handlers.GetReceiptProcessingSettingsById)
	router.Post("/", handlers.CreateReceiptProcessingSettings)
	router.Post("/getPagedProcessingSettings", handlers.GetPagedReceiptProcessingSettings)
	router.Put("/{id}", handlers.UpdateReceiptProcessingSettingsById)
	router.Delete("/{id}", handlers.DeleteReceiptProcessingSettingsById)

	return router
}
