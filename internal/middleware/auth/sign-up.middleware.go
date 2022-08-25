package auth_middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"receipt-wrangler/api/internal/models"
	httpUtils "receipt-wrangler/api/internal/utils/http"
)

func SetBodyData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		bodyData, err := httpUtils.GetBodyData(w, r)

		if err != nil {
			httpUtils.WriteErrorResponse(w, err, 500)
			return
		}

		marshalErr := json.Unmarshal(bodyData, &user)
		if marshalErr != nil {
			httpUtils.WriteErrorResponse(w, marshalErr, 500)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}
