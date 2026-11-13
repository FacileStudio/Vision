# Vision

Self-hosted web analytics platform. Go API + SvelteKit dashboard + PostgreSQL.

## Tech Stack

- **API**: Go 1.26, Chi router, GORM (Postgres driver), OIDC auth support
- **Client**: SvelteKit (Svelte 5 runes), TypeScript, Tailwind CSS v4, shadcn-svelte (nova style), Vite
- **Database**: PostgreSQL 16, schema in goose SQL migrations applied at boot by `tronc/migrate`
- **Runtime**: Bun (client build only), Go (API and the built client)
- **Infra**: Docker Compose (2 services: `vision-db`, and `vision`, one Go binary serving `/api/*` and the SPA), one root `Dockerfile`

## Key Commands

### Client (`apps/client/`)

```sh
bun install                  # install dependencies
bun run dev                  # dev server (Vite, default :5173)
bun run build                # static SPA build (adapter-static, index.html fallback)
bun run check                # svelte-check + TypeScript
```

### API (`apps/api/`)

```sh
go run .                     # start API server (default :4000)
go build -o bin/api .        # compile binary
```

### Docker (project root)

```sh
docker compose up             # full stack (db + the one app container)
docker compose up -d          # detached
```

## Environment

All env vars are optional with working defaults for docker-compose. See `.env.example` and `apps/api/.env.example`.

Key variables:
- `DATABASE_URL` -- Postgres connection string (default: `postgres://postgres:postgres@db:5432/vision?sslmode=disable`)
- `PORT` -- API port (default: `4000`)
- `DOMAIN` -- Frontend origin for CORS (default: `http://localhost:5173`)
- `CLIENT_DIR` -- where the built SPA lives (`/client` in the image, `./client` by default)
- `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URL` -- optional OIDC/SSO config

## Project Structure

```
Vision/
  docker-compose.yml          # full-stack orchestration
  .env.example                # root env template
  ROADMAP.md                  # feature checklist
  apps/
    api/                      # Go backend
      main.go                 # entrypoint, router setup, graceful shutdown
      internal/               # shared infra (db, env, middleware, logging, auth helpers)
      modules/                # feature domains
        analytics/            # dashboard data queries
        auth/                 # local + OIDC authentication
        events/               # pageview/event ingestion, SSE hub, active tracker
        sites/                # site CRUD
        webhooks/             # periodic webhook reports + scheduler
      migrations/             # ordered goose SQL, embedded and applied at boot
      schemas/                # GORM models (query layer) + boot data steps
      tracker/                # tracker.js (client-side snippet)
    client/                   # SvelteKit frontend
      src/
        lib/
          backend.ts          # server-side API client
          components/         # UI components (map, shadcn-svelte primitives)
          stores/             # Svelte stores (user state)
        routes/
          (app)/              # authenticated layout group (dashboard, sites, profile, settings)
          login/              # login page
          share/[token]/      # public shared dashboard
      static/                 # favicon, fonts (Goga Test), tracker script (s.js)
```

## Architecture Notes

- One container, the suite shape: `main.go` mounts every module on the `/api` router, `/docs`, `/health` and `/ready` at the root, and `tronc/spa` last as the catch-all. The client is a static SPA (`ssr = false`) and calls `/api`, relative. Vite proxies `/api` and `/docs` to `:4000` in dev. `events.Preflight` wraps the router and answers `OPTIONS /api/e/*` for any origin before tronc's CORS middleware, which would 403 a third-party preflight. The `/e` group sets `Access-Control-Allow-Origin: *`, `Cache-Control: no-cache, no-store` and a JSON default content type, and rate limits per client address (300 burst, 60/s). Routing lives in the Dokploy domain (service `vision`, port 4000), not in compose labels. `GET /profile` is a Go 308 to `/settings`.
- Database migrations run automatically on API startup via `schemas.MigrateWithIssuer`: the goose SQL in `apps/api/migrations/` through `tronc/migrate`, then the idempotent data steps in `schemas.Reconcile`. A schema change is a new numbered migration plus a hand edit of the struct; `scripts/check.sh` fails on `.AutoMigrate(` outside tests.
- The API has a self-documenting `/docs` endpoint built from per-module `Documentation` structs.
- Tracker script lives in two places: `apps/api/tracker/tracker.js` (source) and `apps/client/static/s.js` (served to browsers). The path is intentionally short/renamed to resist ad blockers.
- Real-time active visitors use in-memory heartbeat tracking (30s pings, 60s expiry window).
- With OIDC on, Registre owns shared workspaces (`modules/workspaces/directory.go`, `apply.go`, porte/directory). At boot Vision imports every unlinked workspace with more than one member once, retrying with backoff while Registre answers 404, then applies Registre's snapshot at boot, every 5 minutes and after a nudge on `POST /api/registre/sync`. `workspaces.registre_space_id` links a workspace to its space. Registre `member` is Vision `editor`. `viewer` is Vision-only: never imported, never removed or changed by a sync unless Registre lists that person. Single-member workspaces are personal: never imported, never linked by name. Membership writes answer 409 for every workspace, and renaming or deleting answers 409 for linked ones. A user in no workspace gets a personal one on their first list.

## Conventions

- API modules follow a consistent pattern: `types.go`, `service.go`, `router.go`, `controller.go` (where applicable), `documentation.go`.
- Client uses shadcn-svelte with the `nova` style and `neutral` base color. Component aliases: `$lib/components/ui`.
- Svelte 5 runes mode is enforced via `svelte.config.js` (`runes: true` for non-node_modules files).
- The client uses Bun's built-in test runner (`bun test`, no extra dependency); `*.test.ts` files under `src/` are excluded from `svelte-check` via `tsconfig.json`. The API has no test runner configured.
