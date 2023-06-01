package simpleutils

import (
	"os"
	"path/filepath"
)

func BuildGroupPathString(groupId string, groupName string) (string, error) {
	basePath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	groupPath := groupId + "-" + groupName
	return filepath.Join(basePath, "data", groupPath), nil
}
