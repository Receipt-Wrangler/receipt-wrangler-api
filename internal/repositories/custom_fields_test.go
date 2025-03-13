package repositories

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

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

func setupCustomFieldRepositoryTest() {
	createTestCustomFields()
}

func teardownCustomFieldRepositoryTest() {
	TruncateTestDb()
}

func TestShouldCreateCustomField(t *testing.T) {
	defer teardownCustomFieldRepositoryTest()

	repository := NewCustomFieldRepository(nil)
	createdBy := uint(1)

	// Create a TEXT type custom field
	command := commands.UpsertCustomFieldCommand{
		Name:        "Test New Text Field",
		Type:        models.TEXT,
		Description: "A new test text field",
		Options:     []commands.UpsertCustomFieldOptionCommand{},
	}

	customField, err := repository.CreateCustomField(command, &createdBy)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Validate the created custom field
	if customField.ID == 0 {
		utils.PrintTestError(t, "Custom field ID should not be 0", nil)
	}

	if customField.Name != "Test New Text Field" {
		utils.PrintTestError(t, customField.Name, "Test New Text Field")
	}

	if customField.Type != models.TEXT {
		utils.PrintTestError(t, customField.Type, models.TEXT)
	}

	if customField.Description != "A new test text field" {
		utils.PrintTestError(t, customField.Description, "A new test text field")
	}

	if *customField.CreatedBy != createdBy {
		utils.PrintTestError(t, *customField.CreatedBy, createdBy)
	}
}

func TestShouldCreateCustomFieldWithOptions(t *testing.T) {
	defer teardownCustomFieldRepositoryTest()

	repository := NewCustomFieldRepository(nil)
	createdBy := uint(1)

	// Create a SELECT type custom field with options
	command := commands.UpsertCustomFieldCommand{
		Name:        "Test New Select Field",
		Type:        models.SELECT,
		Description: "A new test select field",
		Options: []commands.UpsertCustomFieldOptionCommand{
			{
				Value: "Option A",
			},
			{
				Value: "Option B",
			},
		},
	}

	customField, err := repository.CreateCustomField(command, &createdBy)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Validate the created custom field
	if customField.ID == 0 {
		utils.PrintTestError(t, "Custom field ID should not be 0", nil)
	}

	if customField.Type != models.SELECT {
		utils.PrintTestError(t, customField.Type, models.SELECT)
	}

	if len(customField.Options) != 2 {
		utils.PrintTestError(t, len(customField.Options), 2)
		return
	}

	if customField.Options[0].Value != "Option A" {
		utils.PrintTestError(t, customField.Options[0].Value, "Option A")
	}

	if customField.Options[1].Value != "Option B" {
		utils.PrintTestError(t, customField.Options[1].Value, "Option B")
	}
}

func TestShouldGetPagedCustomFieldsWithDefaultSorting(t *testing.T) {
	defer teardownCustomFieldRepositoryTest()
	setupCustomFieldRepositoryTest()

	repository := NewCustomFieldRepository(nil)

	// Create paged request with default settings
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
}

func TestShouldGetPagedCustomFieldsWithLimit(t *testing.T) {
	defer teardownCustomFieldRepositoryTest()
	setupCustomFieldRepositoryTest()

	repository := NewCustomFieldRepository(nil)

	// Create paged request with limit
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

	// Total count should be 4
	if count != 4 {
		utils.PrintTestError(t, count, 4)
	}

	// But only 2 items should be returned
	if len(customFields) != 2 {
		utils.PrintTestError(t, len(customFields), 2)
	}
}

func TestShouldGetPagedCustomFieldsWithSecondPage(t *testing.T) {
	defer teardownCustomFieldRepositoryTest()
	setupCustomFieldRepositoryTest()

	repository := NewCustomFieldRepository(nil)

	// Create paged request for second page
	pagedRequest := commands.PagedRequestCommand{
		Page:          2,
		PageSize:      2,
		OrderBy:       "name",
		SortDirection: commands.ASCENDING,
	}

	customFields, count, err := repository.GetPagedCustomFields(pagedRequest)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Total count should be 4
	if count != 4 {
		utils.PrintTestError(t, count, 4)
	}

	// 2 items on the second page
	if len(customFields) != 2 {
		utils.PrintTestError(t, len(customFields), 2)
	}

	// The items on the second page should be different from the first page
	if customFields[0].Name != "Test Select Field" {
		utils.PrintTestError(t, customFields[0].Name, "Test Select Field")
	}

	if customFields[1].Name != "Test Text Field" {
		utils.PrintTestError(t, customFields[1].Name, "Test Text Field")
	}
}

func TestShouldGetPagedCustomFieldsWithDescendingOrder(t *testing.T) {
	defer teardownCustomFieldRepositoryTest()
	setupCustomFieldRepositoryTest()

	repository := NewCustomFieldRepository(nil)

	// Create paged request with descending order
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

func TestShouldReturnErrorWithInvalidOrderBy(t *testing.T) {
	defer teardownCustomFieldRepositoryTest()
	setupCustomFieldRepositoryTest()

	repository := NewCustomFieldRepository(nil)

	// Create paged request with invalid orderBy field
	pagedRequest := commands.PagedRequestCommand{
		Page:          1,
		PageSize:      10,
		OrderBy:       "invalid_column",
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

func TestShouldValidateOrderByColumn(t *testing.T) {
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

func TestShouldGetCustomFieldById(t *testing.T) {
	defer teardownCustomFieldRepositoryTest()
	setupCustomFieldRepositoryTest()

	repository := NewCustomFieldRepository(nil)
	db := GetDB()

	// Get a custom field from the test DB to use its ID
	var expectedCustomField models.CustomField
	db.Where("name = ?", "Test Text Field").First(&expectedCustomField)

	// Get the custom field by ID
	customField, err := repository.GetCustomFieldById(expectedCustomField.ID)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Validate the fetched custom field
	if customField.ID != expectedCustomField.ID {
		utils.PrintTestError(t, customField.ID, expectedCustomField.ID)
	}

	if customField.Name != "Test Text Field" {
		utils.PrintTestError(t, customField.Name, "Test Text Field")
	}

	if customField.Type != models.TEXT {
		utils.PrintTestError(t, string(customField.Type), string(models.TEXT))
	}

	if customField.Description != "A test text field" {
		utils.PrintTestError(t, customField.Description, "A test text field")
	}
}

func TestShouldGetCustomFieldByIdWithSelectType(t *testing.T) {
	defer teardownCustomFieldRepositoryTest()
	setupCustomFieldRepositoryTest()

	repository := NewCustomFieldRepository(nil)
	db := GetDB()

	// Get a SELECT type custom field from the test DB
	var expectedCustomField models.CustomField
	db.Where("name = ?", "Test Select Field").First(&expectedCustomField)

	// Get the custom field by ID
	customField, err := repository.GetCustomFieldById(expectedCustomField.ID)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Validate the fetched custom field
	if customField.ID != expectedCustomField.ID {
		utils.PrintTestError(t, customField.ID, expectedCustomField.ID)
	}

	if customField.Name != "Test Select Field" {
		utils.PrintTestError(t, customField.Name, "Test Select Field")
	}

	if customField.Type != models.SELECT {
		utils.PrintTestError(t, string(customField.Type), string(models.SELECT))
	}

	// Verify that options are preloaded
	if len(customField.Options) != 2 {
		utils.PrintTestError(t, len(customField.Options), 2)
		return
	}

	// Verify option values
	foundOption1 := false
	foundOption2 := false

	for _, option := range customField.Options {
		if option.Value == "Option 1" {
			foundOption1 = true
		}
		if option.Value == "Option 2" {
			foundOption2 = true
		}
	}

	if !foundOption1 {
		utils.PrintTestError(t, "Missing Option 1", "Option 1 should be present")
	}

	if !foundOption2 {
		utils.PrintTestError(t, "Missing Option 2", "Option 2 should be present")
	}
}

func TestShouldReturnErrorForNonExistentCustomFieldId(t *testing.T) {
	defer teardownCustomFieldRepositoryTest()
	setupCustomFieldRepositoryTest()

	repository := NewCustomFieldRepository(nil)

	// Try to get a custom field with a non-existent ID
	nonExistentId := uint(999)
	_, err := repository.GetCustomFieldById(nonExistentId)

	// Should return an error
	if err == nil {
		utils.PrintTestError(t, "Expected error for non-existent ID", nil)
	}
}

func TestShouldDeleteCustomField(t *testing.T) {
	defer teardownCustomFieldRepositoryTest()
	setupCustomFieldRepositoryTest()

	repository := NewCustomFieldRepository(nil)
	db := GetDB()

	// Get a TEXT type custom field to delete
	var customField models.CustomField
	db.Where("name = ?", "Test Text Field").First(&customField)
	customFieldId := customField.ID

	// Delete the custom field
	err := repository.DeleteCustomField(customFieldId)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Try to get the deleted custom field - should return error
	_, err = repository.GetCustomFieldById(customFieldId)
	if err == nil {
		utils.PrintTestError(t, "Expected error when getting deleted custom field", nil)
	}
}

func TestShouldDeleteCustomFieldWithOptions(t *testing.T) {
	defer teardownCustomFieldRepositoryTest()
	setupCustomFieldRepositoryTest()

	repository := NewCustomFieldRepository(nil)
	db := GetDB()

	// Get a SELECT type custom field with options
	var customField models.CustomField
	db.Where("name = ?", "Test Select Field").First(&customField)
	customFieldId := customField.ID

	// Verify options exist before deletion
	var optionsCount int64
	db.Model(&models.CustomFieldOption{}).Where("custom_field_id = ?", customFieldId).Count(&optionsCount)
	if optionsCount == 0 {
		utils.PrintTestError(t, "Expected options to exist before deletion", nil)
		return
	}

	// Delete the custom field
	err := repository.DeleteCustomField(customFieldId)
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}

	// Verify custom field was deleted
	var customFieldExists bool
	err = db.Model(&models.CustomField{}).
		Select("count(*) > 0").
		Where("id = ?", customFieldId).
		Find(&customFieldExists).
		Error
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}
	if customFieldExists {
		utils.PrintTestError(t, "Custom field should be deleted", nil)
	}

	// Verify options were deleted
	var optionsExist bool
	err = db.Model(&models.CustomFieldOption{}).
		Select("count(*) > 0").
		Where("custom_field_id = ?", customFieldId).
		Find(&optionsExist).
		Error
	if err != nil {
		utils.PrintTestError(t, err, nil)
		return
	}
	if optionsExist {
		utils.PrintTestError(t, "Custom field options should be deleted", nil)
	}
}
