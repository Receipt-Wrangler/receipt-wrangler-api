[![Build](https://github.com/Noah231515/receipt-wrangler-api/actions/workflows/docker-image.yml/badge.svg)](https://github.com/Noah231515/receipt-wrangler-api/actions/workflows/docker-image.yml) [![codecov](https://codecov.io/gh/Receipt-Wrangler/receipt-wrangler-api/graph/badge.svg?token=EUQMLBEKPK)](https://codecov.io/gh/Receipt-Wrangler/receipt-wrangler-api)

## Development

To run the api for local development:

1. Clone the repository
2. Install GO on your OS. See https://go.dev/doc/install for more details
3. Install tesseract dependencies. If you are running a debian derivative (recommended), run `sudo sh set-up-tesseract-env.sh`, otherwise see https://github.com/otiai10/gosseract for other installation details.
4. Install OpenAPI generator https://openapi-generator.tech/docs/installation (I personally use the jar installation)
5. Set up a development db instance (mariadb/mysql/postgresql/sqlite). The dev directory contains scripts to set some env variables needed for each db flavor. An easy way to set up a mariadb instance is: `docker run --name receipt-wrangler-db -e MYSQL_ROOT_PASSWORD=123456 -e MYSQL_USER=wrangler -e MYSQL_PASSWORD=123456 -e MYSQL_DATABASE=wrangler -p 9001:3306  -d library/mariadb`. 
6. Add a config.dev.json and a feature-config.dev.json, copy pasta the samples and modify as needed.
7. Once configs are added, and db is up, run `go run .` in the root directory of the project to run the api.

## Tech overview

This project uses:

- Go for the main API, Python for the imap client
- mariadb/mysql/postgresql/sqlite
- GORM as the ORM (currently no migration tracking. Any needed data backpops/fixes are done after deployment via endpoint)
- Overall, no framework is used for the API
- Uses built in test runner
- Uses OpenAPI 3.0, maintained by hand to generate clients. Example command to generate client: `java -jar swagger-codegen-cli.jar generate -i ./receipt-wrangler-api/swagger.yml -o ./receipt-wrangler-core/projects/core/src/lib/api/ -l typescript-angular`
  ` or use generate-core-swagger.sh
