# Vision

Self-hosted web analytics platform. Go API + SvelteKit dashboard + PostgreSQL.

## Tech Stack

- **API**: Go 1.24, Chi router, GORM (Postgres driver), OIDC auth support
- **Client**: SvelteKit (Svelte 5 runes), TypeScript, Tailwind CSS v4, shadcn-svelte (nova style), Vite
- **Database**: PostgreSQL 16 (auto-migrated via GORM)
- **Runtime**: Bun (client build + runtime), Go (API)
- **Infra**: Docker Compose (3 services: db, api, client)

## Key Commands

### Client (`apps/client/`)

```sh
bun install                  # install dependencies
bun run dev                  # dev server (Vite, default :5173)
bun run build                # production build (adapter-node)
bun run check                # svelte-check + TypeScript
```

### API (`apps/api/`)

```sh
go run .                     # start API server (default :4000)
go build -o bin/api .        # compile binary
```

### Docker (project root)

```sh
docker compose up             # full stack (db + api + client)
docker compose up -d          # detached
```

## Environment

All env vars are optional with working defaults for docker-compose. See `.env.example` and `apps/api/.env.example`.

Key variables:
- `DATABASE_URL` -- Postgres connection string (default: `postgres://postgres:postgres@db:5432/vision?sslmode=disable`)
- `PORT` -- API port (default: `4000`)
- `DOMAIN` -- Frontend origin for CORS (default: `http://localhost:5173`)
- `API_URL` -- Client-side env, where the SvelteKit server proxies API requests (default: `http://localhost:4000`)
- `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URL` -- optional OIDC/SSO config
- `SSO_ONLY` -- disable local auth when set to `true`

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
      schemas/                # GORM models + migration
      tracker/                # tracker.js (client-side snippet)
    client/                   # SvelteKit frontend
      src/
        hooks.server.ts       # API proxy (/api/* -> Go backend)
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

- The SvelteKit server acts as an API proxy: all `/api/*` requests from the browser are forwarded to the Go backend via `hooks.server.ts`. Event/tracking requests (`/api/e/*`) get special CORS handling.
- Database migrations run automatically on API startup via `schemas.Migrate()` (GORM AutoMigrate).
- The API has a self-documenting `/docs` endpoint built from per-module `Documentation` structs.
- Tracker script lives in two places: `apps/api/tracker/tracker.js` (source) and `apps/client/static/s.js` (served to browsers). The path is intentionally short/renamed to resist ad blockers.
- Real-time active visitors use in-memory heartbeat tracking (30s pings, 60s expiry window).

## Conventions

- API modules follow a consistent pattern: `types.go`, `service.go`, `router.go`, `controller.go` (where applicable), `documentation.go`.
- Client uses shadcn-svelte with the `nova` style and `neutral` base color. Component aliases: `$lib/components/ui`.
- Svelte 5 runes mode is enforced via `svelte.config.js` (`runes: true` for non-node_modules files).
- The client uses Bun's built-in test runner (`bun test`, no extra dependency); `*.test.ts` files under `src/` are excluded from `svelte-check` via `tsconfig.json`. The API has no test runner configured.
- The proxy strips `content-encoding`, `content-length`, and `transfer-encoding` from upstream responses: `fetch` transparently decompresses the body, so forwarding those headers verbatim would describe a body the client never receives.
