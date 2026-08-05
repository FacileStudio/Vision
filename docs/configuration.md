# Vision — Configuration

Every environment variable the two containers actually read, taken from
`apps/api/internal/env/env.go`, `apps/client/src/hooks.server.ts` and the Dockerfiles.

Vision keeps its own config loader rather than using `tronc/env`, so its defaults differ
from the rest of the Go family — most importantly, **nothing is required**. A misconfigured
Vision starts happily and fails later against a database that is not there.

## API

| Variable | Required | Default | What it does |
|---|---|---|---|
| `DATABASE_URL` | no | `postgres://postgres:postgres@db:5432/vision?sslmode=disable` | Postgres connection string. The default names a host `db`, which no longer exists in the Compose file — the service is `vision-db` |
| `PORT` | no | `4000` | HTTP listen port. Must parse as an integer in 1–65535 or startup fails |
| `DOMAIN` | no | `http://localhost:5173` | The dashboard origin, and the sole entry in the API's CORS allow-list. Also the default OIDC success redirect |
| `LOG_LEVEL` | no | `info` | One of `debug`, `info`, `warn`, `error`. Anything else fails startup |
| `STORAGE_DIR` | no | `./data` | Root for stored avatars. `STORAGE_DIR/avatars` is created at startup and served from `GET /files/*` |
| `SSO_ONLY` | no | `false` | Compared against the literal string `true`, case-insensitively. Any other value is false |
| `JOURNAL_URL` | no | — | Journal ingest URL. Shipping needs both this and the token |
| `JOURNAL_TOKEN` | no | — | Per-app Journal key |

`DOMAIN` is the CORS allow-list, not a fallback for one. Setting it wrong does not break the
dashboard, which goes through the SvelteKit proxy and is therefore same-origin — it breaks
anything calling the API cross-origin. The tracker endpoints under `/e/` set
`Access-Control-Allow-Origin: *` on their own and are unaffected.

## OIDC

Off until `OIDC_ISSUER` is set. Once it is, three more variables become required and
`env.Load` returns an error, which stops the process.

| Variable | Required | Default | What it does |
|---|---|---|---|
| `OIDC_ISSUER` | no | — | Discovery URL, e.g. `https://porte.facile.studio/application/o/vision/` |
| `OIDC_CLIENT_ID` | with issuer | — | Client ID |
| `OIDC_CLIENT_SECRET` | with issuer | — | Client secret |
| `OIDC_REDIRECT_URL` | with issuer | — | The provider's callback target. It has to be a URL that reaches the API, which in the deployed topology means going through the SvelteKit proxy — `https://<host>/api/auth/oidc/callback`, not the API's own `/auth/oidc/callback` |
| `OIDC_SUCCESS_URL` | no | value of `DOMAIN` | Where the callback sends the browser, with `#token=…` appended |

## Client container

The SvelteKit server reads exactly two variables.

| Variable | Required | Default | What it does |
|---|---|---|---|
| `API_URL` | no | `http://localhost:4000` in code, `http://vision-api:4000` in the Dockerfile | Where `hooks.server.ts` forwards everything under `/api/` |
| `PORT` | no | `3000`, set in the Dockerfile | Listen port for the adapter-node server |

`API_URL` is read from `process.env` at module load, so it must be present in the server
process's environment — not baked in at build time.

## Compose-only variables

| Variable | Default | What it does |
|---|---|---|
| `POSTGRES_USER` | `postgres` | Database superuser |
| `POSTGRES_PASSWORD` | `postgres` | Its password |
| `POSTGRES_DB` | `vision` | Database name |

`docker-compose.yml` hardcodes the API's `DATABASE_URL` to
`postgres://postgres:postgres@vision-db:5432/vision?sslmode=disable`, so changing these three
without changing that string breaks the connection.

## What is not configured through the environment

- **Sites and tracking.** A site's domain is a database row, created through the dashboard.
  The tracker needs no site ID or API key — it identifies the site by hostname.
- **API keys.** Created through `/api-keys`, returned in full once, stored hashed.
- **Webhooks.** URL, secret and interval are per-webhook database rows.
