package handlers

import (
	"net/http"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/utils"

	"github.com/go-chi/chi/v5"
)

func GetGroupsForUser(w http.ResponseWriter, r *http.Request) {
	errMsg := "Error retrieving groups."
	token := utils.GetJWT(r)

	groups, err := services.GetGroupsForUser(utils.UintToString(token.UserId))
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

	utils.SetJSONResponseHeaders(w)
	w.WriteHeader(200)
	w.Write(bytes)
}

func GetGroupById(w http.ResponseWriter, r *http.Request) {
	errMsg := "Error retrieving group."
	id := chi.URLParam(r, "groupId")

	groups, err := repositories.GetGroupById(id, true)
	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	bytes, err := utils.MarshalResponseData(groups)
	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	utils.SetJSONResponseHeaders(w)
	w.WriteHeader(200)
	w.Write(bytes)
}

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	errMsg := "Error creating group."
	token := utils.GetJWT(r)
	group := r.Context().Value("group").(models.Group)

	group, err := repositories.CreateGroup(group, token.UserId)

	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	bytes, err := utils.MarshalResponseData(group)
	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	utils.SetJSONResponseHeaders(w)
	w.WriteHeader(200)
	w.Write(bytes)
}

func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	errMsg := "Error updating group."
	group := r.Context().Value("group").(models.Group)
	groupId := chi.URLParam(r, "groupId")

	updatedGroup, err := repositories.UpdateGroup(group, groupId)

	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	bytes, err := utils.MarshalResponseData(updatedGroup)
	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	utils.SetJSONResponseHeaders(w)
	w.WriteHeader(200)
	w.Write(bytes)
}

func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	errMsg := "Error deleting group."
	id := chi.URLParam(r, "groupId")

	err := services.DeleteGroup(id)
	if err != nil {
		handler_logger.Println(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
