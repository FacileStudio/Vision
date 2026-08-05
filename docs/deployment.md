# Vision — Deployment

What Compose brings up, why Vision is the suite's documented exception to the
one-container rule, and what to watch during an upgrade.

## Why three containers

Every other Go app in the suite ships as one binary serving both `/api/*` and the built SPA
behind a single Traefik router. Vision does not, and the reason is the tracker.

`/api/e/*` is the collection endpoint for a script embedded on **third-party sites**. It has
to answer preflights with `Access-Control-Allow-Origin: *` and propagate `CF-IPCountry`.
Putting it behind the app's own Traefik router would subject it to the app's CORS policy,
which allows exactly one origin — `DOMAIN`. Every tracked site would stop reporting, with no
visible error anywhere: the pixel endpoints return a valid GIF regardless, so the failure
would be silent on both ends.

Collapsing Vision into one container therefore requires first porting that collection point
into Go with a permissive CORS policy scoped to that route alone. Until then, three
containers is the deliberate, reviewed choice — not drift. It is recorded as such in the
suite's `ROADMAP.md`, Appendix C.

## Compose topology

```
                    ┌──────────────── published ────────────────┐
Traefik ──▶ vision-client (bun, :3000) ──▶ vision-api (:4000) ──▶ vision-db (:5432)
                                                                        │
                                                                   db_data volume
```

| Service | Image | Notes |
|---|---|---|
| `vision-db` | `postgres:16-alpine` | `pg_isready` healthcheck every 5s, 10 retries |
| `vision-api` | built from `apps/api/Dockerfile` | `expose: 4000`, waits for the database to be healthy |
| `vision-client` | built from `apps/client/Dockerfile` | `expose: 3000`, `API_URL=http://vision-api:4000`, depends on the API |

Neither app service publishes a port. Only `vision-client` is reachable from outside, and
the API is only reachable from the client over the Compose network. That is the security
property that makes the split acceptable: the Go API is not on the internet.

One named volume, `vision_db_data`. There is no volume for `STORAGE_DIR`, so **avatars do
not survive a container rebuild**.

## Images

`apps/api/Dockerfile` builds a static binary with `CGO_ENABLED=0` from `golang:1.24-alpine`
and copies it into `gcr.io/distroless/static-debian12`. `apps/client/Dockerfile` builds with
`oven/bun:1`, then runs `bun build/index.js` on `oven/bun:1-slim` with `node_modules` copied
across — the SvelteKit adapter is `adapter-node`, so it needs a runtime, not just static
files.

## Routing

`docker-compose.yml` carries **no Traefik labels**. Routing is configured in Dokploy at
`gare.facile.studio`, whose `Domain` record points at a Compose service name — here,
`vision-client`. Two consequences:

- Renaming or removing `vision-client` fails the deployment with
  `Domain "x" is attached to service "y" which does not exist in the compose`. Repoint with
  `dokploy domain update` first.
- The Go API's own paths are not publicly reachable. `https://vision.facile.studio/health`
  is a 404; `https://vision.facile.studio/api/health` returns `{"status":"ok"}` because it
  goes through the proxy. Monitor the second one.

Vision sits behind Cloudflare, which is what supplies `CF-IPCountry`. Losing that header
does not break ingestion — country simply goes blank.

## Healthchecks

`tronc/health` mounts `/health` and `/ready` on the Go router, plus `/api/health` and
`/api/ready`. `/ready` pings the database. Through the public hostname only the `/api/*`
pair is reachable, and each is served by the Go API after the proxy strips the prefix.

The API image is distroless, with no shell and no `curl`, so any container-level healthcheck
has to invoke the binary itself — `tronc`'s `healthcheck.Handle` intercepts `os.Args` before
anything else in `main` for exactly that.

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
