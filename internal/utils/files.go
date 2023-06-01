package utils

import (
	"errors"
	"os"
	"path/filepath"
	db "receipt-wrangler/api/internal/database"
	"receipt-wrangler/api/internal/models"
)

func BuildFilePath(rid string, fid string, fname string) (string, error) {
	db := db.GetDB()
	var receipt models.Receipt

	err := db.Model(models.Receipt{}).Where("id = ?", rid).Select("group_id").Find(&receipt).Error
	if err != nil {
		return "", err
	}

	groupPath, err := BuildGroupPath(receipt.GroupId, "")
	if err != nil {
		return "", err
	}

	fileName := BuildFileName(rid, fid, fname)
	path := filepath.Join(groupPath, fileName)

	return path, nil
}

func BuildFileName(rid string, fid string, fname string) string {
	return rid + "-" + fid + "-" + fname
}

func BuildGroupPath(groupId uint, alternateGroupName string) (string, error) {
	db := db.GetDB()
	var groupNameToUse string

	basePath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if len(alternateGroupName) > 0 {
		groupNameToUse = alternateGroupName
	} else {
		var group models.Group
		err = db.Model(models.Group{}).Where("id = ?", groupId).Select("name").Find(&group).Error
		if err != nil {
			return "", err
		}
		groupNameToUse = group.Name
	}

	strGroupId := UintToString(groupId)
	groupPath := strGroupId + "-" + groupNameToUse

	return filepath.Join(basePath, "data", groupPath), nil
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
		err = MakeDirectory(dir)
		if err != nil {
			return err
		}
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func MakeDirectory(dir string) error {
	err := os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
