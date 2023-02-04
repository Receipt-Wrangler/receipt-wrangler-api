package handlers

import (
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	userData := r.Context().Value("user").(models.User)
	validatorErrors := validateSignUpData(userData)

	if len(validatorErrors.Errors) > 0 {
		handler_logger.Print(validatorErrors)
		utils.WriteValidatorErrorResponse(w, validatorErrors, 500)
		return
	}

	// Hash password
	bytes, err := bcrypt.GenerateFromPassword([]byte(userData.Password), 14)
	if err != nil {
		handler_logger.Print(err.Error())
		utils.WriteErrorResponse(w, err, 500)
	}
	userData.Password = string(bytes)

	err = db.Transaction(func(tx *gorm.DB) error {
		var usrCnt int64
		// Set user's role
		 db.Model(models.User{}).Count(&usrCnt)
		// Save User
		if (usrCnt == 0) {
			userData.UserRole = models.ADMIN
		} else {
			userData.UserRole = models.USER
		}

		result := db.Create(&userData)

		if result.Error != nil {
			handler_logger.Print(err.Error())
			utils.WriteErrorResponse(w, result.Error, 500)
			return result.Error
		}

		// Create default group with user as group member
		group := models.Group{
			Name:           "Home",
			IsDefaultGroup: true,
			GroupMembers:   models.GroupMember{UserID: userData.ID, GroupRole: models.OWNER},
		}
		result = db.Create(&group)

		if result != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		handler_logger.Print(err.Error())
		return
	}

	w.WriteHeader(200)
}

func validateSignUpData(userData models.User) structs.ValidatorError {
	db := db.GetDB()
	err := structs.ValidatorError{
		Errors: make(map[string]string),
	}

	if len(userData.Username) == 0 {
		err.Errors["username"] = "Username is required"
	} else {
		var count int64
		db.Model(&models.User{}).Where("username = ?", userData.Username).Count(&count)

		if count > 0 {
			err.Errors["username"] = "Username already exists"
		}
	}

	if len(userData.Password) == 0 {
		err.Errors["password"] = "Password is required"
	}

	if len(userData.DisplayName) == 0 {
		err.Errors["displayName"] = "Displayname is required"
	}

	return err
}
