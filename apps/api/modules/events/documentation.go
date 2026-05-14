package events

import documentation "api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "events",
	Description: "Public event ingestion and realtime streaming. Site resolved from Origin/Referer header for ingestion — only registered domains are accepted.",
	Routes: []documentation.Route{
		{
			Method:      "POST",
			Path:        "/e/p",
			Summary:     "Record a pageview",
			Description: "Records a pageview for the site matching the request Origin. Rejects unregistered domains.",
			Auth:        "none (domain-based)",
			RequestBody: "PageviewRequest",
		},
		{
			Method:      "GET",
			Path:        "/events/{siteId}/live",
			Summary:     "Live pageview stream (SSE)",
			Description: "Server-Sent Events stream of pageview events for the given site. Sends JSON-encoded PageviewEvent objects. Keepalive comment every 30 seconds.",
			Auth:        "bearer token",
		},
	},
}
