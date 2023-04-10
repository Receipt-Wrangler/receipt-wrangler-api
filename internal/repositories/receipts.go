package repositories

import (
	"net/http"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func PaginateReceipts(r *http.Request) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		q := r.URL.Query()
		page, _ := strconv.Atoi(q.Get("page"))
		if page <= 0 {
			page = 1
		}

		pageSize, _ := strconv.Atoi(q.Get("pageSize"))
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func GetReceiptById(receiptId string) (models.Receipt, error) {
	db := db.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", receiptId).Find(&receipt).Error
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

func GetPagedReceiptsByGroupId(groupId string, r *http.Request) ([]models.Receipt, error) {
	db := db.GetDB()
	var receipts []models.Receipt

	err := db.Scopes(PaginateReceipts(r)).Model(models.Receipt{}).Where("group_id = ?", groupId).Preload("Tags").Preload("Categories").Find(&receipts).Error
	if err != nil {
		return nil, err
	}

	return receipts, nil
}

func GetReceiptGroupIdByReceiptId(id string) (uint, error) {
	db := db.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", id).Select("group_id").Find(&receipt).Error
	if err != nil {
		return 0, err
	}

	return receipt.GroupId, nil
}

func GetFullyLoadedReceiptById(id string) (models.Receipt, error) {
	db := db.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", id).Preload(clause.Associations).Find(&receipt).Error
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}
