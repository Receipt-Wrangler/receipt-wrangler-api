package models

type ReceiptGroup struct {
	Receipt   Receipt
	ReceiptId uint `gorm:"primaryKey"`
	Group     Group
	GroupId   uint `gorm:"primaryKey"`
}
