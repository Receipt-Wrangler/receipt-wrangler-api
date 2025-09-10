package models

type Pepper struct {
	BaseModel
	Algorithm  string `json:"algorithm"`
	Ciphertext string `json:"ciphertext"`
}
