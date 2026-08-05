# Vision — Architecture

Two servers instead of the suite's usual one, why the tracker drives that split, how a
pageview becomes a row, and how the dashboard reads it back.

## Runtime topology

```
                          ┌─ browser on the dashboard
Internet ──▶ Traefik ──▶ SvelteKit server (bun, :3000) ──┬──▶ /api/e/*  permissive CORS
    │                     hooks.server.ts strips /api    └──▶ /api/*    pass-through
    │                                    │
    └─ browser on a TRACKED third-party site               Go binary (:4000)
       loads /s.js, pixels back to /api/e/p                        │
                                                            Postgres 16
                                                                  │
                                              Journal (log shipping, optional)
                                              Authentik (OIDC, optional)
```

Only the SvelteKit container is published. The Go API has no route through Traefik at all —
`https://vision.facile.studio/health` returns 404, while `/api/health` works, because the
former never reaches the API and the latter is proxied.

`hooks.server.ts` slices four characters off the path (`/api/sites` becomes `/sites`) and
forwards to `API_URL`. The Go router mounts its modules at the root, not under `/api`. Two
paths get special treatment:

- `/api/e/*` — the tracker collection endpoints. The proxy answers `OPTIONS` itself with
  `Access-Control-Allow-Origin: *`, forwards only `Content-Type`, `User-Agent` and
  `CF-IPCountry`, and returns `Cache-Control: no-cache, no-store`.
- everything else under `/api/` — a streaming pass-through that strips hop-by-hop request
  headers and, on the way back, `content-encoding`, `content-length` and
  `transfer-encoding`. `fetch` transparently decompresses the body, so forwarding those
  verbatim would describe a body the client never receives. `set-cookie` is re-appended
  through `getSetCookie()` so multiple cookies survive.

## Components

| Package | Responsibility |
|---|---|
| `main.go` | Config, logger, DB, migrations, service wiring, router, graceful shutdown |
| `internal/env` | Vision's own config loader — see the note below |
| `internal/middleware` | `RequireAuth`, which accepts both session tokens and API keys |
| `internal/siteaccess` | `CanAccess` and `CanWrite`, the workspace-aware permission check |
| `internal/authcrypto` | argon2id passwords, random session tokens |
| `internal/oidcavatar` | Downloads and stores the OIDC `picture` claim |
| `modules/events` | Ingestion, the SSE hub, and the in-memory active-visitor tracker |
| `modules/analytics` | Dashboard aggregation queries, export, public share views |
| `modules/*` | `auth`, `sites`, `goals`, `apikeys`, `workspaces`, `webhooks` |
| `schemas/` | GORM models, `Migrate`, and the personal-workspace backfill |

**Vision does not use `tronc/env`.** It consumes `tronc`'s `health`, `healthcheck`,
`httpx`, `httpjson`, `logger`, `errors` and `middleware` packages, but keeps its own
`internal/env` with its own defaults — including a `DATABASE_URL` that defaults to
`postgres://postgres:postgres@db:5432/vision?sslmode=disable` rather than being required.
That is the opposite of the suite convention, where a missing `DATABASE_URL` exits 1.

## Ingestion

The tracker served to browsers is `apps/client/static/s.js`. It derives its endpoint from
its own `document.currentScript.src`, replacing `/s.js` with `/api/e`, so the same file works
from any hostname without configuration.

Every call is a `new Image()` GET with the payload JSON-encoded into a `data` query
parameter, answered with a 43-byte transparent GIF. No `fetch`, no `sendBeacon`, no CORS
preflight, nothing an ad blocker recognizes as an XHR.

| Endpoint | Sent when | Recorded |
|---|---|---|
| `GET /e/p` | Page load, `history.pushState`, `popstate` | Pageview |
| `GET /e/p?type=perf` | `load` plus 100 ms | Navigation-Timing numbers onto the last pageview |
| `GET /e/t` | `window.vision.track(name, props)` | Custom event |
| `GET /e/h` | Every 30 s while the tab is visible | Touches the active-visitor tracker |
| `POST /e/p` | Not used by `s.js`; available for server-side callers | Pageview |

Site resolution is by hostname. The payload carries `hostname`; when it is missing the
server falls back to parsing the `Referer` header. An unknown hostname is rejected — `403`
on the POST path, silently on the pixel paths, which still return the GIF so a
misconfiguration never shows up as a broken image on someone's site.

The visitor ID is a random string the tracker generates and keeps in `localStorage` under
`_vs_id`. There are no cookies and the `pageviews` table has no IP column: country comes
from `CF-IPCountry` when Cloudflare supplies it, and is otherwise blank. Browser, OS and
device are parsed from the user agent in `modules/events/useragent.go`.

`apps/api/tracker/tracker.js` is **stale**. It posts to `/api/event/pageview` and expects to
be loaded as `/t.js` — neither exists. The file served to browsers is
`apps/client/static/s.js`; treat that as the only tracker.

## Live visitors and streaming

`ActiveTracker` is an in-memory map of site to visitor to last-seen timestamp, with a
60-second window and a sweep every 30 seconds. `GET /analytics/{siteId}/realtime` reads a
count off it. Because it lives in the process, restarting the API zeroes the live count
until heartbeats refill it, and running two API replicas would report two different numbers.

`Hub` is a per-site fan-out of Go channels. `GET /events/{siteId}/live` subscribes and
streams pageviews as server-sent events. It accepts the token in the `Authorization` header
or, since `EventSource` cannot set headers, in a `?token=` query parameter.

## Data model

Eleven tables, `int64` primary keys, migrated by GORM `AutoMigrate` at every boot.

| Table | Key columns | Notes |
|---|---|---|
| `users` | `email` unique, `password_hash`, avatar and OIDC token columns | |
| `sessions` | `token` primary key, `user_id`, `expires_at` | |
| `workspaces` | `name` | Every user gets one, seeded at migration |
| `workspace_members` | `workspace_id`, `user_id`, `role` | Unique on the pair |
| `sites` | `domain` unique, `owner_id`, `workspace_id`, `share_token` unique | |
| `pageviews` | `site_id`, `path`, `referrer`, `browser`, `os`, `device`, `language`, `country`, `visitor_id`, `screen_width`, five `utm_*`, five `perf_*` | Composite index on `(site_id, created_at)` |
| `visitor_sessions` | `site_id`, `visitor_id`, `started_at`, `ended_at`, `entry_path`, `exit_path`, `pageview_count`, `duration` | Derived from consecutive pageviews |
| `custom_events` | `site_id`, `visitor_id`, `path`, `name`, `props` | Composite index on `(site_id, created_at)` |
| `goals` | `site_id`, `owner_id`, `goal_type`, `event_name`, `page_path`, `match_type` | |
| `api_keys` | `key_hash` unique, `prefix`, `key_hint`, `scopes`, `is_active`, `expires_at` | Optionally scoped to a site or workspace |
| `webhooks` | `owner_id`, `workspace_id`, `url`, `secret`, `interval_hours`, `enabled`, `last_sent_at` | |

`Migrate` also runs `seedPersonalWorkspaces`, a backfill that predates workspaces: it
assigns orphan sites to their owner's workspace and creates a personal workspace, owned by
that user, for anyone who has none. It runs on every boot and is a no-op once everyone has
one.

Access control goes through `siteaccess.CanAccess` and `CanWrite`, which resolve a user's
role in the site's workspace. There are no `OnDelete` constraints — orphan rows are
prevented by application code only.

## Authentication

Three credentials reach the same middleware.

1. **Session token.** Email and password, argon2id hashed; login mints a random token stored
   in `sessions` and sent as `Authorization: Bearer`.
2. **API key.** `Authorization: Bearer vis_…` or an `X-API-Key` header. Keys are prefixed
   `vis_ro` (read) or `vis_rw` (read and write), stored only as a SHA-256 hash plus a
   four-character hint, and returned in full exactly once at creation.
3. **OIDC.** Enabled by `OIDC_ISSUER`, discovered at boot through `go-oidc/v3`.

The OIDC flow is the suite's standard one: `GET /auth/oidc` redirects to the provider,
`GET /auth/oidc/callback` verifies the ID token, upserts the user, and redirects to
`OIDC_SUCCESS_URL` with the session token in the **URL fragment** (`#token=…`), which the SPA
reads and discards. `POST /auth/sync-profile` re-reads userinfo for the current user.

`GET /auth/config` reports `sso_only` and `oidc_enabled` so the login page knows what to
render. With `SSO_ONLY=true`, `/auth/register` and `/auth/login` are never registered.

## Cross-app integration

- **Journal.** With both `JOURNAL_URL` and `JOURNAL_TOKEN` set, the tronc logger is wrapped
  in `journal.NewHandler` and every log line ships to Journal.
- **Webhooks.** Vision does not speak the Nook `pool`/`enveloppe` protocol. A scheduler
  ticks every minute, finds webhooks whose `interval_hours` has elapsed since
  `last_sent_at`, and POSTs a per-site traffic report signed with
  `x-vision-signature-256: sha256=<hmac>`.
- **Porte.** OIDC federates to Authentik like the rest of the suite.
