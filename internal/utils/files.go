package utils

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
)

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

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func DirectoryExists(dir string, createIfNotExist bool) error {
	_, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) && createIfNotExist {
		err = MakeDirectory(dir)
		if err != nil {
			return err
		}
	}

	if errors.Is(err, os.ErrNotExist) {
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

func ReadLastFileLine(filePath string) (string, error) {
	readFile, err := os.Open(filePath)
	if err != nil {
		if os.Getenv("ENV") == "test" {
			return "", nil
		}

		return "", err
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string

	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}

	err = readFile.Close()
	if err != nil {
		return "", err
	}

	if len(fileLines) == 0 {
		return "", nil
	}

	return fileLines[len(fileLines)-1], nil
}

func BuildGroupPathString(groupId string, groupName string) (string, error) {
	basePath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	groupPath := groupId + "-" + groupName
	return filepath.Join(basePath, "data", groupPath), nil
}

func BuildFileName(rid string, fid string, fname string) string {
	return rid + "-" + fid + "-" + fname
}

func GetMimeType(bytes []byte) *mimetype.MIME {
	return mimetype.Detect(bytes)
}
