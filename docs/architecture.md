# Vision — Architecture

One Go binary serves the whole hostname, how a pageview becomes a row, and how the dashboard
reads it back.

## Runtime topology

```
                          ┌─ browser on the dashboard
Internet ──▶ Traefik ──▶ Go binary (:4000) ──┬──▶ /api/e/*      tracker, permissive CORS
    │                                         ├──▶ /api/*        API routes
    └─ browser on a TRACKED third-party site  ├──▶ /docs, /health, /ready
       loads /s.js, pixels back to /api/e/p   └──▶ everything else: the built SPA
                                                        │
                                                   Postgres 16
                                                        │
                                       Journal (log shipping, optional)
                                       Registre (OIDC, optional)
```

The suite's one-container shape: `main.go` mounts every module on a router at `/api`, keeps
`/docs`, `/health` and `/ready` at the root, and mounts `tronc/spa` last as the catch-all
over `CLIENT_DIR` (`/client` in the image). The client is a static SvelteKit build
(`adapter-static`, `index.html` fallback, `ssr = false`), so there is no client server and
every browser request is same-origin.

The tracker routes are the one cross-origin surface:

- `events.Preflight` wraps the whole router and answers `OPTIONS /api/e/*` with
  `Access-Control-Allow-Origin: *` before the router's CORS middleware, which would refuse a
  preflight from any origin but `DOMAIN` with 403.
- Every `/api/e/*` answer carries `Access-Control-Allow-Origin: *` and
  `Cache-Control: no-cache, no-store`, and is `application/json` unless it is the GIF pixel.
- Each client address gets 300 requests refilling at 60 a second (`TrackerRate`,
  `TrackerBurst`), then `429`. The key is `RemoteAddr` as tronc's real-IP middleware rewrote
  it from `TRUSTED_PROXIES`, never a header read by hand.
- `GET /profile` answers `308` to `/settings` in Go, so the bookmarked path keeps its status.

## Components

| Package | Responsibility |
|---|---|
| `main.go` | Config, logger, DB, migrations, service wiring, router, graceful shutdown |
| `internal/env` | Vision's own config loader — see the note below |
| `internal/middleware` | `RequireAuth`: an API key first, else porte/kit's session middleware |
| `internal/siteaccess` | `CanAccess` and `CanWrite`, the workspace-aware permission check |
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

Eleven tables, `int64` primary keys, built by the goose SQL under `apps/api/migrations/`,
which `tronc/migrate` applies at every boot.

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

`schemas.Reconcile` runs `seedPersonalWorkspaces` after the migrations, a backfill that predates workspaces: it
assigns orphan sites to their owner's workspace and creates a personal workspace, owned by
that user, for anyone who has none. It runs on every boot and is a no-op once everyone has
one.

Access control goes through `siteaccess.CanAccess` and `CanWrite`, which resolve a user's
role in the site's workspace. There are no `OnDelete` constraints — orphan rows are
prevented by application code only.

## Authentication

Sign-in goes through Registre only: Vision has no register, login or password route. Three
credentials reach the same middleware.

1. **Session.** The OIDC callback issues a porte session as an HttpOnly cookie. The browser
   sends nothing else: it keeps no token in `localStorage` and deletes any an old login left.
   `Authorization: Bearer` still works for CLI tokens.
2. **API key.** `Authorization: Bearer vis_…` or an `X-API-Key` header. Keys are prefixed
   `vis_ro` (read) or `vis_rw` (read and write), stored only as a SHA-256 hash plus a
   four-character hint, and returned in full exactly once at creation.
3. **OIDC.** Enabled by `OIDC_ISSUER`, discovered at boot through `go-oidc/v3`.

The OIDC flow is the suite's standard one: `GET /auth/oidc` redirects to the provider,
`GET /auth/oidc/callback` verifies the ID token, upserts the user, sets the session cookie and
redirects to `OIDC_SUCCESS_URL`. The client's fetch kernel is `@facile/client`: it sends the
cookie same-origin, adds `X-Facile-CSRF` on mutations, and sends the person to `/login` on a
401. `POST /auth/sync-profile` re-reads userinfo for the current user.

`GET /auth/config` reports `sso_only` and `oidc_enabled`. porte mounts no password route
while `OIDC_ISSUER` is set.

Accounts live in Vision's `users` table through porte/kit's store (`modules/auth/store.go`).
An existing account links to a Facile sign-in by verified email only, and only when it holds
no other subject of the same issuer. The stored address changes at sign-in only when Registre
marks it verified. The picture is `users.oidc_picture_url`, and the person's Registre locale
is stored in `users.locale`. Vision has no admin flag: roles live in `workspace_members`.

When Registre erases a person, porte/kit's `EraseAccount` runs `workspaces.EraseData` and then
deletes the account, in one transaction. It deletes their personal workspaces and hands the
sites and goals they authored in shared workspaces to the next member. It runs from the
deletion poll or from Registre's signed push, and a second call changes nothing. See
[erasure.md](erasure.md).

## Cross-app integration

- **Journal.** With both `JOURNAL_URL` and `JOURNAL_TOKEN` set, the tronc logger is wrapped
  in `journal.NewHandler` and every log line ships to Journal.
- **Webhooks.** Vision does not speak the Nook `pool`/`enveloppe` protocol. A scheduler
  ticks every minute, finds webhooks whose `interval_hours` has elapsed since
  `last_sent_at`, and POSTs a per-site traffic report signed with
  `x-vision-signature-256: sha256=<hmac>`.
- **Porte.** OIDC federates to Authentik like the rest of the suite.
