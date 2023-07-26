BINARY_NAME=domainator
PG_PASSWORD ?= pass123

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]

## test: run tests
.PHONY: test
test:
	ENV_FILENAME=.env.test go test ./...

## test/race: run tests with race detector
.PHONY: test/race
test/race:
	ENV_FILENAME=.env.test go test -race ./...

## audit: tidy dependencies, format, vet and test
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	@echo 'Running tests...'
	ENV_FILENAME=.env.test go test -race -vet=off ./...

## dev: run with hot-reloading
.PHONY: dev
dev:
	air .

## build: build binary
.PHONY: build
build:
	@echo 'Building for Linux'
	go build -o=./bin/${BINARY_NAME} ./cmd/web

## pg/up: start PostgreSQL docker container by running docker-compose.yml
.PHONY: pg/up
pg/up:
	@echo 'Starting PostgreSQL docker container'
	docker compose up -d

## pg/stop: stop PostgreSQL docker container
.PHONY: pg/stop
pg/stop:
	@echo 'Stopping PostgreSQL docker container'
	docker compose stop

## pg/down: tear down PostgreSQL docker container
.PHONY: pg/down
pg/down: confirm
	@echo 'Stopping PostgreSQL docker container'
	docker compose down

## pg/migrate/init: init tern project
.PHONY: pg/migrate/init
pg/migrate/init: confirm
	@echo 'Initializing tern project'
	tern init

## pg/migrate/new name=$1: create a new database migration ($ make pg/migrate/new name=create_users_table)
.PHONY: pg/migrate/new
pg/migrate/new:
	@echo 'Creating migration files for ${name}...'
	tern new -m ./migrations ${name}

## pg/migrate/up: run database migrations
.PHONY: pg/migrate/up
pg/migrate/up:
	@echo 'Running migrations...'
	@PG_PASSWORD=${PG_PASSWORD} tern migrate -m ./migrations

## pg/migrate/down n=$1: rollback database N versions ($ make pg/migrate/down n=2)
.PHONY: pg/migrate/down
pg/migrate/down: confirm
	@echo 'Rolling back ${n} migrations..'
	@PG_PASSWORD=${PG_PASSWORD} tern migrate -m ./migrations --destination -${n}
	
## deps: install external dependencies not used in source code
.PHONY: deps
deps: confirm
	@echo 'Installing `air` for hot-reloading'
	go install github.com/cosmtrek/air@latest
	@echo 'Installing `tern` for db migrations'
	go install github.com/jackc/tern/v2@latest

