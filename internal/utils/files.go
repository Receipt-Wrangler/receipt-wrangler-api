package utils

import (
	"os"
	"path/filepath"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
)

func BuildFilePath(rid string, fId string, fname string) (string, error) {
	db := db.GetDB()
	var receipt models.Receipt
	var group models.Group

	err := db.Model(models.Receipt{}).Where("id = ?", rid).Select("group_id").Find(&receipt).Error
	if err != nil {
		return "", err
	}

	basePath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	err = db.Model(models.Group{}).Where("id = ?", receipt.GroupId).Select("id", "name").Find(&group).Error
	if err != nil {
		return "", err
	}

	strGroupId := UintToString(group.ID)

	fileName := rid + "-" + fId + "-" + fname
	groupPath := strGroupId + "-" + group.Name
	path := filepath.Join(basePath, "data", groupPath, fileName)

	return path, nil
}
