# sbcntr-batch

Batch processing application for the backend API from the book "AWS Container Design and Construction [Professional] Introduction 2nd Edition".

## Overview

A Golang-based batch processing service using the echo framework.
The connection between the API server and DB (Postgres) uses sqlx[^sqlx], an O/R mapper library.

[^sqlx]: <https://jmoiron.github.io/sqlx/>

This batch processing provides the following services:

1. Reservation Batch Processing
   - Process pending reservations
   - Check for duplicate reservations
   - Update reservation status
2. Notification Batch Processing
   - Generate notifications

## Usage

Please use this application according to the contents of the book.

## Local Usage

### Prerequisites

- Go version 1.23.x is required.
- Clone this repository code to the appropriate directory according to your GOPATH location.
- Download modules using the following commands:

```bash
go get golang.org/x/lint/golint
go install
go mod download
```

- This backend API requires DB connection. Set the following environment variables for DB connection:
  - DB_HOST
  - DB_USERNAME
  - DB_PASSWORD
  - DB_NAME
  - DB_CONN

### Database Setup

Please start a Postgres server locally beforehand.

### Build & Deploy

#### Running Locally

```text
export DB_HOST=localhost
export DB_USERNAME=sbcntrapp
export DB_PASSWORD=password
export DB_NAME=sbcntrapp
export DB_CONN=1
```

```bash
make all
```

#### Running with Docker

```bash
$ docker build -t sbcntr-batch:latest .
$ docker images
REPOSITORY                      TAG                 IMAGE ID            CREATED             SIZE
sbcntr-batch      latest              cdb20b70f267        58 minutes ago      15.2MB
:
$ docker run -d -e DB_HOST=host.docker.internal \
              -e DB_USERNAME=sbcntrapp \
              -e DB_PASSWORD=password \
              -e DB_NAME=sbcntrapp \
              -e DB_CONN=1 \
              sbcntr-batch:latest
```

### Post-Deployment Verification

Check the batch processing logs to confirm it is working properly.

```bash
docker logs <container-id>
```

## Environment Variables

| Variable    | Description               | Default Value |
| ----------- | ------------------------- | ------------- |
| DB_HOST     | Database host             | localhost     |
| DB_PORT     | Database port             | 5432          |
| DB_USERNAME | Database username         | sbcntrapp     |
| DB_PASSWORD | Database password         | password      |
| DB_NAME     | Database name             | sbcntrapp     |
| DB_CONN     | Database connection pool  | 1             |

## Development Commands

- `make all`: Build and run tests
- `make build`: Build the application
- `make test`: Run tests
- `make test-coverage`: Run tests with coverage
- `make validate`: Validate code
- `make clean`: Clean build artifacts
- `make install-tools`: Install development tools

## Notes

- Tested only on Mac OS Sequoia 15.6.

## License

Apache License 2.0