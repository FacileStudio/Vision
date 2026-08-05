# Vision — API

Every HTTP route the Go binary registers, grouped by module, generated from the routers in
`apps/api/modules/`.

**Paths are given as the Go router registers them, at the root.** The Go API is not exposed
publicly; browsers reach it through the SvelteKit proxy, which strips a leading `/api`. So
`/sites` here is `https://vision.facile.studio/api/sites` from a browser, and
`http://vision-api:4000/sites` on the Compose network.

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

| Method | Path | Notes |
|---|---|---|
| GET | `/e/p` | Pageview. `data` is a `PageviewRequest`. `?type=perf` stores Navigation-Timing numbers instead |
| POST | `/e/p` | Same payload as a JSON body. Resolves the site from `Origin`/`Referer` and returns `204`, or an error |
| OPTIONS | `/e/p` | Preflight, `204`, `Access-Control-Max-Age: 86400` |
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
| POST | `/auth/register` | no | `{ email, password }` → `{ user_id, token }`. Absent when `SSO_ONLY` |
| POST | `/auth/login` | no | `{ email, password }` → `{ user_id, token }`. Absent when `SSO_ONLY` |
| GET | `/auth/me` | yes | `{ id, email, name, avatar_url, avatar_source, created_at }` |
| PUT | `/auth/me` | yes | `{ name, email }` |
| PUT | `/auth/password` | yes | `{ current_password, new_password }` |
| GET | `/auth/oidc` | no | Redirects to the provider |
| GET | `/auth/oidc/callback` | no | Redirects to `OIDC_SUCCESS_URL#token=…` |
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
