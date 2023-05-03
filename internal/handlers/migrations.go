package handlers

import (
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"

	"gorm.io/gorm"
)

func MigratetionMigrateIsResolvedToStatus(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error migrating is resolved to status",
		Writer:       w,
		Request:      r,
		ResponseType: "",
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			err := db.GetDB().Transaction(func(tx *gorm.DB) error {

				err := tx.Model(&models.Receipt{}).Where("is_resolved = ?", true).Update("status", models.RESOLVED).Error
				if err != nil {
					return err
				}

				err = tx.Model(&models.Receipt{}).Where("is_resolved = ?", false).Update("status", models.OPEN).Error
				if err != nil {
					return err
				}

				err = tx.Migrator().DropColumn(&models.Receipt{}, "is_resolved")
				if err != nil {
					return err
				}

				return nil
			})

			if err != nil {
				return 500, err
			}

			w.WriteHeader(200)
			return 0, nil
		},
	}

	HandleRequest(handler)
}
