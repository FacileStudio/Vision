# Changelog

All notable changes to Vision are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
