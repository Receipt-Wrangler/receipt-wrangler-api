package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
)

func GetAboutData(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error deleting comment",
		Writer:       w,
		Request:      r,
		ResponseType: "",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			commentId := chi.URLParam(r, "commentId")
			commentRepository := repositories.NewCommentRepository(nil)
			token := structs.GetJWT(r)

			err := commentRepository.DeleteComment(commentId, token.UserId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			return 0, nil
		},
	}

	HandleRequest(handler)
}
