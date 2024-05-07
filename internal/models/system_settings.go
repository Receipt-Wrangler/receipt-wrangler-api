package models

type SystemSettings struct {
	BaseModel
	EnableLocalSignUp    bool `json:"enableLocalSignUp;" gorm:"default:false"`
	AiPoweredReceipts    bool `json:"aiPoweredReceipts;" gorm:"default:false"`
	EmailPollingInterval int  `json:"emailPollingInterval;" gorm:"default:1800"`
}
