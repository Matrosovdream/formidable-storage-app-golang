# formidable-storage-app-golang

Go service that exposes the Formidable Storage HTTP API: Sanctum-style user
auth, per-site bearer tokens, CRUD over sites and form-entry data, bulk
fixture generation, and a Redis-backed work queue. Two binaries — `web` (the
Fiber HTTP server) and `worker` (background processor) — share a single
domain layer in `internal/`.

## Stack

- **Language:** Go 1.26
- **HTTP:** [Fiber v2](https://github.com/gofiber/fiber)
- **DB:** PostgreSQL 16 (pgx driver)
- **Cache / queue:** Redis 7
- **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate) (SQL files in [db/migrations/](db/migrations/))
- **Config:** Viper, fed from a `.env` file
- **Container:** multi-stage Dockerfile → distroless runtime

API contract: [internal/docs/openapi.yaml](internal/docs/openapi.yaml).

## Layout

```
cmd/
  web/       HTTP server entry point
  worker/    background worker entry point
  seed/      one-shot seeder for dev/demo data
internal/
  config/    viper-based env loading
  delivery/  HTTP handlers, middleware, routes
  usecase/   business logic
  repository/ DB access
  entity/    domain types
  model/     request/response DTOs
  gateway/   external service clients
db/migrations/   SQL up/down files
docker/    Dockerfile lives at repo root; this folder is reserved
```

---

## Dev install

Requires Docker + Docker Compose. Code is mounted into the container and
hot-reloaded with [air](https://github.com/air-verse/air), so you edit
locally and the running server picks it up.

```bash
git clone <repo>
cd formidable-storage-app-golang
cp .env.example .env          # defaults work out of the box for dev
make dev                      # docker compose -f compose.dev.yaml up --build
```

Services started:
- `web` on http://localhost:8080 (hot-reload)
- `worker` (hot-reload)
- `postgres` on `127.0.0.1:5432` (creds from `.env`)
- `redis` on `127.0.0.1:6379`
- `adminer` on http://localhost:8081 (point at host `postgres`)

Common commands:
```bash
make dev-logs                 # tail web + worker logs
make dev-down                 # stop the stack (keeps volumes)
make seed                     # populate demo data — prints a site bearer token
make test                     # go test ./...
```

Migrations run automatically when the dev stack boots. Add a new migration
by dropping `db/migrations/NNNN_name.up.sql` + `.down.sql` and restarting
the stack.

---

## Prod install

Designed to run behind [Traefik](https://traefik.io/) on a Linux host.
Postgres and Redis stay on the internal compose network — only the `web`
service is exposed, and only via Traefik with Let's Encrypt TLS.

### Host prerequisites

- Docker + Compose plugin
- A running Traefik container attached to an external network named
  `traefik-network`, with a `letsencrypt` cert resolver and a `websecure`
  entrypoint (those exact names — they're referenced in compose labels).
- DNS pointing your domain at the host
- ≥ 4 GB RAM **plus 4 GB swap** (Go builds are RAM-hungry; without swap a
  2 vCPU / 4 GB VM will OOM during the build)

Add swap once:
```bash
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile && sudo mkswap /swapfile && sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

### First-time deploy

```bash
git clone <repo> /opt/apps/formidable-storage-app-golang
cd /opt/apps/formidable-storage-app-golang

# Create .env from the example, then edit for production:
#   APP_ENV=production
#   APP_DEBUG=false
#   APP_URL=https://your-domain
#   APP_DOMAIN=your-domain                 (used by Traefik labels)
#   DB_PASSWORD=<strong random>
#   CORS_ALLOWED_ORIGINS=https://your-spa-domain
cp .env.example .env
nano .env

docker network create traefik-network      # if it doesn't exist yet
make prod                                  # builds web, then worker, then `up -d`
```

`make prod` always builds services sequentially so small VMs can compile
without OOM. Combined with BuildKit cache mounts in the Dockerfile, this
makes the worker build (after web) and every subsequent redeploy fast.

The `migrate` service runs once on boot and applies any pending
migrations — verify with `docker compose -f compose.prod.yaml logs migrate`.

### Update an already-deployed instance

```bash
cd /opt/apps/formidable-storage-app-golang
git pull
make prod
```

That rebuilds the changed images sequentially, recreates `web`/`worker`,
and re-runs migrations. Brief `web` downtime (~1–2 s) during recreate.

### Seed production data (optional, idempotent)

```bash
make seed
```

Detects the running stack and runs the seeder. Safe to re-run — it upserts.

### Useful prod commands

```bash
docker compose -f compose.prod.yaml ps                       # status
docker compose -f compose.prod.yaml logs -f web worker       # tail logs
make prod-down                                               # stop (keeps volumes)

# Manual migration control
docker compose -f compose.prod.yaml run --rm migrate \
  -path=/migrations \
  -database="postgres://${DB_USERNAME}:${DB_PASSWORD}@postgres:5432/${DB_DATABASE}?sslmode=disable" \
  version
```

---

## Configuration

All runtime config comes from `.env` (loaded by viper). Full list of
recognized variables and defaults: [internal/config/config.go](internal/config/config.go).
The two environment-specific values you must set in prod are `APP_DOMAIN`
(Traefik routing) and `CORS_ALLOWED_ORIGINS` (set to the SPA origin, not `*`).
