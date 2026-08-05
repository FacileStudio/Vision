# Vision — Development

Local setup, the three processes, the tests, and the quality gate that guards pushes.

## Prerequisites

| Tool | Version | Why |
|---|---|---|
| Go | 1.24 | `apps/api/go.mod` declares `go 1.24.0`; `mise.toml` pins the toolchain |
| Bun | 1.x | Client package manager, test runner, and production runtime |
| PostgreSQL | 16 | What Compose and production run |
| mise | any | Task runner for `install`, `check`, `format`, `hooks` |
| Docker | any | Only if you want the Compose database instead of a local one |

## Setup

```sh
mise run install       # bun install --frozen-lockfile in apps/client
mise run hooks         # git config core.hooksPath .githooks
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
API_URL=http://localhost:4000 bun run dev     # Vite on :5173
```

The Vite dev server runs the same `hooks.server.ts` proxy as production, so the browser only
ever talks to `:5173` and `/api/*` reaches the API with its prefix stripped. Keep `DOMAIN` at
its `http://localhost:5173` default so anything that does hit the API cross-origin is
allowed.

Migrations run on every API start through `schemas.Migrate` — GORM `AutoMigrate` plus the
`seedPersonalWorkspaces` backfill. There is no migration file to write and no down
migration.

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

Two test files exist: `apps/api/modules/workspaces/controller_test.go` and
`apps/client/src/hooks.server.test.ts`, the latter covering the proxy's header handling. The
client uses Bun's built-in test runner, no extra dependency; `*.test.ts` files under `src/`
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

`.githooks/pre-push` runs the **full** `scripts/check.sh`, both halves. Vision is not one of
the repos that had to fall back to `--go-only`, so a push that fails the client type-check is
telling you something real. Bypass once with `git push --no-verify`.

## Conventions

- API modules follow `types.go`, `service.go`, `router.go`, `controller.go` where it applies,
  and `documentation.go`. Modules are registered at the router root in `main.go` — not under
  `/api`, which is added by the proxy.
- Handlers return through `httpjson.WriteJSON` and `httpjson.WriteError` from tronc.
- Any handler that takes a `siteId` must go through `siteaccess.CanAccess` or `CanWrite`
  before touching data; ownership is workspace-scoped, not `owner_id`-scoped.
- The client is Svelte 5 runes only, enforced through `dynamicCompileOptions` in
  `svelte.config.js`. UI primitives under `src/lib/components/ui/` are shadcn-svelte managed.
- `apps/api/tracker/tracker.js` is stale and unused. Edit `apps/client/static/s.js`.
