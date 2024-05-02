BINARY_NAME=domainator

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
	go test ./...

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
	go build -ldflags "-s -w" -tags prod -o=./bin/${BINARY_NAME} ./cmd/web

## docker/up: start PostgreSQL + Redis docker containers
.PHONY: docker/up
docker/up:
	@echo 'Starting docker-compose'
	docker compose up -d

## docker/stop: stop docker containers
.PHONY: docker/stop
docker/stop:
	@echo 'Stopping docker-compose'
	docker compose stop

## docker/down: tear down PostgreSQL docker container
.PHONY: docker/down
docker/down: confirm
	@echo 'Stopping docker-compose'
	docker compose down

## db/migrate/up: run database migrations
.PHONY: db/migrate/up
db/migrate/up:
	@echo 'Running migrations...'
	@go run ./cmd/migrate -action up

## db/migrate/down: rollback latest database migration
.PHONY: db/migrate/down
db/migrate/down: confirm
	@echo 'Rolling back latest migration..'
	@go run ./cmd/migrate -action down

## db/cli: connect to local database using pgcli
.PHONY: db/cli
db/cli:
	@echo 'Connecting to database...'
	pgcli -h localhost -p 5432 -U postgres -d domainator

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
	@echo 'Installing `templ` for html templating'
	go install github.com/a-h/templ/cmd/templ@latest

## worker/run: run worker
.PHONY: worker/run
worker/run:
	@echo 'Running worker...'
	go run ./cmd/worker

## worker/build: build worker
.PHONY: worker/build
worker/build:
	@echo 'Building worker binary...'
	go build -o=./bin/${BINARY_NAME}_worker ./cmd/worker

## scripts/keys: generate new key-pair
.PHONY: scripts/keys
scripts/keys:
	go run ./cmd/keys

## scripts/token u=$1: generate an auth token ($ make scripts/token u=<user_id>)
.PHONY: scripts/token
scripts/token:
	go run ./cmd/token "$u"
