package middleware

import (
	"fmt"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
)

func ValidateCommentUserId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		comment := r.Context().Value("comment").(models.Comment)
		token := utils.GetJWT(r)

		fmt.Println(comment.UserId)
		fmt.Println(token.UserId)

		if *comment.UserId != token.UserId {
			utils.WriteCustomErrorResponse(w, "User not allowed to comment as another user", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
