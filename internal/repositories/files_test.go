package repositories

import (
	"archive/zip"
	"bytes"
	"io"
	"receipt-wrangler/api/internal/utils"
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
		utils.PrintTestError(t, err, nil)
	}

	// Verify zip contents
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if len(zipReader.File) != len(filenames) {
		utils.PrintTestError(t, len(zipReader.File), len(filenames))
		return
	}

	// Check each file in the zip
	for i, zipFile := range zipReader.File {
		if zipFile.Name != filenames[i] {
			utils.PrintTestError(t, zipFile.Name, filenames[i])
		}

		rc, err := zipFile.Open()
		if err != nil {
			utils.PrintTestError(t, err, nil)
			continue
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			utils.PrintTestError(t, err, nil)
			continue
		}

		if string(content) != string(fileContents[i]) {
			utils.PrintTestError(t, string(content), string(fileContents[i]))
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
		utils.PrintTestError(t, err, expectedError)
	}
}

func TestShouldReturnErrorWhenNoFilesAreProvided(t *testing.T) {
	repository := NewFileRepository(nil)

	filenames := []string{}
	fileContents := [][]byte{}

	_, err := repository.ZipFiles(filenames, fileContents)

	expectedError := "no files to zip"
	if err == nil || err.Error() != expectedError {
		utils.PrintTestError(t, err, expectedError)
	}
}

func TestShouldHandleEmptyFileContent(t *testing.T) {
	repository := NewFileRepository(nil)

	filenames := []string{"empty.txt"}
	fileContents := [][]byte{[]byte("")}

	zipData, err := repository.ZipFiles(filenames, fileContents)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Verify zip contents
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if len(zipReader.File) != 1 {
		utils.PrintTestError(t, len(zipReader.File), 1)
		return
	}

	rc, err := zipReader.File[0].Open()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	content, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if len(content) != 0 {
		utils.PrintTestError(t, len(content), 0)
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
		utils.PrintTestError(t, err, nil)
		return
	}

	// Verify zip contents
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	rc, err := zipReader.File[0].Open()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	content, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	if len(content) != len(largeContent) {
		utils.PrintTestError(t, len(content), len(largeContent))
	}
}

func TestShouldHandleSpecialCharactersInFilenames(t *testing.T) {
	repository := NewFileRepository(nil)

	filenames := []string{"special!@#$%^&*.txt", "path/with/slashes.txt", "空白.txt"}
	fileContents := [][]byte{
		[]byte("Special content"),
		[]byte("Path content"),
		[]byte("Unicode content"),
	}

	zipData, err := repository.ZipFiles(filenames, fileContents)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Verify zip contents
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Check each filename exists in the zip
	for i, expectedName := range filenames {
		found := false
		for _, file := range zipReader.File {
			if file.Name == expectedName {
				found = true

				rc, err := file.Open()
				if err != nil {
					utils.PrintTestError(t, err, nil)
					break
				}

				content, err := io.ReadAll(rc)
				rc.Close()
				if err != nil {
					utils.PrintTestError(t, err, nil)
					break
				}

				if string(content) != string(fileContents[i]) {
					utils.PrintTestError(t, string(content), string(fileContents[i]))
				}

				break
			}
		}

		if !found {
			utils.PrintTestError(t, "File not found", expectedName)
		}
	}
}

// Example of a table test in the style of the existing tests
func TestShouldValidateZipFilesInput(t *testing.T) {
	repository := NewFileRepository(nil)

	tests := map[string]struct {
		filenames    []string
		fileContents [][]byte
		expectErr    bool
		expectedMsg  string
	}{
		"mismatched counts": {
			filenames:    []string{"file1.txt", "file2.txt"},
			fileContents: [][]byte{[]byte("Content")},
			expectErr:    true,
			expectedMsg:  "number of filenames does not match number of file contents",
		},
		"no files": {
			filenames:    []string{},
			fileContents: [][]byte{},
			expectErr:    true,
			expectedMsg:  "no files to zip",
		},
		"valid input": {
			filenames:    []string{"file.txt"},
			fileContents: [][]byte{[]byte("Content")},
			expectErr:    false,
		},
	}

	for _, test := range tests {
		zipData, err := repository.ZipFiles(test.filenames, test.fileContents)

		if test.expectErr {
			if err == nil || err.Error() != test.expectedMsg {
				utils.PrintTestError(t, err, test.expectedMsg)
			}
		} else {
			if err != nil {
				utils.PrintTestError(t, err, nil)
			}

			// Basic check that the zip was created
			if zipData == nil || len(zipData) == 0 {
				utils.PrintTestError(t, "empty data", "non-empty zip data")
			}
		}
	}
}
