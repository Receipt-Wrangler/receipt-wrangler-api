package corspolicy

import (
	config "receipt-wrangler/api/internal/env"

	"github.com/rs/cors"
)

func GetCorsPolicy() *cors.Cors {
	env := config.GetDeployEnv()
	if env == "dev" {
		return cors.New(cors.Options{
			AllowedOrigins:      []string{"http://localhost:4200"},
			AllowCredentials:    true,
			AllowPrivateNetwork: true,
			AllowedMethods:      []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:      []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		})
	}

	if env == "prod" {
		return cors.Default()
	}

	return cors.Default()
}
