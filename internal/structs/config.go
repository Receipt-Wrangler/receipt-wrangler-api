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

type Config struct {
	SecretKey            string          `json:"-"`
	OpenAiKey            string          `json:"-"`
	AiSettings           AiSettings      `json:"-"`
	EmailPollingInterval int             `json:"-"`
	EmailSettings        []EmailSettings `json:"-"`
	Features             FeatureConfig   `json:"features"`
}
