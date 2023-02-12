package handlers

import (
	"encoding/json"
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

type ItemView struct {
	ItemId       uint `json:"id"`
	ReceiptId    uint
	PaidByUserId uint
	ItemAmount   decimal.Decimal
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	errMsg := "Error retrieving users."
	var users []structs.APIUser

	err := db.Model(models.User{}).Find(&users).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	bytes, err := json.Marshal(users)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	var apiUser structs.APIUser
	bodyData := r.Context().Value("user").(models.User)
	errMsg := "Error creating user."
	createdUser, err := repositories.CreateUser(bodyData)

	if err != nil {
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	err = db.Model(models.User{}).Where("id = ?", createdUser.ID).Find(&apiUser).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	bytes, err := utils.MarshalResponseData(apiUser)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	errMsg := "Error updating user."
	id := chi.URLParam(r, "id")

	u64Id, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	bodyData := r.Context().Value("user").(models.User)
	bodyData.ID = uint(u64Id)

	err = db.Model(&bodyData).Select("username", "display_name", "user_role").Updates(&bodyData).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetAmountOwedForUser(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	errMsg := "Error calculating amount owed."
	var items []ItemView
	total := decimal.NewFromInt(0)
	token := utils.GetJWT(r)
	id := token.UserId
	resultMap := make(map[uint]decimal.Decimal)

	err := db.Table("items").Select("items.id as item_id, items.receipt_id as receipt_id, items.amount as item_amount, items.charged_to_user_id, receipts.id, receipts.is_resolved, receipts.paid_by_user_id").Joins("inner join receipts on receipts.id=items.receipt_id").Where("items.charged_to_user_id=? AND receipts.paid_by_user_id !=? AND receipts.is_resolved=?", id, id, false).Scan(&items).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	for i := 1; i < len(items); i++ {
		item := items[i]
		total = total.Add(item.ItemAmount)
		amount, ok := resultMap[item.PaidByUserId]
		if ok {
			resultMap[item.PaidByUserId] = amount.Add(item.ItemAmount)
		} else {
			resultMap[item.PaidByUserId] = item.ItemAmount
		}
	}
	resultMap[0] = total

	bytes, err := json.Marshal(resultMap)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, 500)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func GetUsernameCount(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	errMsg := "Error getting username count."
	username := chi.URLParam(r, "username")
	var result int64

	err := db.Model(models.User{}).Where("username = ?", username).Count(&result).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	bytes, err := utils.MarshalResponseData(result)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(bytes)
}

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	errMsg := "Error resetting password."
	id := chi.URLParam(r, "id")
	resetPasswordData := r.Context().Value("reset_password").(structs.ResetPassword)

	hashedPassword, err := utils.HashPassword(resetPasswordData.Password)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	err = db.Model(models.User{}).Where("id = ?", id).UpdateColumn("password", hashedPassword).Error
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteCustomErrorResponse(w, errMsg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}
