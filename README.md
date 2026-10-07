# Manafield Web

Official lightweight Web Shell for Manafield.

## v0 goal

The first milestone intentionally stays small:

- Go single-binary HTTP server
- `GET /manafield/health`
- read `GET /modules` from Manafield Core
- render a Shell/status surface at the Instance-assigned base path
- list registered Module name, ID, description, and version
- keep serving a degraded Shell page when Core Registry is unavailable
- no database
- no Docker socket
- no mandatory reverse proxy for Module traffic
- no ownership requirement for the public root Homepage

Homepage content may be provided by a separate replaceable Module such as `manafield-home`.

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | HTTP listen port |
| `MANAFIELD_WEB_ADDR` | empty | Full listen address override, e.g. `0.0.0.0:8080` |
| `MANAFIELD_CORE_URL` | `http://127.0.0.1:8080` | Manafield Core base URL |
| `MANAFIELD_WEB_BASE_PATH` | `/` | Instance-assigned public base path, e.g. `/_manafield` |

The Web Shell preserves its assigned prefix. With `MANAFIELD_WEB_BASE_PATH=/_manafield`, the Shell page is served at `/_manafield/` and its assets use the same prefix.

The internal Health Operation remains available at `/manafield/health`. A prefixed Web route also exposes `/_manafield/manafield/health` when that base path is configured.

## Run

```bash
go run ./cmd/manafield-web
```

Or with Docker:

```bash
docker build -t manafield-web:local .
docker run --rm -p 18082:8080 \
  -e MANAFIELD_CORE_URL=http://host.docker.internal:18080 \
  -e MANAFIELD_WEB_BASE_PATH=/_manafield \
  manafield-web:local
```

On Linux, when Core is on the host, use an appropriate host gateway or run both services in the same Docker network.

## Architecture

`manafield-web` is a presentation Shell Module, not a mandatory gateway and not the required Homepage implementation.

```text
Browser
  ↓
Ingress
  ├─ /              → Homepage Module
  ├─ /_manafield/*  → manafield-web
  ├─ /echo/*        → Echo
  └─ /account/*     → Account
```

Module traffic does not need to pass through this process.
