## Development
To run the api for local development:
1. Clone the repository
2. Install go on your OS. See https://go.dev/doc/install for more details.
3. Set up a mariadb instance. The easiest way to do this is via docker. F.ex: ``` docker run --name receipt-wrangler-db -e MYSQL_ROOT_PASSWORD=123456 -e MYSQL_USER=wrangler -e MYSQL_PASSWORD=123456 -e MYSQL_DATABASE=wrangler -p 9001:3306  -d library/mariadb ```
4. Add a config and a feature config, copy pasta the samples and modify as needed
5. Once configs are added, and db is up, run ``` go run . ``` in the root directory of the project to run the api
   
## Tech overview
This project uses:
* Go as the language
* Mariadb for database
* GORM as the ORM (currently no migration tracking. Any needed data backpops/fixes are done after deployment via endpont)
* No framework except chi routers and some other dependencies for auth. Mostly raw go.
* Eventually will add OpenAPI specs, and unit tests 
