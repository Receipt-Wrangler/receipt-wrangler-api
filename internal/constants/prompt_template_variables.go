package constants

import "receipt-wrangler/api/internal/structs"

func GetPromptTemplateVariables() []structs.PromptTemplateVariable {
	return []structs.PromptTemplateVariable{
		structs.CATEGORIES,
		structs.TAGS,
		structs.OCR_TEXT,
		structs.CURRENT_YEAR,
	}
}
