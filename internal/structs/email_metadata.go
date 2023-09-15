package structs

import "time"

type Attachment struct {
	FileName string `json:"fileName"`
}

type EmailMetadata struct {
	Date        time.Time    `json:"date"`
	Subject     string       `json:"subject"`
	To          string       `json:"to"`
	FromName    string       `json:"fromName"`
	FromEmail   string       `json:"fromEmail"`
	Attachments []Attachment `json:"attachments"`
}
