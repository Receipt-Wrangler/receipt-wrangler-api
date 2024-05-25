package structs

import "time"

type Attachment struct {
	Filename string `json:"filename"`
	FileType string `json:"fileType"`
	Size     uint   `json:"size"`
}

type EmailMetadata struct {
	Date             time.Time    `json:"date"`
	Subject          string       `json:"subject"`
	To               string       `json:"to"`
	FromName         string       `json:"fromName"`
	FromEmail        string       `json:"fromEmail"`
	Attachments      []Attachment `json:"attachments"`
	GroupSettingsIds []uint       `json:"groupSettingsIds"`
}
