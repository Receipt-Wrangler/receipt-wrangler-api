package repositories

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

func TestShouldZipMultipleFilesSuccessfully(t *testing.T) {
	repository := NewFileRepository(nil)

	filenames := []string{"file1.txt", "file2.txt", "file3.txt"}
	fileContents := [][]byte{
		[]byte("Content of file 1"),
		[]byte("Content of file 2"),
		[]byte("Content of file 3"),
	}

	zipData, err := repository.ZipFiles(filenames, fileContents)
	if err != nil {
		t.Errorf("ZipFiles() error = %v, want nil", err)
	}

	// Verify zip contents
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Errorf("Failed to read zip content: %v", err)
		return
	}

	if len(zipReader.File) != len(filenames) {
		t.Errorf("ZipFiles() created zip with %d files, want %d", len(zipReader.File), len(filenames))
		return
	}

	// Check each file in the zip
	for i, zipFile := range zipReader.File {
		if zipFile.Name != filenames[i] {
			t.Errorf("ZipFiles() file[%d] name = %v, want %v", i, zipFile.Name, filenames[i])
		}

		rc, err := zipFile.Open()
		if err != nil {
			t.Errorf("Failed to open zip file %v: %v", zipFile.Name, err)
			continue
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Errorf("Failed to read zip file %v: %v", zipFile.Name, err)
			continue
		}

		if string(content) != string(fileContents[i]) {
			t.Errorf("ZipFiles() file[%d] content = %v, want %v", i, string(content), string(fileContents[i]))
		}
	}
}

func TestShouldReturnErrorWhenFilenamesAndContentsDontMatch(t *testing.T) {
	repository := NewFileRepository(nil)

	filenames := []string{"file1.txt", "file2.txt"}
	fileContents := [][]byte{[]byte("Content of file 1")}

	_, err := repository.ZipFiles(filenames, fileContents)

	expectedError := "number of filenames does not match number of file contents"
	if err == nil || err.Error() != expectedError {
		t.Errorf("ZipFiles() error = %v, want %v", err, expectedError)
	}
}

func TestShouldReturnErrorWhenNoFilesAreProvided(t *testing.T) {
	repository := NewFileRepository(nil)

	filenames := []string{}
	fileContents := [][]byte{}

	_, err := repository.ZipFiles(filenames, fileContents)

	expectedError := "no files to zip"
	if err == nil || err.Error() != expectedError {
		t.Errorf("ZipFiles() error = %v, want %v", err, expectedError)
	}
}

func TestShouldHandleEmptyFileContent(t *testing.T) {
	repository := NewFileRepository(nil)

	filenames := []string{"empty.txt"}
	fileContents := [][]byte{[]byte("")}

	zipData, err := repository.ZipFiles(filenames, fileContents)
	if err != nil {
		t.Errorf("ZipFiles() error = %v, want nil", err)
		return
	}

	// Verify zip contents
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Errorf("Failed to read zip content: %v", err)
		return
	}

	if len(zipReader.File) != 1 {
		t.Errorf("ZipFiles() created zip with %d files, want 1", len(zipReader.File))
		return
	}

	rc, err := zipReader.File[0].Open()
	if err != nil {
		t.Errorf("Failed to open zip file: %v", err)
		return
	}

	content, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		t.Errorf("Failed to read zip file: %v", err)
		return
	}

	if len(content) != 0 {
		t.Errorf("ZipFiles() file content length = %d, want 0", len(content))
	}
}

func TestShouldHandleLargeFileContent(t *testing.T) {
	repository := NewFileRepository(nil)

	// Create a 100KB file
	largeContent := bytes.Repeat([]byte("A"), 100*1024)

	filenames := []string{"large.txt"}
	fileContents := [][]byte{largeContent}

	zipData, err := repository.ZipFiles(filenames, fileContents)
	if err != nil {
		t.Errorf("ZipFiles() error = %v, want nil", err)
		return
	}

	// Verify zip contents
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Errorf("Failed to read zip content: %v", err)
		return
	}

	rc, err := zipReader.File[0].Open()
	if err != nil {
		t.Errorf("Failed to open zip file: %v", err)
		return
	}

	content, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		t.Errorf("Failed to read zip file: %v", err)
		return
	}

	if len(content) != len(largeContent) {
		t.Errorf("ZipFiles() file content length = %d, want %d", len(content), len(largeContent))
	}
}
