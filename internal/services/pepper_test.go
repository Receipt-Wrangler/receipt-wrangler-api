package services

import (
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func TestPepperService_CreatePepper(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	pepperService := NewPepperService(nil)
	cleartextPepper, encryptedPepper, err := pepperService.CreatePepper()

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}
	if cleartextPepper == "" {
		utils.PrintTestError(t, cleartextPepper, "a non-empty string")
	}
	if encryptedPepper == "" {
		utils.PrintTestError(t, encryptedPepper, "a non-empty string")
	}
}

func TestPepperService_GetDecryptedPepper(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	pepperService := NewPepperService(nil)
	cleartextPepper, encryptedPepper, err := pepperService.CreatePepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	pepper := models.Pepper{
		Ciphertext: encryptedPepper,
		Algorithm:  "AES-256-GCM",
	}

	pepperRepository := repositories.NewPepperRepository(nil)
	err = pepperRepository.CreatePepper(pepper)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	decryptedPepper, err := pepperService.GetDecryptedPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if decryptedPepper != cleartextPepper {
		utils.PrintTestError(t, decryptedPepper, cleartextPepper)
	}
}

func TestPepperService_InitPepper_NoPepperExists(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()

	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Verify pepper was created by getting it
	cleartextPepper, err := pepperService.GetDecryptedPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}
	if cleartextPepper == "" {
		utils.PrintTestError(t, cleartextPepper, "a non-empty string")
	}

	var pepperCount int64
	repositories.GetDB().Model(&models.Pepper{}).Count(&pepperCount)
	if pepperCount != 1 {
		utils.PrintTestError(t, pepperCount, 1)
	}
}

func TestPepperService_InitPepper_PepperExists(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	pepperService := NewPepperService(nil)
	// First, create a pepper
	cleartextPepper, encryptedPepper, err := pepperService.CreatePepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}
	pepper := models.Pepper{
		Ciphertext: encryptedPepper,
		Algorithm:  "AES-256-GCM",
	}
	pepperRepository := repositories.NewPepperRepository(nil)
	err = pepperRepository.CreatePepper(pepper)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Now, run InitPepper again
	err = pepperService.InitPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	// Get the pepper to verify it's the same
	retrievedPepper, err := pepperService.GetDecryptedPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if retrievedPepper != cleartextPepper {
		utils.PrintTestError(t, retrievedPepper, cleartextPepper)
	}

	// Ensure no new pepper was created
	var pepperCount int64
	repositories.GetDB().Model(&models.Pepper{}).Count(&pepperCount)
	if pepperCount != 1 {
		utils.PrintTestError(t, pepperCount, 1)
	}
}

func TestPepperService_GetDecryptedPepper_NoPepper(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer repositories.TruncateTestDb()

	pepperService := NewPepperService(nil)
	_, err := pepperService.GetDecryptedPepper()

	if err == nil {
		utils.PrintTestError(t, err, "an error")
	}
}

func TestPepperService_CreatePepper_NoEncryptionKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "")
	defer repositories.TruncateTestDb()

	pepperService := NewPepperService(nil)
	_, _, err := pepperService.CreatePepper()

	if err == nil {
		utils.PrintTestError(t, err, "an error")
	}
}

func TestPepperService_GetDecryptedPepper_NoEncryptionKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "")
	defer repositories.TruncateTestDb()

	pepperService := NewPepperService(nil)
	pepper := models.Pepper{
		Ciphertext: "test-pepper",
		Algorithm:  "AES-256-GCM",
	}

	pepperRepository := repositories.NewPepperRepository(nil)
	err := pepperRepository.CreatePepper(pepper)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	_, err = pepperService.GetDecryptedPepper()

	if err == nil {
		utils.PrintTestError(t, err, "an error")
	}
}

func TestPepperService_InitPepper_CreatePepperError(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "")
	defer repositories.TruncateTestDb()

	pepperService := NewPepperService(nil)
	err := pepperService.InitPepper()

	if err == nil {
		utils.PrintTestError(t, err, "an error")
	}
}
