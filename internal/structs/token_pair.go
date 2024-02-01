package structs

type TokenPair struct {
	Jwt          string `json:"jwt"`
	RefreshToken string `json:"refreshToken"`
}
