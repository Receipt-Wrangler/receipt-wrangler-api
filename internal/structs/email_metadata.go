package structs

import "time"

type Attachment struct {
	Filename string `json:"filename"`
}

type EmailMetadata struct {
	Date        time.Time    `json:"date"`
	Subject     string       `json:"subject"`
	To          string       `json:"to"`
	FromName    string       `json:"fromName"`
	FromEmail   string       `json:"fromEmail"`
	Attachments []Attachment `json:"attachments"`
}
