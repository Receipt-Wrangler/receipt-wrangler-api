package structs

type EmailSettings struct {
	Host     string
	Port     int
	Username string
	Password string
}

type AiSettings struct {
	AiType AiClientType `json:"type"`
	Key    string       `json:"key"`
	Url    string       `json:"url"`
}

type FeatureConfig struct {
	EnableLocalSignUp bool `json:"enableLocalSignUp"`
	AiPoweredReceipts bool `json:"aiPoweredReceipts"`
}

type DatabaseConfig struct {
	RootPassword string
	User         string
	Password     string
	Name         string
	Host         string
	Port         int
	Engine       string
	Filename     string
}

type Config struct {
	SecretKey            string          `json:"secretKey"`
	AiSettings           AiSettings      `json:"aiSettings"`
	EmailPollingInterval int             `json:"emailPollingInterval"`
	EmailSettings        []EmailSettings `json:"emailSettings"`
	Features             FeatureConfig   `json:"features"`
	Database             DatabaseConfig  `json:"database"`
}
