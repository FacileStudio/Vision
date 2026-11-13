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

The schema is ordered goose SQL in `apps/api/migrations/`, embedded in the binary and applied
at every boot by `tronc/migrate`, behind a Postgres advisory lock. A failed migration stops the
boot with a non-zero exit. `00001_baseline.sql` is production's schema as GORM `AutoMigrate`
built it. A database that predates v0.6.0 (Vision's tables, no applied migration) is stamped at
version 1 on its first boot, so the baseline is recorded without running.

After the migrations, `schemas.Reconcile` runs the data steps every boot ran before, in the same
order: it empties placeholder pictures, runs `seedPersonalWorkspaces` (orphan sites go to their
owner's workspace, and any user without a workspace gets a personal one), and moves legacy
credentials onto porte's tables. All of them are idempotent.

`docker run <image> migrate status` (also `up`, `version` and `baseline <n>`) answers without a
shell.

- Adding a struct field needs a migration. `TestGORMModelsMatchTheMigratedSchema` fails until
  the column exists.
- Rolling back to v0.5.x runs `AutoMigrate` again, which drops the defaults `00002` added while
  the ledger still records version 2. Before deploying v0.6.0 or later again after such a
  rollback, run `DELETE FROM goose_db_version WHERE version_id = 2`.
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
