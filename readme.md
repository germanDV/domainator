# Domainator

Keep Track Of Your Domains, The Easy Way.

Domainator is a simple tool to that helps you keep track of your domains by monitoring the expiration of TLS certificates and notifying you before they expire.

## Env Vars

Make a copy of `.env.test` and name it `.env`. Replace the values within it.

Make sure to keep `.env.test` as it used in the tests.

## 3rd party tools

In addition to `go` and `make`. You will need:

- `air` for hot reloading
- `tern` for database migrations
- `templ` for html templating

You may install them with `make deps`.
