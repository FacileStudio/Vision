# Vision — Deployment

What Compose brings up, how the hostname is routed, and what to watch during an upgrade.

## One container

Vision follows the suite's deployment shape: one Go binary behind one Traefik router on one
hostname. It answers `/api/*`, keeps `/docs`, `/health` and `/ready` at the root, and serves
the built SvelteKit client for every other path.

The tracker was the reason Vision once ran a SvelteKit server in front of the API: the
collection endpoint `/api/e/*` must answer preflights from third-party sites, and the API's
CORS policy allows one origin, `DOMAIN`. The binary now answers those preflights itself,
before its CORS middleware (`events.Preflight`), so the proxy is gone.

## Compose topology

```
Traefik ──▶ vision (:4000) ──▶ vision-db (:5432)
                                    │
                               db_data volume
```

| Service | Image | Notes |
|---|---|---|
| `vision-db` | `postgres:16-alpine` | `pg_isready` healthcheck every 5s, 10 retries |
| `vision` | built from the root `Dockerfile` | `expose: 4000`, waits for the database to be healthy |

Neither service publishes a port. One named volume, `vision_db_data`.

## Image

The root `Dockerfile` builds the client with `oven/bun` (`bun install --frozen-lockfile`,
then `vite build` with `adapter-static`), builds the API with `golang:1.26.10-alpine` and
`CGO_ENABLED=0`, stamping the version from `apps/client/package.json` through
`scripts/version.sh`, and copies both into `gcr.io/distroless/static-debian12:nonroot`: the
binary at `/api`, the SPA at `/client` (`CLIENT_DIR`). The container writes nothing, so it
runs as `nonroot`. Anything that adds a writable volume must drop `:nonroot` and the `USER`
line.

## Routing

`docker-compose.yml` carries **no Traefik labels**. Routing is configured in Dokploy at
`gare.facile.studio`, whose `Domain` record for `vision.facile.studio` points at the Compose
service `vision`, port `4000`, path `/`, no path stripping. Two consequences:

- Renaming or removing `vision` fails the deployment with
  `Domain "x" is attached to service "y" which does not exist in the compose`. Repoint the
  domain first.
- The `/api` prefix must reach the binary intact. A strip-prefix middleware 404s every route.

Vision sits behind Cloudflare, which is what supplies `CF-IPCountry`. Losing that header
does not break ingestion — country simply goes blank.

## Healthchecks

`tronc/health` mounts `/health` and `/ready`, plus `/api/health` and `/api/ready`, all public.
`/ready` pings the database.

The image is distroless, with no shell and no `curl`, so the container healthcheck runs
`/api healthcheck`: `tronc`'s `healthcheck.Handle` intercepts `os.Args` before anything else
in `main` for exactly that.

## Migrations

There is no migration step. `schemas.Migrate` runs GORM `AutoMigrate` on every boot, then
`seedPersonalWorkspaces`, which assigns orphan sites to their owner's workspace and creates a
personal workspace for any user without one. Both are idempotent.

- Adding a struct field adds a column. Removing one leaves the column behind, unread.
- Two instances booting at once both migrate. Deploy one at a time.
- There are no `OnDelete` constraints on any relation, so deleting a site does not cascade
  to its pageviews at the database level. Application code is the only thing preventing
  orphan rows.

## What an upgrade risks

Ingestion is the part that fails quietly. After any deploy, check that a tracked site is
still recording — the tracker returns a valid GIF whether the pageview landed or not, so
neither the browser console nor the API access log will tell you it stopped. Compare a
recent `created_at` in `pageviews`, or watch the live counter with a tab open on a tracked
site.

The live-visitor count is in-process memory and resets to zero on every restart. It refills
within 30 seconds from heartbeats. Running more than one API replica would split that count
and split the SSE hub, so Vision scales vertically only.
