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

type Config struct {
	SecretKey            string
	OpenAiKey            string
	AiSettings           AiSettings
	EmailPollingInterval int
	EmailSettings        []EmailSettings
}
