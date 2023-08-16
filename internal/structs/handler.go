package structs

import (
	"net/http"
	"receipt-wrangler/api/internal/models"
)

type Handler struct {
	ErrorMessage    string
	Writer          http.ResponseWriter
	Request         *http.Request
	GroupRole       models.GroupRole
	GroupId					string
	HandlerFunction func(http.ResponseWriter, *http.Request) (int, error)
	ResponseType    string
}
