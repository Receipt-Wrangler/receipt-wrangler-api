package models

type Widget struct {
	BaseModel
	Name        string    `json:"name"`
	Dashboard   Dashboard `json:"-"`
	DashboardId uint      `gorm:"not null;" json:"dashboardId"`
}
