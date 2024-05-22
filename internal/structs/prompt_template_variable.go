package structs

type PromptTemplateVariable string

const (
	CATEGORIES   PromptTemplateVariable = "@categories"
	TAGS         PromptTemplateVariable = "@tags"
	OCR_TEXT     PromptTemplateVariable = "@ocrText"
	CURRENT_YEAR PromptTemplateVariable = "@currentYear"
)
