# Vision — Development

Local setup, the three processes, the tests, and the quality gate that guards pushes.

## Prerequisites

| Tool | Version | Why |
|---|---|---|
| Go | 1.26 | `apps/api/go.mod` declares `go 1.26`; `mise.toml` pins the toolchain |
| Bun | 1.x | Client package manager, test runner, and production runtime |
| PostgreSQL | 16 | What Compose and production run |
| mise | any | Task runner for `install`, `check`, `format`, `hooks` |
| Docker | any | Only if you want the Compose database instead of a local one |

## Setup

```sh
mise run install       # bun install --frozen-lockfile in apps/client
mise run hooks         # lefthook install --reset-hooks-path
```

Bring up a database:

```sh
docker compose up -d vision-db
```

## Running

```sh
cd apps/api
DATABASE_URL=postgres://postgres:postgres@localhost:5432/vision?sslmode=disable go run .
```

```sh
cd apps/client
bun run dev     # Vite on :5173
```

Vite proxies `/api` and `/docs` to `http://127.0.0.1:4000`, so the browser only ever talks to
`:5173` and every path reaches the API unchanged, as in production. Keep `DOMAIN` at
its `http://localhost:5173` default so anything that does hit the API cross-origin is
allowed.

Migrations run on every API start through `schemas.MigrateWithIssuer`: the goose SQL under
`apps/api/migrations/`, then the boot data steps in `schemas.Reconcile`. To change the schema,
add the next numbered file under `apps/api/migrations/` with both `-- +goose Up` and
`-- +goose Down`, update the struct in `schemas/` by hand, and regenerate `apps/api/schema.sql`:

```sh
pg_dump --schema-only --no-owner --no-privileges --no-comments --no-tablespaces --schema=public \
  --exclude-table=goose_db_version --exclude-table=goose_db_version_id_seq "$DATABASE_URL" \
  | grep -vE '^(SET |SELECT pg_catalog\.set_config|\\restrict|\\unrestrict|--|CREATE SCHEMA)' | cat -s
```

`go run . migrate status` lists what is applied. `scripts/check.sh` fails on any
`.AutoMigrate(` outside a test file.

## Testing the tracker locally

The tracker resolves its endpoint from its own `src`, so serving a test page from the dev
server is enough:

```html
<script defer src="http://localhost:5173/s.js"></script>
```

Create a site in the dashboard whose domain matches the hostname you load that page from,
or ingestion rejects it. When debugging, remember the pixel endpoints answer with a GIF
whether or not the hostname resolved — a working image is not proof of a recorded pageview.
Check the `pageviews` table.

## Tests

```sh
cd apps/api    && go test ./...
cd apps/client && bun test
```

The client uses Bun's built-in test runner, no extra dependency; `*.test.ts` files under `src/`
are excluded from `svelte-check` through `tsconfig.json`.

## The quality gate

`scripts/check.sh` is the gate. It depends on nothing but `go` and, for the client half,
`bun`. It is a shell script rather than a mise task body on purpose: `mise run` resolves
every tool in the merged config before running anything, so one broken tool in your global
config would otherwise take the gate down with it.

```sh
sh scripts/check.sh              # gofmt -l, go vet, go test, then svelte-check
sh scripts/check.sh --go-only    # Go half only
sh scripts/check.sh --format     # rewrite Go sources in place
```

Equivalent mise tasks: `mise run check`, `mise run check-go`, `mise run format`.

The script resolves `go` and `gofmt` from `GOROOT` when it is set. mise exports `GOROOT` for
the pinned version but leaves an unrelated `go` earlier on `PATH`, and mixing the two
produces `compile: version "X" does not match go tool version "Y"`.

If `bun` is missing, the client half is skipped with a warning rather than failing.

## The pre-push hook

`lefthook.yml` is tracked. Its pre-push runs `filet gate .`, which covers both units
(`apps/api` and `apps/client`): gofmt, `filet check`, `go vet`, staticcheck, `bun run check`,
the Go and bun tests and douane, then `scripts/check-version.sh`. The commit-msg hook comes from
the shared `FacileStudio/hooks` config. A push therefore needs `go`, `bun`, `filet`, `douane`
and `TEST_DATABASE_URL` in the shell. Bypass once with `git push --no-verify`.

`filet.yml` is `preset: facile` with no exception. `.filet-baseline.json` is the ratchet
baseline; it is empty because every standing finding is a warning.

## Conventions

- API modules follow `types.go`, `service.go`, `router.go`, `controller.go` where it applies,
  and `documentation.go`. Modules are registered on the `/api` router in `main.go`;
  `TestEveryAPIRouteHangsOffTheAPIPrefix` fails on anything else outside `/docs`.
- Handlers return through `httpjson.WriteJSON` and `httpjson.WriteError` from tronc.
- Any handler that takes a `siteId` must go through `siteaccess.CanAccess` or `CanWrite`
  before touching data; ownership is workspace-scoped, not `owner_id`-scoped.
- The client is Svelte 5 runes only, enforced through `dynamicCompileOptions` in
  `svelte.config.js`. UI primitives under `src/lib/components/ui/` are shadcn-svelte managed.
- `apps/api/tracker/tracker.js` is stale and unused. Edit `apps/client/static/s.js`.
