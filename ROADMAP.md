# Vision — Roadmap

Self-hosted analytics. Go + SvelteKit.

## Done

- [x] Pageview tracking (path, referrer, country, language, visitor_id)
- [x] Browser, OS, device detection from user agent
- [x] Screen width tracking
- [x] World map with zoom/pan and tooltips
- [x] Area chart with dual series (pageviews + unique visitors over time)
- [x] Top pages, referrers, countries, browsers, OS, devices
- [x] Real-time active visitors count (in-memory heartbeats, 30s pings, 60s window)
- [x] Tracker visibility detection (pauses heartbeats on hidden tabs)
- [x] Extended date presets (today, this week, 7d, this month, 30d, 90d, this year, custom range)
- [x] Custom date range with inline date inputs
- [x] Human-readable date range label
- [x] KPI cards with trend indicators (% change vs previous period)
- [x] Views per visitor derived metric
- [x] Previous period comparison data from backend
- [x] Unique visitors per day from backend
- [x] Traffic sources donut chart (direct, search, social, other)
- [x] Hourly distribution bar chart (traffic by hour of day)
- [x] Pie charts for devices and screens
- [x] Bento grid layout (3-column asymmetric for bottom stats)
- [x] Dark mode
- [x] Pastel glass color system (translucent cards, backdrop blur, OKLch pastels)
- [x] Map zoom centers on cursor (scroll) and viewport center (buttons)
- [x] Favicon next to referrer URLs and site logos on dashboard/sites list
- [x] Copy tracking script snippet
- [x] Settings page (profile, password)
- [x] OIDC/SSO authentication
- [x] Ad-blocker resistant tracking (renamed paths)
- [x] Image pixel tracking (CORS-free)
- [x] Live polling dashboard (5s overview, 10s realtime)
- [x] Site management (CRUD)
- [x] Session tracking (30min gap, bounce rate, avg duration, pages/session)
- [x] Public shareable dashboard links (token-based, read-only, full dashboard)
- [x] Webhook periodic reports (hourly/daily/weekly/monthly) with HMAC-SHA256 signing
- [x] System-wide webhook configuration (reports for all sites per webhook)
- [x] Webhook settings UI in settings page (CRUD, toggle, test)
- [x] Nook integration (Vision webhook provider for event aggregation)
- [x] Solar Linear / Iconify icons across UI

## Short-term

- [ ] UTM parameter tracking (utm_source, utm_medium, utm_campaign parsed from URL)
- [ ] Entry pages tracking (first path per session — data already in visitor_sessions)
- [ ] Exit pages tracking (last path per session — data already in visitor_sessions)
- [ ] Dashboard filters (filter by country, browser, path, device, referrer)
- [ ] Custom event tracking (`vision.track('signup', { plan: 'pro' })`)
- [ ] Comparison view (previous period ghost overlay on area chart)
- [ ] Granularity selector on main chart (hourly / daily / weekly / monthly bucketing)
- [ ] Data export (CSV/JSON for raw or aggregated data)

## Medium-term

- [ ] Goal tracking (mark paths or events as goals, show conversion rates)
- [ ] API keys for programmatic access
- [ ] Multi-user / team access with roles (viewer, admin)
- [ ] Email reports (weekly/monthly summary via scheduled job)
- [ ] Page load performance tracking (measure timing from tracker)
- [ ] Sparklines in KPI cards (tiny area trends)
- [ ] Lightweight embeddable widget (small stats badge for sites)
- [ ] "See all" drawers for top pages / referrers (paginated, searchable)

## Long-term / Infrastructure

- [ ] ClickHouse or TimescaleDB for high-volume analytics
- [ ] Batch inserts (buffer in memory, flush every N seconds)
- [ ] CDN-hosted tracker script (jsDelivr/unpkg) to eliminate Private Network Access issues
- [ ] Data retention policies (auto-delete after N months)
- [ ] GDPR compliance tools (data deletion requests, consent mode)
- [ ] Webhooks for real-time event forwarding (not just periodic reports)
- [ ] Import data from other analytics tools (Umami, Plausible, GA)
- [ ] Mobile app / PWA for dashboard
