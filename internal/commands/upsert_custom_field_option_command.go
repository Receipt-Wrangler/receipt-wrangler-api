package commands

type UpsertCustomFieldOptionCommand struct {
	Value         string `json:"value"`
	CustomFieldId uint   `json:"custom_field_id"`
}
