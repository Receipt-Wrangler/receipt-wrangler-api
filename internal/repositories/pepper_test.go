package repositories

import (
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func TestCreatePepper(t *testing.T) {
	defer TruncateTestDb()
	pepperRepository := NewPepperRepository(nil)
	pepper := models.Pepper{
		Algorithm:  "test-algorithm",
		Ciphertext: "test-pepper",
	}

	err := pepperRepository.CreatePepper(pepper)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	var createdPepper models.Pepper
	GetDB().First(&createdPepper)

	if createdPepper.Ciphertext != "test-pepper" {
		utils.PrintTestError(t, createdPepper.Ciphertext, "test-pepper")
	}
}

func TestGetPepper(t *testing.T) {
	defer TruncateTestDb()
	pepperRepository := NewPepperRepository(nil)
	pepper := models.Pepper{
		Algorithm:  "test-algorithm",
		Ciphertext: "test-pepper",
	}
	err := pepperRepository.CreatePepper(pepper)
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	retrievedPepper, err := pepperRepository.GetPepper()
	if err != nil {
		utils.PrintTestError(t, err, "no error")
	}

	if retrievedPepper.Ciphertext != "test-pepper" {
		utils.PrintTestError(t, retrievedPepper.Ciphertext, "test-pepper")
	}
}

func TestGetPepperNotFound(t *testing.T) {
	defer TruncateTestDb()
	pepperRepository := NewPepperRepository(nil)

	_, err := pepperRepository.GetPepper()
	if err == nil {
		utils.PrintTestError(t, err, "error")
	}
}
