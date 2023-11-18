package corspolicy

import (
	config "receipt-wrangler/api/internal/env"

	"github.com/rs/cors"
)

func GetCorsPolicy() *cors.Cors {
	env := config.GetDeployEnv()
	if env == "dev" {
		return cors.New(cors.Options{
			AllowedOrigins:      []string{"http://localhost:4200", "http://localhost:8100"},
			AllowCredentials:    true,
			AllowPrivateNetwork: true,
			AllowedMethods:      []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:      []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		})
	}

	if env == "prod" {
		return cors.New(cors.Options{
			AllowedOrigins:      []string{"http://localhost:4200", "http://localhost:8100"},
			AllowCredentials:    true,
			AllowPrivateNetwork: true,
			AllowedMethods:      []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:      []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		})
	}

	return cors.Default()
}
