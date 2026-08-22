# go-mvc-app

A Go REST API built with Gin, GORM (PostgreSQL), Zap logging, JWT auth, and Swagger docs — following an MVC structure.

---

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.25+ | https://go.dev/dl |
| PostgreSQL | 16+ | https://www.postgresql.org/download or via Docker |
| `swag` CLI | latest | `make install-swag` |
| `air` CLI | latest | `make install-air` *(only for live-reload dev)* |
| User Service | reachable over HTTP | this service resolves `user_id` against it on every booking read/write |

---

## 1 — Clone & install dependencies

```bash
git clone <repo-url>
cd <project-name>

# Download Go modules
make tidy
```

---

## 2 — Configure environment

```bash
cp .env.example .env
```

Open `.env` and fill in the required values:

```env
APP_ENV=dev          # dev | qa | prod
APP_PORT=8080

# PostgreSQL — required
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=gomvc_dev

# External API — required
EXTERNAL_API_BASE_URL=https://api.example.com
EXTERNAL_API_KEY=your-api-key

# JWT — required
JWT_SECRET=your-jwt-secret

# User Service — required; booking reads/writes resolve user_id against it
USER_SERVICE_BASE_URL=http://localhost:8081

# CORS — comma-separated origins, or * for dev
CORS_ORIGINS=*

# Log sinks: files | console | files,console
LOGS_TARGETS=files,console
```

> Optional overrides (`DB_PORT`, `DB_SSLMODE`, `DB_TIMEZONE`, `EXTERNAL_API_TIMEOUT`, `JWT_EXPIRY_HOURS`, `USER_SERVICE_TIMEOUT`) are commented out in `.env.example` — uncomment to use.
>
> `USER_SERVICE_BASE_URL` must point to a running instance of the User Service — booking endpoints call `GET {USER_SERVICE_BASE_URL}/api/v1/users/{id}` to validate `user_id` before reading/writing a booking.

---

## 3 — Run database migrations

Migrations do **not** run automatically when the server starts — they're a separate, explicit step (`database.Migrate()` only runs behind the `--migrate` flag in `src/cmd/main.go`). Run this once after Postgres is up and `.env` is configured, and again any time models change.

```bash
make migrate
# equivalent to:
go run ./src/cmd/main.go --migrate
```

This connects to the DB, runs `AutoMigrate` for every model registered in `src/utils/database/database.go`, then exits — no port is bound, no HTTP server starts.

> **Docker:** the `app` container's `CMD` runs the server directly, not `--migrate`. Run migrations as a one-off container against the same env before (or after) `docker compose up`:
> ```bash
> docker compose run --rm app ./server --migrate
> ```

---

## 4 — Run the app

Pick one path — both end with the API on `localhost:8080`.

### Option A — With Docker (app + db together)

No local Go toolchain, Postgres, or `swag`/`air` install needed.

```bash
docker compose up --build
```

The `app` service reads `.env` automatically (`DB_HOST` is overridden to `db` internally) and waits for Postgres to be healthy before starting.

> `USER_SERVICE_BASE_URL` still needs to point somewhere reachable from inside the container — `http://localhost:8081` won't resolve from `app`. Point it at the User Service's container/network address (e.g. `http://host.docker.internal:8081` if it runs on your host, or its service name if it's on the same Docker network).

### Option B — Without Docker (local Go toolchain)

**a. Start PostgreSQL**

Docker for just the DB (recommended):

```bash
docker compose up db -d
```

Starts Postgres 16 on port `5432` with DB `gomvc_dev` / user `postgres` / password `postgres`. Update `.env` to match if you changed those values.

Or, fully local (no Docker at all) — create the database manually:

```bash
psql -U postgres -c "CREATE DATABASE gomvc_dev;"
```

**b. Install Go tooling** (skip if already done)

```bash
make tidy
make install-swag
make install-air     # only needed for live-reload
```

**c. Run the server**

```bash
make dev             # one-shot: regen swagger + build + run (dev env)
make watch           # live-reload — rebuilds on every save
make qa              # APP_ENV=qa
make prod            # APP_ENV=prod
```

Or build a standalone binary:

```bash
make build
./dist/server
```

---

## 5 — Verify it's running

```bash
curl http://localhost:8080/api/v1/health
```

Swagger UI is available at:

```
http://localhost:8080/swagger/index.html
```

---

## Project structure

```
src/
  cmd/            # Entry point (main.go)
  config/         # Env loading, common constants
  clients/        # HTTP clients for sibling services (e.g. User Service)
  controllers/    # HTTP handlers
  services/       # Business logic
  repositories/   # DB layer (GORM)
  models/         # GORM models
  routes/         # Route registration
  middleware/     # Auth, logging, CORS
  helpers/        # Shared utilities
  utils/          # Low-level helpers
docs/             # Auto-generated Swagger files
logs/             # Log output (when LOGS_TARGETS=files)
```

---

## Database schema

Two schemas currently live side by side in the same Postgres database:

- **`bookings`** — the original single-sided booking table (`src/models/booking.go`), wired to the live `/api/v1/bookings` API.
- **Marketplace schema** — a fuller two-sided hiring-platform design (clients, employers/agencies, employees, categories, matching, payments, reviews, KYC) translated from `services-marketplace-hld.docx` into GORM models under `src/models/`. Not yet wired to any API. Tables: `marketplace_users`, `addresses`, `categories`, `employers`, `employees`, `clients`, `employee_categories`, `category_pricing`, `automation_jobs`, `service_requests`, `booking_assignment_attempts`, `payments`, `reviews`, `employee_availability`, `kyc_verifications`, `kyc_documents`.

> **Why `marketplace_users` and not `users`:** this database already has a live `users` table (UUID id, name/email/password/role) owned by the existing User Service (`src/clients/user_http_client.go`) — real rows, not to be touched. The marketplace design's own `users` table (phone-OTP, bigint id) is a different, incompatible shape for the same concept, so it's kept as a separate table until the two are reconciled.

See `docs/HLD-service-hiring-platform.md` for the extension design and rollout phasing.

---

## Makefile targets

| Target | Description |
|--------|-------------|
| `make tidy` | Download / tidy Go modules |
| `make dev` | Regen Swagger + run (dev) |
| `make watch` | Live-reload dev server |
| `make qa` | Run with `APP_ENV=qa` |
| `make prod` | Run with `APP_ENV=prod` |
| `make build` | Compile binary to `dist/server` |
| `make migrate` | Run DB migrations only, then exit (no server, no port bind) |
| `make swagger` | Regenerate Swagger docs only |
| `make install-swag` | Install `swag` CLI (one-time) |
| `make install-air` | Install `air` CLI (one-time) |
