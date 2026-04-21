# Service Wait

A lightweight CLI tool that waits for HTTP, MongoDB, or PostgreSQL endpoints to become available before proceeding. Useful as an init container or entrypoint wrapper in containerised deployments.

## Features

- **HTTP** – probe one or multiple HTTP endpoints
- **PostgreSQL** – wait for a Postgres database to accept connections
- **MongoDB** – wait for a Mongo instance to respond to pings
- Configurable timeout and polling interval
- Configuration via CLI flags or environment variables

## Install

### Prerequisites

- [mise](https://mise.jdx.dev)

```bash
mise trust
mise install
```

### Build from source

```bash
make build        # outputs bin/service-wait
make install      # installs to ~/.local/bin
```

### Docker

```bash
docker build -t service-wait .
```

## Usage

```bash
service-wait [flags]
```

### Flags

| Flag | Short | EnvVar | Default | Description |
|------|-------|--------|---------|-------------|
| `--url` | `-u` | `SERVICE_WAIT_URL` | | HTTP endpoint to probe |
| `--urls` | `-U` | `SERVICE_WAIT_URLS` | | Multiple HTTP endpoints (comma-separated) |
| `--timeout` | `-t` | `SERVICE_WAIT_TIMEOUT` | `30s` | Max time to wait |
| `--interval` | `-i` | `SERVICE_WAIT_INTERVAL` | `30s` | Polling interval |
| `--psql-host` | `-H` | `SERVICE_WAIT_PSQL_HOST` / `PGHOST` | | PostgreSQL host |
| `--psql-port` | `-p` | `SERVICE_WAIT_PSQL_PORT` / `PGPORT` | `5432` | PostgreSQL port |
| `--psql-user` | | `SERVICE_WAIT_PSQL_USER` / `PGUSER` | | PostgreSQL user |
| `--psql-password` | `-P` | `SERVICE_WAIT_PSQL_PASSWORD` / `PGPASSWORD` | | PostgreSQL password |
| `--psql-database` | `-d` | `SERVICE_WAIT_PSQL_DATABASE` / `PGDATABASE` | | PostgreSQL database |
| `--psql-sslmode` | `-s` | `SERVICE_WAIT_PSQL_SSLMODE` / `PGSSLMODE` | `disable` | PostgreSQL SSL mode |
| `--psql-dsn` | | `SERVICE_WAIT_PSQL_DSN` | | PostgreSQL DSN (overrides individual flags) |
| `--mongo-host` | | `SERVICE_WAIT_MONGO_HOST` | | MongoDB host |
| `--mongo-port` | | `SERVICE_WAIT_MONGO_PORT` | `27017` | MongoDB port |
| `--mongo-user` | | `SERVICE_WAIT_MONGO_USER` | | MongoDB user |
| `--mongo-password` | | `SERVICE_WAIT_MONGO_PASSWORD` | | MongoDB password |
| `--mongo-database` | | `SERVICE_WAIT_MONGO_DATABASE` | | MongoDB database |
| `--mongo-auth-source` | | `SERVICE_WAIT_MONGO_AUTH_SOURCE` | | MongoDB auth source |
| `--verbose` | `-V` | `SERVICE_WAIT_DEBUG` | `false` | Enable debug logging |

### Examples

Wait for an HTTP endpoint:

```bash
service-wait --url http://localhost:8080/health --timeout 60s --interval 2s
```

Wait for PostgreSQL:

```bash
service-wait --psql-host localhost --psql-user postgres --psql-password secret --psql-database mydb
```

Wait for MongoDB:

```bash
service-wait --mongo-host localhost --mongo-user admin --mongo-password secret --mongo-database mydb
```

Using environment variables:

```bash
export SERVICE_WAIT_URL=http://localhost:8080/health
export SERVICE_WAIT_TIMEOUT=60s
service-wait
```

## Development

```bash
make build    # compile
make test     # run tests (requires Docker for testcontainers)
make fmt      # format code
make vet      # run go vet
make tidy     # tidy go modules
```