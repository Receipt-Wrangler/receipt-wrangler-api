package services

import (
	"errors"
	"fmt"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
)

func CreateDefaultPrompt() (models.Prompt, error) {
	db := repositories.GetDB()
	var defaultPromptCount int64
	db.Model(models.Prompt{}).Where("name = ?", constants.DEFAULT_PROMPT_NAME).Count(&defaultPromptCount)

	defaultPrompt := fmt.Sprintf(`
	Find the receipt's name, total cost, and date. Format the found data as:
	{
		"name": store name,
		"amount": amount as a number,
		"date": date in ISO 18601 format in UTC with ALL time values set as 0,
		"categories": categories,
		"tags": tags
	}
	If a store name cannot be confidently found, use 'Default store name' as the default name.
	Omit any value if not found with confidence. Assume the date is in the year @currentYear if not provided.
	The amount must be a float or integer.

	Please do NOT add any additional information, only valid JSON.
	Please return the json in plaintext ONLY, do not ever return it in a code block or any other format.

	Choose up to 2 categories from the given list based on the receipt's items and store name. If no categories fit, please return an empty array for the field and do not select any categories. When selecting categories, select only the id, like:
	{
		Id: category id
	}

	Emphasize the relationship between the category and the receipt, and use the description of the category to fine tune the results. Do not return categories that have an empty name or do not exist.


	Categories: @categories

	Follow the same process as described for categories for tags.

	Tags: @tags

	Receipt text: @ocrText
`)

	if defaultPromptCount > 0 {
		return models.Prompt{}, errors.New("Default prompt.go already exists")
	}

	promptRepository := repositories.NewPromptRepository(nil)
	command := commands.UpsertPromptCommand{
		Name:        constants.DEFAULT_PROMPT_NAME,
		Description: "Default prompt.go used for previous versions of Receipt Wrangler.",
		Prompt:      defaultPrompt,
	}

	return promptRepository.CreatePrompt(command)
}
