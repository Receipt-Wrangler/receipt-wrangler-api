package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
)

func SetGeneralBodyData(contextKey string, dataType interface{}) (mw func(http.Handler) http.Handler) {
	mw = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyData, err := utils.GetBodyData(w, r)

			if err != nil {
				utils.WriteErrorResponse(w, err, 500)
				return
			}

			switch dataType.(type) {
			case models.Group:
				var group models.Group
				err = json.Unmarshal(bodyData, &group)
				shouldReturn := checkError(err, w)
				if shouldReturn {
					return
				}
				serveWithContext(r, w, h, contextKey, group)
			case models.Comment:
				var comment models.Comment
				err = json.Unmarshal(bodyData, &comment)
				shouldReturn := checkError(err, w)
				if shouldReturn {
					return
				}
				serveWithContext(r, w, h, contextKey, comment)

			default:
				return
			}
		})
	}
	return
}

func checkError(err error, w http.ResponseWriter) bool {
	if err != nil {
		middleware_logger.Print(err.Error())
		utils.WriteErrorResponse(w, err, 500)
		return true
	}
	return false
}

func serveWithContext(r *http.Request, w http.ResponseWriter, h http.Handler, contextKey string, content interface{}) {
	ctx := context.WithValue(r.Context(), contextKey, content)
	h.ServeHTTP(w, r.WithContext(ctx))
}
