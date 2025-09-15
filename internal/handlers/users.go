package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

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
	handler := structs.Handler{
		ErrorMessage: "Error retrieving users.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			userRepository := repositories.NewUserRepository(nil)
			users, err := userRepository.GetAllUserViews()
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := json.Marshal(users)
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

func CreateUser(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error creating user.",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := repositories.GetDB()
			var UserView structs.UserView
			bodyData := r.Context().Value("user").(commands.SignUpCommand)

			userRepository := repositories.NewUserRepository(nil)
			createdUser, err := userRepository.CreateUser(bodyData)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			// TODO: Move to repo
			err = db.Model(models.User{}).Where("id = ?", createdUser.ID).Find(&UserView).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(UserView)
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

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error updating user.",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := repositories.GetDB()
			id := chi.URLParam(r, "id")
			bodyData := r.Context().Value("user").(commands.SignUpCommand)

			//TODO: Move to repo
			err := db.Table("users").Select("username", "display_name", "user_role").Where("id = ?", id).Updates(&bodyData).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			return 0, nil
		},
	}

	HandleRequest(handler)
}

func GetAmountOwedForUser(w http.ResponseWriter, r *http.Request) {
	groupId := r.URL.Query().Get("groupId")
	err := r.ParseForm()
	receiptIds := r.Form["receiptIds"]

	handler := structs.Handler{
		ErrorMessage: "Error calculating amount owed.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		ReceiptIds:   receiptIds,
		GroupId:      groupId,
		GroupRole:    models.VIEWER,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			if err != nil {
				return http.StatusInternalServerError, err
			}

			db := repositories.GetDB()
			var itemsOwed []ItemView
			var itemsOthersOwe []ItemView
			total := decimal.NewFromInt(0)
			token := structs.GetClaims(r)
			id := token.UserId
			resultMap := make(map[uint]decimal.Decimal)
			totalReceiptIds := make([]uint, 0)
			totalGroupIds := make([]string, 0)

			if len(groupId) > 0 {
				groupRepository := repositories.NewGroupRepository(nil)
				receiptRepository := repositories.NewReceiptRepository(nil)

				uintGroupId, err := utils.StringToUint(groupId)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				isAllGroup, err := groupRepository.IsAllGroup(uintGroupId)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				if isAllGroup {
					groupMemberRepository := repositories.NewGroupMemberRepository(nil)
					userGroupIds, err := groupMemberRepository.GetGroupIdsByUserId(utils.UintToString(token.UserId))
					if err != nil {
						return http.StatusInternalServerError, err
					}

					for _, userGroupId := range userGroupIds {
						totalGroupIds = append(totalGroupIds, utils.UintToString(userGroupId))
					}
				} else {
					totalGroupIds = append(totalGroupIds, groupId)
				}

				groupReceiptIds, err := receiptRepository.GetReceiptsByGroupIds(totalGroupIds, "id", "")
				if err != nil {
					return http.StatusInternalServerError, err
				}

				for _, receipt := range groupReceiptIds {
					totalReceiptIds = append(totalReceiptIds, receipt.ID)
				}
			}

			if len(receiptIds) > 0 {
				for _, receiptId := range receiptIds {
					receiptIdUint, err := utils.StringToUint(receiptId)
					if err != nil {
						return http.StatusInternalServerError, err
					}

					totalReceiptIds = append(totalReceiptIds, receiptIdUint)
				}
			}

			err = db.Table("items").Select("items.id as item_id, items.receipt_id as receipt_id, items.amount as item_amount, items.charged_to_user_id, receipts.id, items.status, receipts.paid_by_user_id").Joins("inner join receipts on receipts.id=items.receipt_id").Where("items.charged_to_user_id=? AND receipts.paid_by_user_id !=? AND receipts.id IN ? AND items.status=?", id, id, totalReceiptIds, models.ITEM_OPEN).Scan(&itemsOwed).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = db.Table("items").Select("items.id as item_id, items.receipt_id as receipt_id, items.amount as item_amount, items.charged_to_user_id, receipts.id, items.status, receipts.paid_by_user_id").Joins("inner join receipts on receipts.id=items.receipt_id").Where("items.charged_to_user_id !=? AND receipts.paid_by_user_id =? AND receipts.id IN ? AND items.status=?", id, id, totalReceiptIds, models.ITEM_OPEN).Scan(&itemsOthersOwe).Error
			if err != nil {
				return http.StatusInternalServerError, err
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
					resultMap[item.ChargedToUserId] = item.ItemAmount.Mul(decimal.NewFromInt(-1))
				}
			}

			bytes, err := json.Marshal(resultMap)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(200)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}

func GetUsernameCount(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting username count.",
		Writer:       w,
		Request:      r,
		ResponseType: constants.TextPlain,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := repositories.GetDB()
			username := chi.URLParam(r, "username")
			var result int64

			// TODO: user repo count func
			err := db.Model(models.User{}).Where("username = ?", username).Count(&result).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(result)
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

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error resetting password.",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			db := repositories.GetDB()
			id := chi.URLParam(r, "id")
			resetPasswordData := r.Context().Value("reset_password").(structs.ResetPasswordCommand)

			// TODO: move to service
			hashedPassword, err := utils.HashPassword(resetPasswordData.Password)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = db.Model(models.User{}).Where("id = ?", id).UpdateColumn("password", hashedPassword).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
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
		UserRole:     models.ADMIN,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			var dbUser models.User
			db := repositories.GetDB()
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
		UserRole:     models.ADMIN,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			id := chi.URLParam(r, "id")
			token := structs.GetClaims(r)
			if utils.UintToString(token.UserId) == id {
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
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)
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
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)
			db := repositories.GetDB()
			updateProfileCommand := r.Context().Value("updateProfileCommand").(commands.UpdateProfileCommand)

			if len(updateProfileCommand.DisplayName) == 0 {
				return http.StatusBadRequest, errors.New("displayName is undefined")
			}

			// TODO: move to repo
			err := db.Table("users").Where("id = ?", token.UserId).Updates(map[string]interface{}{"display_name": updateProfileCommand.DisplayName, "default_avatar_color": updateProfileCommand.DefaultAvatarColor}).Error
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			return 0, nil
		},
	}

	HandleRequest(handler)
}

func GetAppData(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error getting app data",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)
			appData, err := services.GetAppData(token.UserId, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(appData)
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

func BulkDeleteUsers(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error bulk deleting users",
		Writer:       w,
		Request:      r,
		UserRole:     models.ADMIN,
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			command := commands.BulkUserDeleteCommand{}
			err := command.LoadDataFromRequest(w, r)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			token := structs.GetClaims(r)
			currentUserId := utils.UintToString(token.UserId)

			// Check if the current user is trying to delete themselves
			for _, userId := range command.UserIds {
				if userId == currentUserId {
					return http.StatusBadRequest, errors.New("user cannot delete itself")
				}
			}

			err = services.BulkDeleteUsers(command.UserIds)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(http.StatusOK)
			return 0, nil
		},
	}

	HandleRequest(handler)
}
