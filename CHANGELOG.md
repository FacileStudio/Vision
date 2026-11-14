# Changelog

All notable changes to Vision are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Security

- MCP confirm page: only the person's own browser session confirms a pending action (porte v0.20.1 `mcpauth.RequireOwnSession`). A CLI session, a bearer next to the cookie or a delegated token gets 403; an API key gets 401. A cookie from before this release is refused too: `/pending` then shows a sign-in-again notice in en, fr, es and de instead of the raw error. Sign in once in the browser to confirm again.
- tronc v0.23.1: an open confirmation is reused only for the same OAuth client.

### Changed

- tronc v0.23.1, porte v0.20.1, porte/mcpauth v0.3.1, porte/forfait v0.1.3, tronc/testdb v0.1.2.
- MCP: the state key, the confirm page routes, the account check and the waiting and declined texts come from tronc and porte instead of Vision's own copies. Pending actions resolved or expired more than 30 days ago are pruned every hour.

## [0.9.0] - 2026-11-14

### Added

- Landing: an animated dashboard demo under the hero (mise-en-scene v0.2.2). Visitors and pageviews roll, the live count climbs and the current hour's bar grows. Phones get a compact layout, and reduced motion shows a still frame.
- Landing: a fourth feature block, goals with `vision.track`.

### Changed

- Landing: the hero and its demo sit left-aligned in one column, and every landing string is shorter in en, fr, es and de.
- muse v0.23.0: Magnetik is the typeface for titles and body, and the landing footer starts at the left edge on phones.

## [0.8.2] - 2026-11-13

### Changed

- muse v0.21.0: icon buttons (map zoom and reset, webhook test and edit, team rename) show their label as a muse tooltip on hover and keyboard focus.

## [0.8.1] - 2026-11-13

### Added

- Tracker: `data-exclude="/room/,/share/"` on the `s.js` tag sends nothing (pageview, performance, custom events, heartbeats) while the raw path starts with one of the prefixes, on first load and after every `pushState`, `replaceState` or `popstate` navigation.

## [0.8.0] - 2026-11-13

### Added

- `GET /api/analytics/summary?space=all|<id>&from=&to=`: every site in the caller's workspaces with visitors, pageviews, previous-window counts, live visitors and a zero-filled hourly (up to 48 h) or daily pageview series, from three grouped queries. A space the caller is not a member of answers an empty list. Existing routes are unchanged.
- `/sites` is a cross-space stats overview: `?space=` filter (all by default) and `?period=` (24h, 7d, 30d), totals for visitors, pageviews and live visitors, then one row per site with its space, change against the previous period and a sparkline, busiest first.
- Tracker: a `data-mask` attribute on the `s.js` tag replaces identifier path segments with `:id` and cuts referrers to origin plus masked path (cross-origin: origin only). Without it the payloads are unchanged.
- Vision counts its own production traffic through the masked tracker (`app.html`, only on `vision.facile.studio`).

### Changed

- The sidebar space switcher is gone. The add-site drawer picks the space, and API keys and webhooks in settings get a space picker when the person can write to more than one space.
- Tracker: `s.js` sends nothing from `localhost`, `127.0.0.1`, `*.localhost` or a non-http(s) page.
- `scripts/check.sh` runs the client's bun tests, including a check that runs the served `s.js` against a fake page.

## [0.7.1] - 2026-11-13

### Changed

- chore(deploy): healthchecks run every 30 s with a 1 s start interval.

## [0.7.0] - 2026-11-13

### Added

- Public shared dashboards (`/share/<token>`) may be framed by `https://docs.facile.studio`: that route alone extends the chassis CSP to `frame-ancestors 'self' https://docs.facile.studio` and drops `X-Frame-Options`. Every other page keeps `frame-ancestors 'none'` and `X-Frame-Options: DENY`; `TestOnlySharePagesCanBeFramedByDocs` pins both.
- `?embed=1` on a shared dashboard hides the site title, live dot and "powered by" footer and tightens the padding, so Docs shows only the dashboard.

## [0.6.0] - 2026-11-13

### Changed

- Database: the schema is ordered goose SQL under `apps/api/migrations/`, applied at boot by `tronc/migrate` v0.1.1 behind a Postgres advisory lock. A failed migration stops the boot with exit code 1. GORM `AutoMigrate` and the hand-written `CREATE TABLE IF NOT EXISTS` (porte's tables, `directory_deletion_state`, `mcp.PendingSchema`) are gone from the boot path; the GORM structs stay as the query layer. `TestGORMModelsMatchTheMigratedSchema` fails when a struct names a column the migrations do not create, two drift tests fail when `porte/pg.Schema` or `mcp.PendingSchema` changes a table Vision keeps, and `scripts/check.sh` fails on an `.AutoMigrate(` call outside a test file.
- Database: `00001_baseline.sql` is production's schema as `pg_dump --schema-only` printed it on 2026-11-13, orphan `*_id` columns and the two foreign keys included. A database built before this release (Vision's tables, no applied migration) is stamped at version 1 on its first boot, so the baseline is recorded without running. A database built from the migrations dumps identical to production. `apps/api/schema.sql` is the normalized dump of the whole set.
- Boot: the data steps that ran at every boot still run at every boot, after the migrations and in the same order (`schemas.Reconcile`): the placeholder picture cleanup, the personal workspace seed, and the porte credential adoption (identity re-key, legacy session carry-over, passwords, OIDC subjects).
- build: `github.com/FacileStudio/tronc/migrate` v0.1.1 with `pressly/goose/v3` v3.27.3; goose raises `segmentio/asm` from v1.1.3 to v1.2.1.

### Added

- `docker run <image> migrate status` (also `up`, `version`, `baseline <n>`) answers without a shell.
- Database: `00002_column_defaults.sql` gives every `created_at` and `updated_at` a `now()` default, and the users columns porte/kit does not write the values `fillNewAccount` writes (`''` and `0001-01-01 00:00:00+00`), so a row inserted outside GORM no longer gets NULL there. It has a Down. Rolling back to v0.5.x drops these defaults again, because AutoMigrate removes a default its struct tags do not declare, while the ledger still records version 2: before deploying this release again after such a rollback, run `DELETE FROM goose_db_version WHERE version_id = 2`.

## [0.5.3] - 2026-11-12

### Fixed

- fix(auth): an account re-created for a person Registre erased (a sign-in, a first MCP call or a profile sync racing the erasure) is refused while the subject is in the deletion ledger, and the next poll or push erases any such account again without a restart. The directory sync no longer pre-provisions a member Vision already erased. porte v0.19.0 (directory v0.5.0): `Deletions.Holds = HoldsIdentity` and `oidc.Deps.Erased`.

## [0.5.2] - 2026-11-11

### Fixed

- fix(brand): favicons, touch and manifest icons, the link preview image and `site.webmanifest` come from muse v0.19.2's brand generator, and `app.html` carries its head block: the catalog title "Vision · Web analytics that respects visitors", its description, theme colours and Open Graph and Twitter tags. The landing page no longer overrides the title and description, so every locale shows the catalog's.

## [0.5.1] - 2026-11-11

### Fixed

- fix(deps): vite 7.3.8, @sveltejs/kit 2.70.4 and svelte 5.57.3, plus overrides for devalue 5.9.4, postcss 8.5.32, nanoid 3.3.20, source-map-js 1.2.2 and cookie 0.7.2. Clears every high and medium npm advisory douane reported in the client lockfile.

## [0.5.0] - 2026-11-11

### Added

- A remote MCP server at `POST /mcp` (spec 2026-07-28, stateless, on `tronc/mcp` v0.22.1 and
  `porte/mcpauth` v0.2.1), authorized by Registre tokens bound to
  `https://vision.facile.studio/mcp` with the scopes `vision.read` and `vision.write`. Twelve
  tools: `vision_list_spaces`, `vision_list_sites`, `vision_get_stats`, `vision_list_goals`,
  `vision_create_site`, `vision_update_site`, `vision_unshare_site`, `vision_create_goal`,
  `vision_update_goal`, and the confirmed `vision_delete_site`, `vision_share_site` and
  `vision_delete_goal`. Each calls the service of the matching REST route, so roles hold as in
  the web app. MCP is free on every plan, and a person's first call creates their account.
- `/pending`, where the signed-in person confirms or dismisses what an MCP client that cannot
  ask them asked for, backed by `GET /api/mcp/pending` and its confirm and dismiss routes.
  Only the person's own browser session can confirm.
- `MCP_RESOURCE` (Compose default `https://vision.facile.studio/mcp`) mounts the server when
  SSO is on.

### Changed

- The boot migration creates the `mcp_actions` table (additive). Erasing a person also deletes
  their pending MCP actions.
- tronc v0.22.1, porte v0.18.4.

## [0.4.1] - 2026-11-11

### Security

- build(api): Go 1.26.11 (`go` directive, `golang:1.26.11-alpine` builder, mise pin). It fixes GO-2026-6848 to GO-2026-6859 in the standard library and toolchain, which douane reported as unscored gaps and which failed the gate.

## [0.4.0] - 2026-11-11

### Changed

- Vision ships as one container, the suite's deployment shape. The Go binary mounts every API
  route under `/api`, keeps `/docs`, `/health` and `/ready` at the root, and serves the
  SvelteKit client, now a static build (`adapter-static`, `index.html` fallback), as the
  catch-all. The SvelteKit server, `hooks.server.ts` and both per-app Dockerfiles are gone;
  the root `Dockerfile` builds both halves into `distroless/static-debian12:nonroot`.
- The Compose service is `vision` on port 4000; `vision-api` and `vision-client` are gone. The
  Dokploy domain must point at `vision:4000` before this deploys.
- The tracker's CORS moved into Go: `OPTIONS /api/e/*` answers `204` with
  `Access-Control-Allow-Origin: *` for any origin, and every `/api/e/*` answer carries
  `Access-Control-Allow-Origin: *` and `Cache-Control: no-cache, no-store`. The routes now see
  the visitor's `Origin`, `Referer` and address, which the old proxy dropped, so `POST /api/e/p`
  from a registered site records a pageview instead of meeting SvelteKit's CSRF refusal.
- `/health` and `/ready` answer at the root as well as under `/api`. An unknown page answers
  `200` with the SPA, which renders its own not-found state, instead of a server-rendered `404`.
  `GET /profile` still answers `308` to `/settings`.

### Added

- A rate limit on the public `/e/*` tracker routes: 300 requests per client address, refilling
  at 60 a second, then `429` with `Retry-After`. Production peaked near 5.5 tracker requests a
  second, so a real visitor never meets it.

### Security

- chi v5.3.1 to v5.5.0 and go-jose v4.1.4 to v4.1.9, the two findings douane reported (go-jose
  high, chi medium). The mounted routes are the same before and after.

## [0.3.0] - 2026-11-11

### Added

- Pending Registre space invitations in the user menu. `GET /api/me/invitations` (porte/kit
  `MountInvitations`) answers behind the session when SSO is on, and muse's `SideBar` and
  `MobileNav` draw an Invitations row and an avatar dot.
- `docs/erasure.md`: what Vision deletes when Registre erases a person, and the test that scans
  every text column for their email, name and subject before and after.

### Changed

- Licensed under the GNU Affero General Public License v3.0 only (`AGPL-3.0-only`). See `LICENSE`.
- Libraries: tronc v0.17.1 to v0.21.0, porte v0.13.0 to v0.18.3, porte/directory v0.3.0 to
  v0.4.0, `@facile/muse` v0.16.1 to v0.18.0, Go 1.26.10. No schema change: porte's tables are
  the same.
- Sign-in goes through porte/kit's account store over Vision's `users` table. The stored email
  now changes at sign-in only when Registre marks it verified.
- Erasure runs through porte/kit's `EraseAccount`, which accepts the issuer with or without a
  trailing slash. A shared workspace still named after the erased person is renamed `Workspace`.
- The client's fetch kernel, 401 redirect and theme store are `@facile/client` v0.2.0. A 401 on
  any call but the session probe sends the person to `/login`.
- Gates: `filet.yml` is `preset: facile`, CI pins filet v0.19.2 and runs douane 0.7.2
  (`-fail critical -scope prod`), and `lefthook.yml` replaces `.githooks` (pre-push:
  `filet gate .`). Run `mise run hooks` once in an existing clone.

### Security

- The browser no longer keeps or sends a bearer token from `localStorage`; the session is
  porte's HttpOnly cookie. Any `vision.token` or `token` key an old login left is deleted on load.
- Go 1.26.10 clears the standard library and toolchain advisories douane rated critical.

## [0.2.6] - 2026-11-07

### Changed

- `@facile/muse` v0.14.4 to v0.16.1. Buttons use the `secondary` and `ghost-danger` variants in place of the deprecated `outline` and `danger`.
- `GET /version` is baked into the API image from the version in `apps/client/package.json` (`scripts/version.sh`), so it matches the release tag without any deploy environment change. The `APP_VERSION` build argument is gone.

## [0.2.5] - 2026-11-07

### Changed

- `@facile/muse` v0.14.2 to v0.14.4: the language chosen in the user menu re-renders the whole UI without a reload, and a single 100% slice in a donut chart is drawn as a clean ring.
- The user menu's Subscription row drops its empty `plan` placeholder and shows only its label.

## [0.2.4] - 2026-11-07

### Security

- Registre tokens are sealed at rest. porte to v0.13.0: `porte_identities.access_token` and `refresh_token`
  are encrypted with AES-256-GCM under `PORTE_TOKEN_KEY`. Existing rows are sealed once, in the background,
  after boot, and the count is logged. An invalid `PORTE_TOKEN_KEY` stops the boot. With the key unset nothing changes.
- Losing `PORTE_TOKEN_KEY` makes the stored tokens unreadable: people sign in again, nothing else is lost.
  Do not roll back to a build older than porte v0.13.0 after the first reseal.

## [0.2.3] - 2026-11-07

### Changed

- `@facile/muse` v0.14.1 to v0.14.2: the user menu clamps itself inside the viewport at the right edge of a header.
- The user menu has a Subscription row that opens `https://sso.facile.studio/spaces`, where plans are managed.
- API keys: the create drawer has no expiry field and the list has no expiry column, because Vision keys never expire.

## [0.2.2] - 2026-11-07

### Changed

- One user menu and one settings shell (`@facile/muse` v0.13.0 to v0.14.1). The sidebar card and the mobile avatar open the muse `UserMenu`: account, settings, language, theme, help and sign out. The cog link is gone. Settings use `SettingsLayout` in the canonical order: Profile, Preferences, Notifications, API tokens. Profile is read-only and links to Facile, so the name form and the in-settings sign-out are removed.
- The API keys section is muse `TokensSection`. The API calls are unchanged (same scopes, same workspace, a key is still created from a name and a scope and never expires, so the expiry choice in the drawer is not sent). The endpoint row stays in that section.
- `appearance` is now `preferences` and `api` is now `tokens`; the old addresses redirect. `/settings` shows the section list on a phone.
- A language picked in the menu is stored in the `fc_locale` cookie and wins over the Registre profile locale on the next visit.

## [0.2.1] - 2026-11-07

### Changed

- Libraries: porte v0.12.0 to v0.12.1 and `@facile/muse` v0.12.1 to v0.13.0 (borderless controls with the light `fc-field` fill, `Select` as a custom combobox, badges that never wrap). No app code change: every `Select` sits in a labelled `Field` or carries an `aria-label`.

## [0.2.0] - 2026-11-06

### Added

- The whole interface in French, English, Spanish and German (es and de need native review), with dates, numbers and country names formatted for the session locale
- `GET /version`, set from the `APP_VERSION` build argument
- GDPR deletion: when Registre erases a person, Vision deletes their account and their personal workspaces, and hands the sites and goals they authored in shared workspaces to the next member (`POST /api/registre/deletions` for Registre's signed push, plus a poll)
- `users.locale`, stored from the Registre `locale` claim and returned by `GET /auth/me`
- `.github/workflows/ci.yml`: gofmt, vet, build and race tests against Postgres, client check, tests and build, filet

### Changed

- `PUT /auth/me` saves the name only and ignores `email`: the address comes from Registre, and a user who could set it could squat another person's address and block their first Facile sign-in
- `porte` v0.5.2 to v0.12.0, `porte/directory` v0.1.0 to v0.3.0, `tronc` v0.16.0 to v0.17.1, `@facile/muse` v0.11.0 to v0.12.1
- The login page is muse `AuthScreen`, the landing page uses `LandingHeader` and `LandingFooter`, and every sign-in button is `SignInWithFacile`
- An existing account links to a Facile sign-in only by a verified email, and is refused when it already holds another subject of the same issuer
- `porte_login_codes` gains `provider` and `subject`, and the boot DDL creates `porte_magic_links` and `directory_deletion_state` (all idempotent)

### Removed

- Local sign-up and password sign-in: `POST /auth/register`, `POST /auth/login`, `PUT /auth/password`, the register tab, the password form and the password settings
- The `SSO_ONLY` setting, `internal/authcrypto`, and the English country-name table
- The semantic colour aliases in `app.css`, which only the old login markup used

## [0.1.1] - 2026-11-06

### Added

- Landing page: a "Pricing" section showing the live plans from the public catalog (muse `PricingPlans`, hidden while the catalog is empty)

### Changed

- `@facile/muse` v0.7.4 to v0.11.0

## [0.1.0] - 2026-11-03

First tagged release of the Facile Suite's cookie-free web analytics.

### Added

- Tracking script that collects pageviews through a pixel, with a random `localStorage`
  visitor ID instead of a cookie and no stored IP address
- Referrers, UTM parameters, browser, OS, device, language, screen width, country and
  Navigation-Timing performance numbers
- Live visitor count from a 30-second heartbeat, and new pageviews streamed to the
  dashboard over server-sent events
- Custom events and conversion goals, matched on event name or page path
- Muse-based SvelteKit dashboard with breakdowns, charts and a world map
- Read-only public dashboards shared through a token link
- JSON and CSV export of a date range, and scoped `vis_ro` / `vis_rw` API keys
- Periodic HMAC-signed traffic reports posted to a webhook
- Sign-in through Registre over porte, with member pictures from the `picture` claim
- Workspaces mirrored from Registre shared spaces, managed in Registre

### Changed

- Spaces use the share-circle icon (muse v0.7.4)

[Unreleased]: https://github.com/FacileStudio/Vision/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/FacileStudio/Vision/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/FacileStudio/Vision/releases/tag/v0.1.0
