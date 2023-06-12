package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/simpleutils"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

type ItemView struct {
	ItemId          uint `json:"id"`
	ReceiptId       uint
	PaidByUserId    uint
	ChargedToUserId uint
	ItemAmount      decimal.Decimal
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
	var itemsOwed []ItemView
	var itemsOthersOwe []ItemView
	total := decimal.NewFromInt(0)
	token := utils.GetJWT(r)
	id := token.UserId
	resultMap := make(map[uint]decimal.Decimal)

	groupId := chi.URLParam(r, "groupId")

	if groupId == "all" {
		groupIds, err := repositories.GetGroupIdsByUserId(simpleutils.UintToString(token.UserId))
		if err != nil {
			handler_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return
		}

		err = db.Table("items").Select("items.id as item_id, items.receipt_id as receipt_id, items.amount as item_amount, items.charged_to_user_id, receipts.id, items.status, receipts.paid_by_user_id").Joins("inner join receipts on receipts.id=items.receipt_id").Where("items.charged_to_user_id=? AND receipts.paid_by_user_id !=? AND receipts.group_id IN ? AND items.status=?", id, id, groupIds, models.ITEM_OPEN).Scan(&itemsOwed).Error
		if err != nil {
			handler_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return
		}

		err = db.Table("items").Select("items.id as item_id, items.receipt_id as receipt_id, items.amount as item_amount, items.charged_to_user_id, receipts.id, items.status, receipts.paid_by_user_id").Joins("inner join receipts on receipts.id=items.receipt_id").Where("items.charged_to_user_id !=? AND receipts.paid_by_user_id =? AND receipts.group_id IN ? AND items.status=?", id, id, groupIds, models.ITEM_OPEN).Scan(&itemsOthersOwe).Error
		if err != nil {
			handler_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return
		}
	} else {
		err := db.Table("items").Select("items.id as item_id, items.receipt_id as receipt_id, items.amount as item_amount, items.charged_to_user_id, receipts.id, items.status, receipts.paid_by_user_id").Joins("inner join receipts on receipts.id=items.receipt_id").Where("items.charged_to_user_id=? AND receipts.paid_by_user_id !=? AND receipts.group_id =? AND items.status=?", id, id, groupId, models.ITEM_OPEN).Scan(&itemsOwed).Error
		if err != nil {
			handler_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return
		}

		err = db.Table("items").Select("items.id as item_id, items.receipt_id as receipt_id, items.amount as item_amount, items.charged_to_user_id, receipts.id, items.status, receipts.paid_by_user_id").Joins("inner join receipts on receipts.id=items.receipt_id").Where("items.charged_to_user_id !=? AND receipts.paid_by_user_id =? AND receipts.group_id =? AND items.status=?", id, id, groupId, models.ITEM_OPEN).Scan(&itemsOthersOwe).Error
		if err != nil {
			handler_logger.Print(err.Error())
			utils.WriteCustomErrorResponse(w, errMsg, 500)
			return
		}
	}

	// These are items from receipts that I did not pay for, so I owe these
	for i := 0; i < len(itemsOwed); i++ {
		item := itemsOwed[i]
		total = total.Add(item.ItemAmount)
		amount, ok := resultMap[item.PaidByUserId]
		if ok {
			resultMap[item.PaidByUserId] = amount.Add(item.ItemAmount)
		} else {
			resultMap[item.PaidByUserId] = item.ItemAmount
		}
	}

	// These are items from receipts that I paid for, so they owe me
	for i := 0; i < len(itemsOthersOwe); i++ {
		item := itemsOthersOwe[i]
		total = total.Sub(item.ItemAmount)
		amount, ok := resultMap[item.ChargedToUserId]
		if ok {
			resultMap[item.ChargedToUserId] = amount.Sub(item.ItemAmount)
		} else {
			resultMap[item.ChargedToUserId] = amount.Mul(decimal.NewFromInt(-1))
		}
	}

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
	handler := structs.Handler{
		ErrorMessage: "Error resetting password.",
		Writer:       w,
		Request:      r,
		ResponseType: "",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := db.GetDB()
			id := chi.URLParam(r, "id")
			resetPasswordData := r.Context().Value("reset_password").(structs.ResetPasswordCommand)

			hashedPassword, err := utils.HashPassword(resetPasswordData.Password)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = db.Model(models.User{}).Where("id = ?", id).UpdateColumn("password", hashedPassword).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(200)
			return 0, nil
		},
	}

	HandleRequest(handler)
}

func ConvertDummyUserToNormalUser(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error converting user.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			var dbUser models.User
			db := db.GetDB()
			id := chi.URLParam(r, "id")
			resetPasswordData := r.Context().Value("reset_password").(structs.ResetPasswordCommand)

			err := db.Table("users").Where("id = ?", id).Find(&dbUser).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if !dbUser.IsDummyUser {
				return http.StatusInternalServerError, errors.New("user is already a normal user")
			}

			if len(resetPasswordData.Password) == 0 {
				return http.StatusInternalServerError, errors.New("password cannot be empty")
			}

			hashedPassword, err := utils.HashPassword(resetPasswordData.Password)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = db.Model(models.User{}).Where("id = ?", id).Updates(map[string]interface{}{"password": hashedPassword, "is_dummy_user": false}).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(200)
			return 0, nil
		},
	}

	HandleRequest(handler)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error Deleting User.",
		Writer:       w,
		Request:      r,
		ResponseType: "",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "id")
			token := utils.GetJWT(r)
			if simpleutils.UintToString(token.UserId) == id {
				return 500, errors.New("user cannot delete itself")
			}

			err := services.DeleteUser(id)
			if err != nil {
				return 500, err
			}

			w.WriteHeader(http.StatusOK)
			return 0, nil
		},
	}

	HandleRequest(handler)
}

func GetClaimsForLoggedInUser(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting claims",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := utils.GetJWT(r)
			services.PrepareAccessTokenClaims(*token)

			bytes, err := utils.MarshalResponseData(token)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error updating user profile",
		Writer:       w,
		Request:      r,
		ResponseType: constants.APPLICATION_JSON,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := utils.GetJWT(r)
			db := db.GetDB()
			updateProfileCommand := r.Context().Value("updateProfileCommand").(commands.UpdateProfileCommand)

			if len(updateProfileCommand.DisplayName) == 0 {
				return http.StatusBadRequest, errors.New("displayName is undefined")
			}

			err := db.Table("users").Where("id = ?", token.UserId).Update("display_name", updateProfileCommand.DisplayName).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			return 0, nil
		},
	}

	HandleRequest(handler)
}
