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

## vuln: check for vulnerabilities
.PHONY: vuln
vuln:
	govulncheck ./...

## audit: tidy dependencies, format and vet
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	@echo 'Checking dependencies for vulnerabilities...'
	govulncheck ./...

## dev: run with hot-reloading
.PHONY: dev
dev:
	air .

## templgen: generate templates
.PHONY: templgen
templgen:
	@echo 'Generating templates...'
	templ generate

## build: build binary
.PHONY: build
build: templgen
	@echo 'Building for Linux'
	go build -o=./bin/${BINARY_NAME} ./cmd/web

## db/up: start PostgreSQL docker container by running docker-compose.yml
.PHONY: db/up
db/up:
	@echo 'Starting PostgreSQL docker container'
	docker compose up -d

## db/stop: stop PostgreSQL docker container
.PHONY: db/stop
db/stop:
	@echo 'Stopping PostgreSQL docker container'
	docker compose stop

## db/down: tear down PostgreSQL docker container
.PHONY: db/down
db/down: confirm
	@echo 'Stopping PostgreSQL docker container'
	docker compose down

## db/migrate/init: init tern project
.PHONY: db/migrate/init
db/migrate/init: confirm
	@echo 'Initializing tern project'
	tern init

## db/migrate/new name=$1: create a new database migration ($ make db/migrate/new name=create_users_table)
.PHONY: db/migrate/new
db/migrate/new:
	@echo 'Creating migration files for ${name}...'
	tern new -m ./migrations ${name}

## db/migrate/up: run database migrations
.PHONY: db/migrate/up
db/migrate/up:
	@echo 'Running migrations...'
	@PG_PASSWORD=${PG_PASSWORD} tern migrate -m ./migrations

## db/migrate/down n=$1: rollback database N versions ($ make db/migrate/down n=2)
.PHONY: db/migrate/down
db/migrate/down: confirm
	@echo 'Rolling back ${n} migrations..'
	@PG_PASSWORD=${PG_PASSWORD} tern migrate -m ./migrations --destination -${n}

## deps/upgrade/all: upgrade all dependencies
.PHONY: deps/upgrade/all
deps/upgrade/all:
	@echo 'Upgrading dependencies to latest versions...'
	go get -t -u ./...

## deps/upgrade/patch: upgrade dependencies to latest patch version
.PHONY: deps/upgrade/patch
deps/upgrade:
	@echo 'Upgrading dependencies to latest patch versions...'
	go get -t -u=patch ./...

## deps/ext: install external dependencies not used in source code
.PHONY: deps/ext
deps/ext: confirm
	@echo 'Installing `air` for hot-reloading'
	go install github.com/cosmtrek/air@latest
	@echo 'Installing `tern` for db migrations'
	go install github.com/jackc/tern/v2@latest
	@echo 'Installing `templ` for html templating'
	go install github.com/a-h/templ/cmd/templ@latest
