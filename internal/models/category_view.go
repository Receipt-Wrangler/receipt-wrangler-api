package models

type CategoryView struct {
	Category
	NumberOfReceipts int `json:"numberOfReceipts"`
}
