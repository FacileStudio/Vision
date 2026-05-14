# Vision — Roadmap

Self-hosted analytics. Go + SvelteKit.

## Done

- [x] Pageview tracking (path, referrer, country, language, visitor_id)
- [x] Browser, OS, device detection from user agent
- [x] Screen width tracking
- [x] World map with zoom/pan and tooltips (zoom centers on cursor/viewport)
- [x] Area chart with dual series (pageviews + unique visitors over time)
- [x] Previous period comparison ghost lines on area chart
- [x] Granularity selector on main chart (hourly / daily / weekly / monthly bucketing)
- [x] Top pages, referrers, countries, browsers, OS, devices
- [x] Entry pages tracking (first path per session)
- [x] Exit pages tracking (last path per session)
- [x] Real-time active visitors count (in-memory heartbeats, 30s pings, 60s window)
- [x] Tracker visibility detection (pauses heartbeats on hidden tabs)
- [x] Extended date presets (today, this week, 7d, this month, 30d, 90d, this year, custom range)
- [x] Human-readable date range label
- [x] KPI cards with trend indicators (% change vs previous period)
- [x] Views per visitor derived metric
- [x] Traffic sources donut chart (direct, search, social, other)
- [x] Hourly distribution bar chart (traffic by hour of day)
- [x] Pie charts for devices and screens
- [x] Bento grid layout (3-column asymmetric for bottom stats)
- [x] Clean neutral light theme (Sablier-style, solid opaque cards)
- [x] Favicon next to referrer URLs and site logos on dashboard/sites list/stats header
- [x] Copy tracking script snippet
- [x] Profile page (name, email, change password)
- [x] Settings page (system-wide webhook configuration)
- [x] Sidebar user card with initials and profile link (reactive on update)
- [x] Page titles via svelte:head on all routes
- [x] OIDC/SSO authentication
- [x] Ad-blocker resistant tracking (renamed paths)
- [x] Image pixel tracking (CORS-free)
- [x] Live polling dashboard (5s overview, 10s realtime)
- [x] Site management (CRUD)
- [x] Session tracking (30min gap, bounce rate, avg duration, pages/session)
- [x] UTM parameter tracking (utm_source, utm_medium, utm_campaign, utm_term, utm_content)
- [x] UTM breakdown dashboard section (sources, mediums, campaigns)
- [x] Public shareable dashboard links (token-based, read-only, full dashboard)
- [x] Data export (CSV/JSON download with auth)
- [x] Webhook periodic reports (hourly/daily/weekly/monthly) with HMAC-SHA256 signing
- [x] Nook integration (Vision webhook provider for event aggregation)
- [x] Solar Linear / Iconify icons across UI
- [x] Composite index on pageviews(site_id, created_at) for query performance
- [x] Session race condition fix (GORM transaction)
- [x] Dashboard filters (country, browser, OS, device, path, referrer — clickable stats apply filters)
- [x] Custom event tracking (`vision.track('signup', { plan: 'pro' })` via pixel endpoint)
- [x] Custom events dashboard section (top events by count)
- [x] Page load performance tracking (DNS, TCP, TTFB, DOM load, page load via Navigation Timing API)
- [x] Performance metrics dashboard section (averages with sample count)

## Medium-term

- [ ] Goal tracking (mark paths or events as goals, show conversion rates)
- [ ] API keys for programmatic access
- [ ] Multi-user / team access with roles (viewer, admin)
- [ ] Email reports (weekly/monthly summary via scheduled job)
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
