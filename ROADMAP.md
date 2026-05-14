# Vision — Roadmap

Self-hosted analytics. Go + SvelteKit.

## Done

- [x] Pageview tracking (path, referrer, country, language, visitor_id)
- [x] Browser, OS, device detection from user agent
- [x] Screen width tracking
- [x] World map with zoom/pan and tooltips
- [x] Line chart (pageviews over time)
- [x] Top pages, referrers, countries, browsers, OS, devices
- [x] Real-time active visitors count
- [x] Date range picker (today, 7d, 30d, 90d)
- [x] Dark mode
- [x] Favicon next to referrer URLs
- [x] Copy tracking script snippet
- [x] Settings page (profile, password)
- [x] OIDC/SSO authentication
- [x] Ad-blocker resistant tracking (renamed paths)
- [x] Image pixel tracking (CORS-free)
- [x] Live polling dashboard (5s)
- [x] Site management (CRUD)

## Short-term

- [ ] Session tracking (group pageviews by visitor with <30min gap, enables bounce rate, pages/session, avg duration)
- [ ] UTM parameter tracking (utm_source, utm_medium, utm_campaign parsed from URL)
- [ ] Custom event tracking (`vision.track('signup', { plan: 'pro' })`)
- [ ] Data export (CSV/JSON for raw or aggregated data)
- [ ] Comparison view (this week vs last week overlay on charts)
- [ ] Goal tracking (mark paths as goals, show conversion rates)
- [ ] Public shareable dashboard link (read-only, no auth)

## Medium-term

- [ ] API keys for programmatic access
- [ ] Multi-user / team access with roles (viewer, admin)
- [ ] Email reports (weekly/monthly summary via scheduled job)
- [ ] Page load performance tracking (measure timing from tracker)
- [ ] Exit pages tracking
- [ ] Entry pages tracking
- [ ] Filter/search within dashboard (filter by country, browser, path, etc.)
- [ ] Lightweight embeddable widget (small stats badge for sites)

## Long-term / Infrastructure

- [ ] ClickHouse or TimescaleDB for high-volume analytics
- [ ] Batch inserts (buffer in memory, flush every N seconds)
- [ ] CDN-hosted tracker script (jsDelivr/unpkg) to eliminate Private Network Access issues
- [ ] Data retention policies (auto-delete after N months)
- [ ] GDPR compliance tools (data deletion requests, consent mode)
- [ ] Webhooks (notify external services on events)
- [ ] Import data from other analytics tools (Umami, Plausible, GA)
- [ ] Mobile app / PWA for dashboard
