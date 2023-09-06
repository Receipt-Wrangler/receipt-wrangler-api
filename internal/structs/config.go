package structs

type EmailSettings struct {
	Host     string
	Port     int
	Username string
	Password string
}

type Config struct {
	SecretKey     string
	OpenAiKey     string
	EmailSettings []EmailSettings
}
