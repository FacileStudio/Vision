# Vision

Self-hosted, cookie-free web analytics. Drop a 3.7 KB unminified script on a site, watch
pageviews, visitors, referrers and goals in a SvelteKit dashboard you own.

Visitors are identified by a random ID in `localStorage`, not a cookie, and no IP address is
ever stored — country comes from Cloudflare's `CF-IPCountry` header when it is present.

Live at [vision.facile.studio](https://vision.facile.studio).

## What it does

- Collects pageviews through a tracking pixel, so ad blockers and `sendBeacon` restrictions
  do not silently kill ingestion
- Records referrers, UTM parameters, browser, OS, device, language, screen width and
  Navigation-Timing performance numbers
- Counts live visitors from a 30-second heartbeat with a 60-second expiry window
- Streams new pageviews to the dashboard over server-sent events
- Tracks custom events and conversion goals, matched on event name or page path
- Groups sites into workspaces with member roles, and shares any dashboard read-only via a
  public token link
- Exports a date range as JSON or CSV, and issues scoped `vis_ro` / `vis_rw` API keys
- Posts periodic HMAC-signed traffic reports to a webhook

## Stack

| Layer | Tech |
|---|---|
| API | Go 1.24, Chi v5, GORM, PostgreSQL 16, [tronc](https://github.com/FacileStudio/tronc) v0.3.0 |
| Client | SvelteKit 2 (Svelte 5 runes), Tailwind CSS 4, shadcn-svelte, LayerChart, TopoJSON world map |
| Auth | Bearer session tokens, argon2id passwords, optional OIDC, `vis_*` API keys |
| Deploy | Docker Compose, three containers: database, API, SvelteKit server |

## Quick start

```sh
cp .env.example .env
docker compose up -d --build
```

Then open <http://localhost:3000>. The SvelteKit container is the only one you talk to; it
proxies `/api/*` to the Go API on its own network.

### Local development

```sh
mise run install
cd apps/api    && go run .        # :4000
cd apps/client && bun run dev     # :5173, proxies to API_URL
```

### Tracking a site

Add the site in the dashboard, then put this on the pages you want measured:

```html
<script defer src="https://vision.facile.studio/s.js"></script>
```

The script derives its collection endpoint from its own `src`, so self-hosting under another
hostname needs no configuration. `window.vision.track(name, props)` sends a custom event.

## Configuration

| Variable | What it does |
|---|---|
| `DATABASE_URL` | Postgres connection string |
| `PORT` | API listen port, `4000` by default |
| `DOMAIN` | Dashboard origin. It is the **only** allowed CORS origin on the API |
| `STORAGE_DIR` | Where uploaded avatars live |
| `API_URL` | Client container only: where the SvelteKit proxy forwards `/api/*` |
| `OIDC_ISSUER` | Set it to turn on SSO; three more `OIDC_*` variables become required |
| `SSO_ONLY` | Disables local email/password auth |

Full reference: [docs/configuration.md](docs/configuration.md).

## Structure

```
apps/
  api/       Go backend — modules/ (auth, sites, events, analytics, goals,
             apikeys, workspaces, webhooks), internal/ (env, middleware,
             siteaccess, database), schemas/ (GORM), tracker/ (stale, see docs)
  client/    SvelteKit server — hooks.server.ts is the API proxy,
             static/s.js is the tracker actually served to browsers
docs/        Architecture, configuration, development, deployment, API
scripts/     check.sh, the repository quality gate
```

## Documentation

| Doc | What's in it |
|---|---|
| [Architecture](docs/architecture.md) | Request flow, ingestion, data model, auth |
| [Configuration](docs/configuration.md) | Every environment variable and default |
| [Development](docs/development.md) | Local setup, tests, the quality gate |
| [Deployment](docs/deployment.md) | Compose topology and why Vision is not one container |
| [API](docs/api.md) | HTTP endpoints and payloads |

---

Part of the [Facile Suite](https://facile.studio) — self-hosted tools for creative studios
and freelancers. One login, zero cloud dependency.
