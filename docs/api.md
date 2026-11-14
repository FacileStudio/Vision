# Vision — API

A Scalar reference is served at `GET /docs`, with the OpenAPI 3.1 document at
`GET /docs/openapi.json`, built by `tronc/apiref` from the per-module `documentation.go`
registries. `/api/docs` also works, because the reference derives its spec URL from the
request path.

Every HTTP route the Go binary registers, grouped by module, generated from the routers in
`apps/api/modules/`.

**Paths are given relative to `/api`, the router every module hangs off.** So `/sites` here
is `https://vision.facile.studio/api/sites`.

Responses are JSON via `httpjson.WriteJSON`, errors share the tronc error shape.
Authenticated routes accept either a session token or an API key:

```
Authorization: Bearer <session-token>
Authorization: Bearer vis_ro_<key>      # or vis_rw_
X-API-Key: vis_rw_<key>
```

`GET /docs` returns a self-describing document assembled from each module's `Documentation`
struct. `/health`, `/ready`, `/api/health` and `/api/ready` are unauthenticated; `/ready`
pings the database.

## Collection (public, third-party origins)

Every route here sets `Access-Control-Allow-Origin: *`. The GET variants take their payload
as a JSON-encoded `data` query parameter and answer with a 43-byte transparent GIF whether
or not anything was recorded — a valid image is not proof of ingestion.

Each client address gets a token bucket of 300 requests refilling at 60 a second across all
of `/e/*`. Past it the answer is `429` with `Retry-After` and the `rate_limited` error body.

| Method | Path | Notes |
|---|---|---|
| GET | `/e/p` | Pageview. `data` is a `PageviewRequest`. `?type=perf` stores Navigation-Timing numbers instead |
| POST | `/e/p` | Same payload as a JSON body. Resolves the site from `Origin`/`Referer` and returns `204`, or an error |
| OPTIONS | `/e/*` | Preflight from any origin, `204`, `Access-Control-Max-Age: 86400` |
| GET | `/e/t` | Custom event. `data` is a `CustomEventRequest` |
| GET | `/e/h` | Heartbeat. `{ hostname, visitor_id }`; touches the live-visitor tracker |

```json
{
  "hostname": "example.com",
  "path": "/pricing",
  "referrer": "https://news.example",
  "language": "fr-FR",
  "visitor_id": "…",
  "screen_width": 1440,
  "utm_source": "", "utm_medium": "", "utm_campaign": "",
  "utm_term": "", "utm_content": "",
  "performance": { "dns": 3, "tcp": 11, "ttfb": 82,
                   "dom_load": 410, "page_load": 730 }
}
```

A custom event is `{ hostname, path, visitor_id, event_name, event_props }`, where
`event_props` is an arbitrary object.

The site is resolved by hostname — from the payload, falling back to parsing `Referer`. An
unknown hostname is rejected. `CF-IPCountry` supplies the country; no IP is stored.

## Live stream

| Method | Path | Auth | Notes |
|---|---|---|---|
| GET | `/events/{siteId}/live` | yes | Server-sent events, one message per pageview. Accepts the token in `Authorization` or, for `EventSource`, as `?token=` |

## Auth

| Method | Path | Auth | Notes |
|---|---|---|---|
| GET | `/auth/config` | no | `{ sso_only, oidc_enabled }` |
| GET | `/auth/me` | yes | `{ id, email, name, avatar_url, avatar_source, locale, created_at }` |
| PUT | `/auth/me` | yes | `{ name, email }` |
| GET | `/version` | no | `{ version }`, the version in `apps/client/package.json`, baked in at build time, or `dev` |
| POST | `/registre/deletions` | signed | Registre's signed `user.deleted` notice (public path `/api/registre/deletions`) |
| GET | `/auth/oidc` | no | Redirects to the provider |
| GET | `/auth/oidc/callback` | no | Sets the session cookie and redirects to `OIDC_SUCCESS_URL` |
| POST | `/auth/sync-profile` | yes | Re-reads userinfo for the current user |

The OIDC routes only exist when `OIDC_ISSUER` is set.

## Files

| Method | Path | Auth | Notes |
|---|---|---|---|
| GET | `/files/*` | no | Static server over `STORAGE_DIR`, `Cache-Control: public, max-age=86400, immutable`. Serves avatars |

## Sites

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/sites` | yes | `{ name, domain, workspace_id }`. `domain` is unique across the install |
| GET | `/sites` | yes | |
| GET | `/sites/{id}` | yes | |
| PUT | `/sites/{id}` | yes | `{ name, domain }` |
| DELETE | `/sites/{id}` | yes | |
| POST | `/sites/{id}/share` | yes | Mints a public `share_token` |
| DELETE | `/sites/{id}/share` | yes | Revokes it |

`SiteResponse` is `{ id, name, domain, owner_id, workspace_id, share_token, created_at,
updated_at }`.

## Analytics

| Method | Path | Auth | Notes |
|---|---|---|---|
| GET | `/analytics/summary` | yes | `?space=all\|<id>&from=&to=` (RFC 3339 or YYYY-MM-DD). Every site in the caller's workspaces: visitors, pageviews, previous-window counts, live, zero-filled hourly (window up to 48 h) or daily series. A space the caller is not in answers `sites: []` |
| GET | `/analytics/{siteId}/overview` | yes | The dashboard payload |
| GET | `/analytics/{siteId}/realtime` | yes | `{ "visitors": n }` from the in-memory tracker |
| GET | `/analytics/{siteId}/export` | yes | `?format=csv` for CSV, otherwise JSON. Both sent as an attachment |
| GET | `/share/{token}` | no | Overview for a shared site, by share token |
| GET | `/share/{token}/realtime` | no | Live count for a shared site |

Query parameters, shared by all of the above: `from`, `to`, `granularity` (defaults to
`day`), and the filters `country`, `browser`, `os`, `device`, `path`, `referrer`.

`OverviewResponse` carries totals (`total_pageviews`, `unique_visitors`), the previous
period's equivalents for comparison, `bounce_rate`, `avg_session_duration`,
`pages_per_session`, time series (`pageviews_per_day`, `unique_visitors_per_day`,
`hourly_distribution`, and `prev_*` variants), leaderboards (`top_pages`, `top_referrers`,
`top_countries`, `top_browsers`, `top_os`, `top_devices`, `top_screens`, `top_entry_pages`,
`top_exit_pages`, `top_utm_sources`, `top_utm_mediums`, `top_utm_campaigns`, `top_events`)
and a `performance` block.

CSV export is three columns: `Date`, `Pageviews`, `Unique Visitors`.

Every route taking a `siteId` runs `siteaccess.CanAccess` first and answers `404`, not `403`,
when it fails.

## Goals

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/goals` | yes | `{ site_id, name, goal_type, event_name, page_path, match_type }` |
| GET | `/goals` | yes | |
| PUT | `/goals/{id}` | yes | |
| DELETE | `/goals/{id}` | yes | |
| GET | `/goals/{siteId}/conversions` | yes | `{ goals: [{ id, name, goal_type, conversions, conversion_rate }], total_visitors }` |

`match_type` defaults to `exact`.

## Workspaces

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/workspaces` | yes | `{ name }` |
| GET | `/workspaces` | yes | Each carries your `role`, `member_count`, `site_count` |
| GET | `/workspaces/{id}` | yes | |
| PUT | `/workspaces/{id}` | yes | `{ name }` |
| DELETE | `/workspaces/{id}` | yes | |
| GET | `/workspaces/{id}/members` | yes | |
| POST | `/workspaces/{id}/members` | yes | `{ email, role }` |
| PUT | `/workspaces/{id}/members/{userId}` | yes | `{ role }` |
| DELETE | `/workspaces/{id}/members/{userId}` | yes | |
| POST | `/workspaces/{id}/leave` | yes | |

## API keys

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/api-keys` | yes | `{ name, scopes, site_id?, workspace_id? }` → `{ key, api_key }` |
| GET | `/api-keys` | yes | Metadata only |
| DELETE | `/api-keys/{id}` | yes | Revokes |

`scopes` is `read` or `read,write`, which decides the `vis_ro` / `vis_rw` prefix. The full
key is returned exactly once, at creation; the database stores a SHA-256 hash plus a
four-character `key_hint`. A key can be scoped to one site or one workspace.

## Webhooks

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/webhooks` | yes | `{ url, secret, interval_hours, workspace_id? }` |
| GET | `/webhooks` | yes | Secrets are not returned |
| GET | `/webhooks/{id}` | yes | |
| PUT | `/webhooks/{id}` | yes | `{ url, secret, interval_hours, enabled }` |
| DELETE | `/webhooks/{id}` | yes | |
| POST | `/webhooks/{id}/test` | yes | Sends a report immediately |

`interval_hours` defaults to 24 and is rendered into a human `period` label — `hourly`,
`daily`, `weekly`, `monthly`, or `every Nh` / `every Nd`.

### Outgoing reports

A scheduler ticks every minute and POSTs one report per site to every due webhook, with
`User-Agent: Vision-Webhook/1.0` and `x-vision-signature-256: sha256=<hmac>`, the HMAC-SHA256
of the raw body keyed on the webhook's secret.

```json
{
  "event_type": "…",
  "site": { "id": 1, "name": "…", "domain": "example.com" },
  "period": { "type": "daily", "from": "…", "to": "…" },
  "metrics": { "total_pageviews": 0, "unique_visitors": 0,
               "views_per_visitor": 0, "prev_total_pageviews": 0,
               "prev_unique_visitors": 0, "pageviews_change_pct": 0,
               "visitors_change_pct": 0 },
  "top_pages": [{ "name": "/", "count": 0 }],
  "top_referrers": [], "top_countries": [],
  "top_browsers": [], "top_devices": []
}
```

## MCP

`POST /mcp`, at the root and outside `/api`, is Vision's remote MCP server (spec 2026-07-28,
stateless Streamable HTTP, built on `tronc/mcp`). It takes Registre access tokens bound to
`https://vision.facile.studio/mcp` with the scopes `vision.read` and `vision.write`;
`/.well-known/oauth-protected-resource/mcp` is the RFC 9728 metadata that tells a client how
to get one. MCP is free on every plan. A person's first call creates their Vision account.

| Tool | Scope | Confirmed |
|---|---|---|
| `vision_list_spaces` | read | |
| `vision_list_sites` | read | |
| `vision_get_stats` | read | |
| `vision_list_goals` | read | |
| `vision_create_site` | write | |
| `vision_update_site` | write | |
| `vision_unshare_site` | write | |
| `vision_create_goal` | write | |
| `vision_update_goal` | write | |
| `vision_delete_site` | write | yes |
| `vision_share_site` | write | yes |
| `vision_delete_goal` | write | yes |

Each tool calls the service of the matching REST route, so roles hold as in the web app.
A confirmed tool asks the person through MCP elicitation when the client supports it.
Otherwise it stores a pending action and returns a `confirm_url` to `/pending`, where the
signed-in person confirms or dismisses it:

| Method | Path | Auth | Notes |
|---|---|---|---|
| GET | `/mcp/pending` | session | Open actions with their site or goal as it is now |
| POST | `/mcp/pending/{id}/confirm` | session | Runs it once; 412 when the target changed since |
| POST | `/mcp/pending/{id}/dismiss` | session | Closes it without running it |

Only the person's own browser session can confirm (porte `mcpauth.RequireOwnSession`): the
session cookie a browser sign-in minted, with no `Authorization` header. A bearer token, a CLI
session or a delegated token gets 403, and so does a cookie issued before this release: the
`/pending` page then asks the person to sign in again. An API key gets 401, as it carries no
porte session. Pending actions resolved or expired more than 30 days ago are pruned every hour.
