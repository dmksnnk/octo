# octo

OCTO is a Go HTTP API.

## API Documentation

The API documentation, available in OpenAPI format, can be found in the [docs](./docs/openapi.yaml) folder. You can also view the rendered documentation at the `/docs` endpoint.

## Development

This application uses [sqlc](https://github.com/sqlc-dev/sqlc) to generate Go code from SQL queries. Run `make sqlc-generate` to generate the database query code.

Database migrations are managed using [goose](https://github.com/pressly/goose). Migration files are located in the [migrations](./migrations/) folder. Use `make goose-up` to apply migrations and `make goose-down` to roll them back.

For isolated PostgreSQL testing, the application uses [pgtestdb](https://github.com/peterldowns/pgtestdb).

To generate mocks for testing, the application uses [mockery](https://github.com/vektra/mockery).

### Running the Application Locally

To start the application locally, run:

```sh
make up
```

It will be accessible at `http://localhost:8080`.


To shutdown use:

```sh
make down
```

### Generating fake data

To generate fake data and insert it to the database, use:

```sh
make generate-data
```

See [fake](./cmd/fake/main.go) for CLI options.

### Running Tests

To execute tests, use:

```sh
make test
```

### Liting

Run [golangci-lint](https://golangci-lint.run/) using:

```sh
make golangci
```

### Running Infrastructure Locally

To start the required infrastructure (e.g., PostgreSQL) for testing, run:

```sh
make infra
```

To shutdown:

```sh
make down
```

