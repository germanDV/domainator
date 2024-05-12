# Domainator

Keep Track Of Your Domains, The Easy Way.

Domainator is a simple tool to that helps you keep track of your domains by monitoring the expiration of TLS certificates and notifying you before they expire.

## Env Vars

Make a copy of `.env.test` and name it `.env`. Replace the values within it.

## 3rd party tools

In addition to `go` and `make`. You will need:

- `air` for hot reloading
- `templ` for html templating

You may install them with `make deps`.

To run `make lint`, you will need [golangci-lint](golang.org/x/lint/golint).

## Components

Domainator consists of two components:
- **Server**: the webserver.
- **Worker**: a background worker that updates certificates data, meant to be run as a cron job.
