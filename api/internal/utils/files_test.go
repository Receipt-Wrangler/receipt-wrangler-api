package utils

import (
	"os"
	"path/filepath"
	"strings"
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

func TestFileExistsShouldReturnTrueForExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "exists.txt")
	if err := os.WriteFile(path, []byte("hi"), 0o600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if !FileExists(path) {
		t.Errorf("Expected FileExists to return true for an existing file")
	}
}

func TestFileExistsShouldReturnFalseForMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.txt")

	if FileExists(path) {
		t.Errorf("Expected FileExists to return false for a missing file")
	}
}

func TestMakeDirectoryShouldErrorWhenDirectoryExists(t *testing.T) {
	// t.TempDir() already exists, so re-creating it should error.
	err := MakeDirectory(t.TempDir())
	if err == nil {
		PrintTestError(t, nil, "error")
	}
}

func TestDirectoryExistsShouldReturnErrorWhenCreateFails(t *testing.T) {
	// The parent directory does not exist, so the single-level os.Mkdir
	// inside DirectoryExists fails even with createIfNotExist = true.
	nested := filepath.Join(t.TempDir(), "missing-parent", "child")

	err := DirectoryExists(nested, true)
	if err == nil {
		PrintTestError(t, nil, "error")
	}
}

func TestBuildGroupPathStringShouldBuildPath(t *testing.T) {
	path, err := BuildGroupPathString("42", "Groceries")
	if err != nil {
		PrintTestError(t, err, nil)
	}

	expectedSuffix := filepath.Join("data", "42-Groceries")
	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("Expected path to end with %q, got %q", expectedSuffix, path)
	}
}

func TestBuildFileNameShouldConcatenate(t *testing.T) {
	expected := "r1-f2-receipt.jpg"

	actual := BuildFileName("r1", "f2", "receipt.jpg")
	if actual != expected {
		PrintTestError(t, actual, expected)
	}
}

func TestGetMimeTypeShouldDetectPlainText(t *testing.T) {
	mime := GetMimeType([]byte("just some plain text content"))

	if !strings.HasPrefix(mime.String(), "text/plain") {
		t.Errorf("Expected a text/plain mime type, got %s", mime.String())
	}
}

func TestGetMimeTypeShouldDetectPng(t *testing.T) {
	// PNG file signature
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	mime := GetMimeType(pngHeader)
	if mime.String() != "image/png" {
		PrintTestError(t, mime.String(), "image/png")
	}
}
