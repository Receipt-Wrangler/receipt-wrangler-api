[![Build](https://github.com/Noah231515/receipt-wrangler-api/actions/workflows/docker-image.yml/badge.svg)](https://github.com/Noah231515/receipt-wrangler-api/actions/workflows/docker-image.yml) [![codecov](https://codecov.io/gh/Receipt-Wrangler/receipt-wrangler-api/graph/badge.svg?token=EUQMLBEKPK)](https://codecov.io/gh/Receipt-Wrangler/receipt-wrangler-api)

## Development

To run the api for local development:

1. Clone the repository
2. Install GO on your OS. See https://go.dev/doc/install for more details
3. Install tesseract dependencies. If you are running a debian derivative (recommended), run `sh set-up-tesseract-env-sudo.sh`, otherwise see https://github.com/otiai10/gosseract for other installation details.
4. Install OpenAPI generator https://openapi-generator.tech/docs/installation (I personally use the jar installation)
5. Set up a mariadb instance. The easiest way to do this is via docker. F.ex: `docker run --name receipt-wrangler-db -e MYSQL_ROOT_PASSWORD=123456 -e MYSQL_USER=wrangler -e MYSQL_PASSWORD=123456 -e MYSQL_DATABASE=wrangler -p 9001:3306  -d library/mariadb`
6. Add a config.dev.json and a feature-config.dev.json, copy pasta the samples and modify as needed.
7. Add env variables for db connection. For example, add the following to the bottom of your .bashrc `# Export wrangler db env variables export DB_ROOT_PASSWORD="123456"
export DB_USER="wrangler"
export DB_PASSWORD="123456"
export DB_NAME="wrangler"
export DB_HOST="0.0.0.0:9001"
export DB_ENGINE="mariadb"`
8. Once configs are added, and db is up, run `go run .` in the root directory of the project to run the api.

## Tech overview

This project uses:

- Go as the language
- Mariadb for database
- GORM as the ORM (currently no migration tracking. Any needed data backpops/fixes are done after deployment via endpont)
- No framework except chi routers and some other dependencies for auth. Mostly raw go.
- Uses built in test runner
- Uses OpenAPI 3.0, maintained by hand to generate clients. Example command to generate client: `java -jar swagger-codegen-cli.jar generate -i ./receipt-wrangler-api/swagger.yml -o ./receipt-wrangler-core/projects/core/src/lib/api/ -l typescript-angular`
  `
