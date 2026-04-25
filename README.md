# Travelling Zip

Zip-drone delivery scheduler. Go API backend + React frontend.

## Prerequisites

- Go 1.26+
- Node 24+ (use `nvm use` to pick up `.nvmrc`)

## Quick Start

In one terminal, start the API (port 3001):

```bash
make run-backend
```

In a second terminal, install deps (first run only) and start the frontend (port 5173):

```bash
make install-frontend
make run-frontend
```

Open <http://localhost:5173>.

## Project Structure

```
backend/         Go module (travelingzip)
  cmd/api/         HTTP REST server
  cmd/simulator/   CLI batch runner
  core/            Scheduling business logic
data/            Shared CSV fixtures (hospitals, orders)
frontend/        React 19 + Vite 8 single-page app
```

> The Go module is named `travelingzip` (no double-l) and is referenced by all internal imports as `travelingzip/core`. The repo directory name (`travelling_zip`) is intentionally independent of the module name.

## API

- `GET /health` → `{"status":"ok","implementation":"go"}`
- `POST /api/simulation` with `{"config":{...}}` → simulation snapshot

## Makefile Targets

| Command                 | Description                    |
|-------------------------|--------------------------------|
| `make run-backend`      | Go API on port 3001            |
| `make run-frontend`     | Vite dev server on port 5173   |
| `make run-simulator`    | CLI simulator one-shot         |
| `make install-frontend` | `npm install` in `frontend/`   |
| `make build`            | Compile backend + frontend     |
| `make test`             | Run all tests (Go + frontend)  |
| `make test-backend`     | Run Go tests only              |
| `make test-frontend`    | Run frontend tests (Vitest)    |
| `make help`             | List all targets               |
