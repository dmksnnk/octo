BUILD_DIR ?= ./bin
current_dir = $(shell pwd)
version = $(shell git rev-parse --short HEAD)
image = ghcr.io/dmksnnk/octo:$(version)

.PHONY: build
build:
	@echo "Building into $(BUILD_DIR)"
	GOOS=linux GOARCH=amd64 go build -v -o $(BUILD_DIR) ./cmd/...

.PHONY: build-docker
build-docker:
	@docker build --platform linux/amd64 -f Dockerfile --tag=$(image) .

.PHONY: clean
clean:
	rm -r $(BUILD_DIR)

.PHONY: golangci
golangci: # run golangci-lint
	@docker run --rm \
		-v "$$(go env GOPATH)/pkg:/go/pkg" \
    	-e "GOCACHE=/cache/go" \
    	-v "$$(go env GOCACHE):/cache/go" \
		-e "GOLANGCI_LINT_CACHE=/.golangci-lint-cache" \
		-v "$(current_dir)/.cache:/.golangci-lint-cache" \
		-v $(current_dir):/app \
		-w /app \
		golangci/golangci-lint:v1.64.8-alpine@sha256:ae6460f78db54f22838d2a8aee0f2eaa4f785d5a01f638600072b60848f8deb4 \
		golangci-lint run --timeout 2m

.PHONY: test
test: infra
	@DATABASE_URL=postgres://master:mysecretpassword@localhost:1234/postgres?sslmode=disable go test -race -v -count 1 -timeout 30s ./...

.PHONY: up
up:
	@docker compose up -d --build

.PHONY: down
down:
	@docker compose down --remove-orphans

.PHONY: infra
infra: # runs required infrastructure
	@docker compose up -d postgres

.PHONY: goose-up
goose-up: infra # run migration to most recent version
	@goose -dir migrations postgres "postgres://master:mysecretpassword@localhost:1234/postgres?sslmode=disable" up


.PHONY: goose-down
goose-down: infra  # rollback migration by 1
	@goose -dir migrations postgres "postgres://master:mysecretpassword@localhost:1234/postgres?sslmode=disable" down

.PHONY: sqlc-vet
sqlc-vet:	## lint queries
	@docker run --rm \
		-v $(current_dir):/src \
		-w /src \
		--network host \
		sqlc/sqlc:v1.28.0@sha256:da028c7f0a30afd26cce5c3f7a097fa05b3319b71837fec38f2e9349aeede3e6 \
		vet --file queries/sqlc.yml

.PHONY: sqlc-generate
sqlc-generate: 	sqlc-vet ## generate SQL code with sqlc
	@docker run --rm \
		-v $(current_dir):/src \
		-w /src \
		sqlc/sqlc:v1.28.0@sha256:da028c7f0a30afd26cce5c3f7a097fa05b3319b71837fec38f2e9349aeede3e6 \
		generate --file queries/sqlc.yml

.PHONY: mockery-generate
mockery-generate: # generate mocks
	@docker run --rm \
		-v $(current_dir):/src \
		-w /src \
		vektra/mockery:v3.2@sha256:0aa6d2c8c121a7a0a97ef6f9cf4ec5bae8f507ad9a6d87987cdfbd9faf77304d \

.PHONY: generate-data
generate-data: # generate fake data
	@go run ./cmd/fake/... --database-url postgres://master:mysecretpassword@localhost:1234/postgres?sslmode=disable