# Manafield Web

Official lightweight Web Shell for Manafield.

## v0 goal

The first milestone intentionally stays small:

- Go single-binary HTTP server
- `GET /manafield/health`
- read `GET /modules` from Manafield Core
- render the instance home page
- list registered Module name, ID, description, and version
- keep serving a degraded home page when Core Registry is unavailable
- no database
- no Docker socket
- no mandatory reverse proxy for Module traffic

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | HTTP listen port |
| `MANAFIELD_WEB_ADDR` | empty | Full listen address override, e.g. `0.0.0.0:8080` |
| `MANAFIELD_CORE_URL` | `http://127.0.0.1:8080` | Manafield Core base URL |

## Run

```bash
go run ./cmd/manafield-web
```

Or with Docker:

```bash
docker build -t manafield-web:local .
docker run --rm -p 18082:8080 \
  -e MANAFIELD_CORE_URL=http://host.docker.internal:18080 \
  manafield-web:local
```

On Linux, when Core is on the host, use an appropriate host gateway or run both services in the same Docker network.

## Architecture

`manafield-web` is a presentation Module, not a mandatory gateway.

```text
Browser
  ↓
Ingress
  ├─ /          → manafield-web
  ├─ /echo/*    → Echo
  └─ /account/* → Account
```

Module traffic does not need to pass through this process.
