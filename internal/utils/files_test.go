package utils

import (
	"os"
	"testing"
)

func TestWritesFile(t *testing.T) {
  path := "test.txt"
	fileContents := "test"

	WriteFile(path, []byte(fileContents))

	_, err := os.Stat(path)

	if err != nil {
		PrintTestError(t, err, "Expected file to be written")
	}

	os.Remove(path)
}

func TestFileIsRead(t *testing.T) {
  path := "test.txt"
	fileContents := "test"

	WriteFile(path, []byte(fileContents))

	_, err := os.Stat(path)

	if err != nil {
		PrintTestError(t, err, "Expected file to be written")
	}

	contents, err := ReadFile(path)
	if err != nil {
		PrintTestError(t, err, "Expected contents to be read")
	}

	if string(contents) != "test" {
		PrintTestError(t, contents, "test")
	}

	os.Remove(path)
}

func TestShouldReturnNoErrIfDirExists(t *testing.T) {
  path := "../utils"
	
	err := DirectoryExists(path, false)
	if err != nil {
		PrintTestError(t, err, "Expected directory to exist")
	}
}

func TestShouldReturnErrIfDirDoesNotExists(t *testing.T) {
  path := "./fakeDir"
	
	err := DirectoryExists(path, false)
	if err == nil {
		PrintTestError(t, err, "Expected error to exist")
	}
}

func TestShouldCreateDirIfItDoesntExist(t *testing.T) {
  path := "./fakeDir"
	
	err := DirectoryExists(path, true)
	if err != nil {
		PrintTestError(t, err, "Expected no error")
	}

	err = DirectoryExists(path, false)
	if err != nil {
		PrintTestError(t, err, "Expected directory to exist")
	}

	os.Remove(path)
}

func TestShouldCreateDirectory(t *testing.T) {
  path := "./fakeDir"
	

	err := MakeDirectory(path)
	if err != nil {
		PrintTestError(t, err, "Expected no error")
	}

	err = DirectoryExists(path, false)
	if err != nil {
		PrintTestError(t, err, "Expected no error")
	}

	os.Remove(path)
}
