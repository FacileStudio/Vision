# Vision

Self-hosted web analytics with no tracking cookies and no ad trackers. Drop a 3.7 KB unminified script on a site, watch
pageviews, visitors, referrers and goals in a SvelteKit dashboard you own.

Visitors are identified by a random ID in `localStorage`, not a tracking cookie, and no IP address is
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
| API | Go 1.26, Chi v5, GORM, PostgreSQL 16, [tronc](https://github.com/FacileStudio/tronc) v0.21.0, [porte](https://github.com/FacileStudio/porte) v0.18.3 |
| Client | SvelteKit 2 (Svelte 5 runes), Tailwind CSS 4, shadcn-svelte, LayerChart, TopoJSON world map |
| Auth | Sign-in through Registre (OIDC) and porte sessions, `vis_*` API keys |
| Deploy | Docker Compose, two containers: database, and one Go binary serving the API and the built client |

## Quick start

```sh
cp .env.example .env
docker compose up -d --build
```

Then reach the `vision` service on port 4000. One Go binary answers `/api/*` and serves the
built SvelteKit client for every other path.

### Local development

```sh
mise run install
cd apps/api    && go run .        # :4000
cd apps/client && bun run dev     # :5173, Vite proxies /api and /docs to :4000
```

### Tracking a site

Add the site in the dashboard, then put this on the pages you want measured:

```html
<script defer src="https://vision.facile.studio/s.js"></script>
```

The script derives its collection endpoint from its own `src`, so self-hosting under another
hostname needs no configuration. `window.vision.track(name, props)` sends a custom event.
It sends nothing from `localhost`, `127.0.0.1`, `*.localhost` or a non-http(s) page. Add
`data-mask` to the tag to replace identifier path segments (digits, UUIDs, anything with an
`@`, 16+ characters with a digit) with `:id` and to strip referrers to origin plus masked path.
`data-exclude="/room/,/share/"` sends nothing (pageview, performance, events, heartbeats) while
the raw path starts with one of those prefixes, on first load and after every navigation.

## Configuration

| Variable | What it does |
|---|---|
| `DATABASE_URL` | Postgres connection string |
| `PORT` | API listen port, `4000` by default |
| `DOMAIN` | Dashboard origin. It is the **only** allowed CORS origin on the API |
| `STORAGE_DIR` | Where uploaded avatars live |
| `OIDC_ISSUER` | Set it to turn on SSO; three more `OIDC_*` variables become required |

Full reference: [docs/configuration.md](docs/configuration.md).

## Structure

```
apps/
  api/       Go backend — modules/ (auth, sites, events, analytics, goals,
             apikeys, workspaces, webhooks), internal/ (env, middleware,
             siteaccess, database), schemas/ (GORM), tracker/ (stale, see docs)
  client/    SvelteKit SPA (adapter-static), served by the Go binary;
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

## License

Vision is licensed under the GNU Affero General Public License v3.0 only (`AGPL-3.0-only`); see [LICENSE](LICENSE). A commercial license is available from Facile Studio for use that cannot meet the AGPL's terms.
