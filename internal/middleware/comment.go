package middleware

import (
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

func ValidateComment(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		comment := r.Context().Value("comment").(models.Comment)
		token := utils.GetJWT(r)

		vErr := structs.ValidatorError{
			Errors: make(map[string]string),
		}

		if *comment.UserId != token.UserId {
			utils.WriteCustomErrorResponse(w, "User not allowed to comment as another user", http.StatusForbidden)
			return
		}

		if comment.ReceiptId == 0 {
			vErr.Errors["receiptId"] = "Receipt Id must be assigned"
		}

		if len(vErr.Errors) > 0 {
			utils.WriteValidatorErrorResponse(w, vErr, http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CanDeleteComment(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var comment models.Comment

		db := repositories.GetDB()
		commentId := chi.URLParam(r, "commentId")
		token := utils.GetJWT(r)

		err := db.Model(models.Comment{}).Where("id = ?", commentId).Select("user_id").Find(&comment).Error
		if err != nil {
			utils.WriteCustomErrorResponse(w, "Error deleting comment", http.StatusInternalServerError)
			return
		}

		if *comment.UserId != token.UserId {
			utils.WriteCustomErrorResponse(w, "Not allowed to delete another user's comment", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
