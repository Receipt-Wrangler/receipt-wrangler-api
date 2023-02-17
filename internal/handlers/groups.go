package handlers

import (
	"fmt"
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/utils"
)

func GetGroupsForUser(w http.ResponseWriter, r *http.Request) {
	errMsg := "Error retrieving groups."
	token := utils.GetJWT(r)

	groups, err := services.GetGroupsForUser(token.UserId)
	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	bytes, err := utils.MarshalResponseData(groups)
	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	//errMsg := "Error retrieving groups."
	//token := utils.GetJWT(r)
	group := r.Context().Value("group").(models.Group)
	fmt.Println(group)

	w.WriteHeader(200)
	//w.Write(bytes)
}
