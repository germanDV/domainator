# Domainator

Keep Track Of Your Domains, The Easy Way.

Domainator is a simple tool to that helps you keep track of your domains by providing two services:

- TLS Certs Expiration Monitoring
- Healthcheck Endpoint Monitoring

## Language / Entities

There are two main entities:

- **Cert**
- **Endpoint**

A **Cert** represents a domain whose TLS certificate we will check.
They are stored in the `certs` table.

Each time we check a TLS certificate, we call it a **Check** and we store the results
in the `certchecks` table.

An **Endpoint** represents a URL that we will ping and check it responds with the expected HTTP status.
They are stored in the `endpoints` table.

Each time we ping an endpoint, we call it a **Healthcheck** and we store it in the
`healthchecks` table.

## Env Vars

Make a copy of `.env.test` and name it `.env`. Replace the values within it.

Make sure to keep `.env.test` as it used in the tests.

## 3rd party tools

In addition to `go`, `docker` and `make`. You will need:

- `air` for hot reloading
- `tern` for database migrations

You may install them with `make deps`.
