package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
)

func SetReceiptImageData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// TODO: Come up with a better way to handdle this
			var receipt models.FileData
			bodyData, err := utils.GetBodyData(w, r)

			if err != nil {
				utils.WriteErrorResponse(w, err, 500)
				return
			}

			marshalErr := json.Unmarshal(bodyData, &receipt)
			if marshalErr != nil {
				utils.WriteErrorResponse(w, marshalErr, 500)
				return
			}
			ctx := context.WithValue(r.Context(), "fileData", receipt)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		next.ServeHTTP(w, r)
		return
	})
}
