package utils

import (
	"errors"
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

func WriteFile(path string, data []byte) error {
	// TODO: Fix perms
	err := os.WriteFile(path, data, 777)
	if err != nil {
		return err
	}

	return nil
}

func ReadFile(path string) ([]byte, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}

	return bytes, nil
}

func GetBytesForFileData(fileData models.FileData) ([]byte, error) {
	path, err := BuildFilePath(UintToString(fileData.ReceiptId), UintToString(fileData.ID), fileData.Name)
	if err != nil {
		return nil, err
	}

	bytes, err := ReadFile(path)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func DirectoryExists(dir string, createIfNotExist bool) error {
	_, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) && createIfNotExist {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
