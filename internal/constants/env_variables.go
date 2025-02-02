package constants

type EnvironmentVariable string

const (
	EncryptionKey EnvironmentVariable = "ENCRYPTION_KEY"
	SecretKey     EnvironmentVariable = "SECRET_KEY"
	DbUser        EnvironmentVariable = "DB_USER"
	DbPassword    EnvironmentVariable = "DB_PASSWORD"
	DbHost        EnvironmentVariable = "DB_HOST"
	DbPort        EnvironmentVariable = "DB_PORT"
	DbName        EnvironmentVariable = "DB_NAME"
	DbFileName    EnvironmentVariable = "DB_FILENAME"
	DbEngine      EnvironmentVariable = "DB_ENGINE"
	RedisHost     EnvironmentVariable = "REDIS_HOST"
	RedisPort     EnvironmentVariable = "REDIS_PORT"
	RedisUser     EnvironmentVariable = "REDIS_USER"
	RedisPassword EnvironmentVariable = "REDIS_PASSWORD"
	BasePath      EnvironmentVariable = "BASE_PATH"
	Env           EnvironmentVariable = "ENV"
)
