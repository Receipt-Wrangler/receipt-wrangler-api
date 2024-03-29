package structs

type EmailSettings struct {
	Host     string
	Port     int
	Username string
	Password string
}

type AiSettings struct {
	AiType     AiClientType `json:"type"`
	Key        string       `json:"key"`
	Url        string       `json:"url"`
	NumWorkers int          `json:"numWorkers"`
}

type FeatureConfig struct {
	EnableLocalSignUp bool `json:"enableLocalSignUp"`
	AiPoweredReceipts bool `json:"aiPoweredReceipts"`
}

type DatabaseConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Engine   string `json:"engine"`
	Filename string `json:"filename"`
}

type DebugConfig struct {
	DebugOcr bool `json:"debugOcr"`
}

type Config struct {
	SecretKey            string          `json:"secretKey"`
	AiSettings           AiSettings      `json:"aiSettings"`
	EmailPollingInterval int             `json:"emailPollingInterval"`
	EmailSettings        []EmailSettings `json:"emailSettings"`
	Features             FeatureConfig   `json:"features"`
	Database             DatabaseConfig  `json:"database"`
	Debug                DebugConfig     `json:"debug"`
}
