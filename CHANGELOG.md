# Changelog

All notable changes to Vision are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/FacileStudio/Vision/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/FacileStudio/Vision/releases/tag/v0.1.0
