package repositories

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func setupCustomFieldTest() {
	createTestCustomFields()
}

func createTestCustomFields() {
	db := GetDB()

	// Create custom field with TEXT type
	textField := models.CustomField{
		Name:        "Test Text Field",
		Type:        models.TEXT,
		Description: "A test text field",
	}
	db.Create(&textField)

	// Create custom field with DATE type
	dateField := models.CustomField{
		Name:        "Test Date Field",
		Type:        models.DATE,
		Description: "A test date field",
	}
	db.Create(&dateField)

	// Create custom field with SELECT type and options
	selectField := models.CustomField{
		Name:        "Test Select Field",
		Type:        models.SELECT,
		Description: "A test select field",
	}
	db.Create(&selectField)

	// Add options to the SELECT field
	option1 := models.CustomFieldOption{
		Value:         "Option 1",
		CustomFieldId: selectField.ID,
	}
	option2 := models.CustomFieldOption{
		Value:         "Option 2",
		CustomFieldId: selectField.ID,
	}
	db.Create(&option1)
	db.Create(&option2)

	// Create custom field with CURRENCY type
	currencyField := models.CustomField{
		Name:        "Test Currency Field",
		Type:        models.CURRENCY,
		Description: "A test currency field",
	}
	db.Create(&currencyField)
}

func teardownCustomFieldTest() {
	TruncateTestDb()
}

func TestShouldGetPagedCustomFields(t *testing.T) {
	defer teardownCustomFieldTest()
	setupCustomFieldTest()

	repository := NewCustomFieldRepository(nil)

	// Create paged request with name ordering
	pagedRequest := commands.PagedRequestCommand{
		Page:          1,
		PageSize:      10,
		OrderBy:       "name",
		SortDirection: commands.ASCENDING,
	}

	customFields, count, err := repository.GetPagedCustomFields(pagedRequest)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Should return 4 custom fields
	if count != 4 {
		utils.PrintTestError(t, count, 4)
	}

	if len(customFields) != 4 {
		utils.PrintTestError(t, len(customFields), 4)
	}

	// Check if fields are correctly ordered by name
	if customFields[0].Name != "Test Currency Field" {
		utils.PrintTestError(t, customFields[0].Name, "Test Currency Field")
	}

	if customFields[1].Name != "Test Date Field" {
		utils.PrintTestError(t, customFields[1].Name, "Test Date Field")
	}

	if customFields[2].Name != "Test Select Field" {
		utils.PrintTestError(t, customFields[2].Name, "Test Select Field")
	}

	if customFields[3].Name != "Test Text Field" {
		utils.PrintTestError(t, customFields[3].Name, "Test Text Field")
	}
}

func TestShouldGetPagedCustomFieldsWithTypeOrder(t *testing.T) {
	defer teardownCustomFieldTest()
	setupCustomFieldTest()

	repository := NewCustomFieldRepository(nil)

	// Create paged request with type ordering
	pagedRequest := commands.PagedRequestCommand{
		Page:          1,
		PageSize:      10,
		OrderBy:       "type",
		SortDirection: commands.ASCENDING,
	}

	customFields, count, err := repository.GetPagedCustomFields(pagedRequest)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Should return 4 custom fields
	if count != 4 {
		utils.PrintTestError(t, count, 4)
	}

	// Check if fields are ordered by type
	if customFields[0].Type != models.CURRENCY {
		utils.PrintTestError(t, customFields[0].Type, models.CURRENCY)
	}
}

func TestShouldGetPagedCustomFieldsWithDescriptionOrder(t *testing.T) {
	defer teardownCustomFieldTest()
	setupCustomFieldTest()

	repository := NewCustomFieldRepository(nil)

	// Create paged request with description ordering
	pagedRequest := commands.PagedRequestCommand{
		Page:          1,
		PageSize:      10,
		OrderBy:       "description",
		SortDirection: commands.ASCENDING,
	}

	_, count, err := repository.GetPagedCustomFields(pagedRequest)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Should return 4 custom fields
	if count != 4 {
		utils.PrintTestError(t, count, 4)
	}
}

func TestShouldGetPagedCustomFieldsWithDescendingOrder(t *testing.T) {
	defer teardownCustomFieldTest()
	setupCustomFieldTest()

	repository := NewCustomFieldRepository(nil)

	// Create paged request with descending name ordering
	pagedRequest := commands.PagedRequestCommand{
		Page:          1,
		PageSize:      10,
		OrderBy:       "name",
		SortDirection: commands.DESCENDING,
	}

	customFields, count, err := repository.GetPagedCustomFields(pagedRequest)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Should return 4 custom fields
	if count != 4 {
		utils.PrintTestError(t, count, 4)
	}

	// Check if fields are correctly ordered by name in descending order
	if customFields[0].Name != "Test Text Field" {
		utils.PrintTestError(t, customFields[0].Name, "Test Text Field")
	}

	if customFields[3].Name != "Test Currency Field" {
		utils.PrintTestError(t, customFields[3].Name, "Test Currency Field")
	}
}

func TestShouldGetPagedCustomFieldsWithPagination(t *testing.T) {
	defer teardownCustomFieldTest()
	setupCustomFieldTest()

	repository := NewCustomFieldRepository(nil)

	// Create paged request with pagination (2 items per page)
	pagedRequest := commands.PagedRequestCommand{
		Page:          1,
		PageSize:      2,
		OrderBy:       "name",
		SortDirection: commands.ASCENDING,
	}

	customFields, count, err := repository.GetPagedCustomFields(pagedRequest)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Total count should still be 4
	if count != 4 {
		utils.PrintTestError(t, count, 4)
	}

	// But we should only get 2 items
	if len(customFields) != 2 {
		utils.PrintTestError(t, len(customFields), 2)
	}

	// Get the second page
	pagedRequest.Page = 2
	customFields, count, err = repository.GetPagedCustomFields(pagedRequest)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Total count should still be 4
	if count != 4 {
		utils.PrintTestError(t, count, 4)
	}

	// And we should get the other 2 items
	if len(customFields) != 2 {
		utils.PrintTestError(t, len(customFields), 2)
	}
}

func TestShouldReturnErrorWithInvalidOrderBy(t *testing.T) {
	defer teardownCustomFieldTest()
	setupCustomFieldTest()

	repository := NewCustomFieldRepository(nil)

	// Create paged request with invalid orderBy field
	pagedRequest := commands.PagedRequestCommand{
		Page:          1,
		PageSize:      10,
		OrderBy:       "invalid_field",
		SortDirection: commands.ASCENDING,
	}

	_, _, err := repository.GetPagedCustomFields(pagedRequest)

	// Should return an error
	if err == nil {
		utils.PrintTestError(t, "Expected error for invalid orderBy", nil)
	}

	if err.Error() != "invalid orderBy" {
		utils.PrintTestError(t, err.Error(), "invalid orderBy")
	}
}

func TestShouldValidateOrderBy(t *testing.T) {
	repository := NewCustomFieldRepository(nil)

	// Valid orderBy values
	if err := repository.validateOrderBy("name"); err != nil {
		utils.PrintTestError(t, err, nil)
	}

	if err := repository.validateOrderBy("type"); err != nil {
		utils.PrintTestError(t, err, nil)
	}

	if err := repository.validateOrderBy("description"); err != nil {
		utils.PrintTestError(t, err, nil)
	}

	// Invalid orderBy values
	if err := repository.validateOrderBy("id"); err == nil {
		utils.PrintTestError(t, "Expected error for invalid orderBy 'id'", nil)
	}

	if err := repository.validateOrderBy("created_at"); err == nil {
		utils.PrintTestError(t, "Expected error for invalid orderBy 'created_at'", nil)
	}
}
