package structs

import "net/http"

type Handler struct {
	ErrorMessage    string
	Writer          http.ResponseWriter
	Request         *http.Request
	HandlerFunction func(http.ResponseWriter, *http.Request) (int, error)
	ResponseType    string
}
