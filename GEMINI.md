# Project Overview

This project is the backend API for Receipt Wrangler, a free and open-source receipt management and splitting application. It is written in Go and provides a RESTful API for managing receipts, users, groups, and other resources.

The API uses a PostgreSQL database by default, but can also be configured to use MySQL or SQLite. It uses `go-chi` for routing, `gorm` for database access, and `gosseract` for OCR. The API is documented using the OpenAPI specification in the `swagger.yml` file.

# Building and Running

To build the application, run the following commands:

```bash
go build .
```

To run the tests, use the following commands:

```bash
go test -coverprofile=coverage.out -covermode=atomic -v ./...
python3 -m unittest discover -s ./imap-client
```

# Development Conventions

The project uses a number of development conventions, including:

*   **Gitflow:** The project uses the Gitflow workflow for managing branches.
*   **Conventional Commits:** The project uses the Conventional Commits specification for commit messages.
*   **Go Modules:** The project uses Go Modules for dependency management.
*   **Testing:** The project uses the `testing` package for unit tests.
*   **Linting:** The project uses `golangci-lint` for linting.
