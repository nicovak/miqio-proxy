# miqio-proxy

Stateless Go service that serves published landing pages from Cloudflare R2. It resolves tenant subdomains and custom domains via an in-memory cache backed by Postgres, applies URL normalization and redirections, then fetches HTML from R2.

## Architecture

```
Client Browser
  │
  ▼
Cloudflare (CDN + DNS *.miqio.page + WAF + DDoS)
  │
  │  cache HIT → response from edge
  │  cache MISS ▼
  │
Traefik (Gateway API + TLS + security headers)
  │
  ▼
miqio-proxy (stateless, horizontally scalable)
  ├── in-memory cache (tenants, domains, redirects — synced every 30s)
  ├── Cloudflare R2 (S3 API, bucket miqio-lp-assets)
  └── Postgres (source of truth)
```

## Request flow

1. Extract `Host` header (strip port)
2. Normalize URL — strip trailing slash, `.html`, `/index` → 301
3. Resolve tenant:
   - `{slug}.miqio.page` → lookup slug in tenant cache
   - Custom domain → lookup in domain cache (requires `ssl_active` status)
4. Check redirect cache — if match, return 301/302 with merged query params
5. Build R2 key: `{tenant-uuid}/index.html` or `{tenant-uuid}/{page-slug}/index.html`
6. Fetch from R2, return HTML with `Cache-Control: public, max-age=300`

## Prerequisites

- Go 1.26+
- Postgres
- Cloudflare R2 bucket with API credentials

## Configuration

Copy `.env.example` to `.env` and fill in your values:

| Variable | Description | Default |
|---|---|---|
| `ENV` | Environment label | — |
| `PORT` | HTTP listen port | `8080` |
| `POSTGRES_CONNECTION_STRING` | Postgres DSN | — |
| `R2_ACCOUNT_ID` | Cloudflare account ID | — |
| `R2_ACCESS_KEY_ID` | R2 S3 API access key | — |
| `R2_SECRET_ACCESS_KEY` | R2 S3 API secret key | — |
| `R2_BUCKET_NAME` | R2 bucket name | `miqio-lp-assets` |
| `MIQIO_PAGE_DOMAIN` | Base domain for tenant subdomains | `miqio.page` |
| `CACHE_SYNC_INTERVAL` | In-memory cache refresh interval | `30s` |
| `LOG_LEVEL` | Logging level | `info` |
| `APP_ORIGIN` | Next.js app URL (tracking forward target) | `https://app.miqio.page` |

## Getting started

```bash
# Install tooling (macOS)
make setup

# Create and migrate databases
make db-setup

# Run the server
make dev
```

## Local development

The `.page` TLD is HSTS preloaded — browsers force HTTPS on all `.page` domains. For local browser testing, override `MIQIO_PAGE_DOMAIN` with a non-HSTS domain:

```bash
# .env
MIQIO_PAGE_DOMAIN=miqio.local
```

Add the tenant subdomain to `/etc/hosts`:

```
127.0.0.1 corail-lab-demo.miqio.local
```

Then browse to `http://corail-lab-demo.miqio.local:8080/`.

With curl you can use the real domain and set the Host header directly:

```bash
curl -H "Host: corail-lab-demo.miqio.page" http://localhost:8080/guide-proteger-les-oceans
```

### R2 object structure

The proxy expects HTML files in R2 under this key layout:

```
miqio-lp-assets/
  {tenant-uuid}/
    index.html                          ← root page (/)
    {page-slug}/
      index.html                        ← subpage (/{page-slug})
      assets/
        style.css
        hero.webp
```

Upload files with the AWS CLI:

```bash
TENANT_UUID="00000000-0000-4000-a000-000000000001"
R2_ENDPOINT="https://<account-id>.r2.cloudflarestorage.com"

aws s3 sync ~/path/to/pages/ s3://miqio-lp-assets/$TENANT_UUID/ \
  --endpoint-url "$R2_ENDPOINT"
```

## Commands

| Command | Description |
|---|---|
| `miqio-proxy server` | Start the HTTP proxy server |
| `miqio-proxy migrate` | Run database migrations |

## Makefile targets

| Target | Description |
|---|---|
| `make dev` | Run the server in dev mode |
| `make build` | Build the binary |
| `make setup` | Install and setup dependencies |
| `make db-create` | Create databases |
| `make db-migrate` | Run migrations |
| `make db-reset` | Drop, recreate, and migrate databases |
| `make db-up` | Migrate UP |
| `make db-down` | Migrate DOWN |
| `make db-setup` | Create and migrate databases |
| `make lint` | Run golangci-lint |
| `make test` | Run tests |
| `make gen-mocks` | Generate interface mocks |

## API endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/healthz` | Liveness probe (always 200) |
| `GET` | `/readyz` | Readiness probe (checks Postgres + R2) |
| `POST` | `/api/track` | Forward tracking events to Next.js app (Origin/Referer validated) |
| `GET` | `/` | Serve tenant root page |
| `GET` | `/*` | Serve tenant subpage |

## Project structure

```
├── cmd/                    CLI commands (server, migrate)
├── config/                 Environment configuration
├── internal/
│   ├── cache/              In-memory caches (tenants, domains, redirects, sync)
│   ├── handler/            HTTP handlers (serve, health, routes)
│   ├── middleware/          Echo middleware (request ID, recovery, access log)
│   ├── model/              Data models
│   ├── r2/                 Cloudflare R2 S3 client
│   └── repository/         Postgres queries
├── k8s/kustomize/          Kubernetes manifests (base + prd overlay)
├── migrate/                Database migration runner
├── server/                 Server bootstrap and lifecycle
├── Dockerfile              Multi-stage build with goreleaser
├── docker-compose.yaml     Local Docker setup
└── Makefile                Dev workflow targets
```

## Deployment

The service is deployed to Scaleway Kapsule via Kubernetes. The `k8s/kustomize/` directory contains the base manifests and a production overlay.

The init container runs `miqio-proxy migrate` before the main container starts.

```bash
# Build Docker image
docker compose build

# Run with Docker
docker compose up
```

## Tech stack

- **Go** with [Echo v5](https://echo.labstack.com/) HTTP framework
- **Postgres** (pgx) for tenant, domain, and redirect metadata
- **Cloudflare R2** (AWS S3 SDK) for HTML storage
- **Cobra** for CLI
- **Zap** for structured logging
- **Kustomize** for Kubernetes deployment
- **GoReleaser** for builds
