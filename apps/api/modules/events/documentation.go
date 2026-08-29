package events

import (
	"net/http"

	documentation "github.com/FacileStudio/Vision/apps/api/internal/documentation"
)

var Documentation = documentation.Module{
	Name:        "events",
	Description: "Public event ingestion and realtime streaming. Site resolved from Origin/Referer header for ingestion — only registered domains are accepted.",
	Routes: []documentation.Route{
		{
			Method:      "GET",
			Path:        "/e/p",
			Summary:     "Record a pageview via GET pixel",
			Description: "Records a pageview for the site matching the request Origin via transparent pixel query data.",
			Auth:        "none (domain-based)",
		},
		{
			Method:      "POST",
			Path:        "/e/p",
			Summary:     "Record a pageview",
			Description: "Records a pageview for the site matching the request Origin. Rejects unregistered domains.",
			Auth:        "none (domain-based)",
			RequestBody: PageviewRequest{},
			Status:      http.StatusNoContent,
		},
		{
			Method:  "OPTIONS",
			Path:    "/e/p",
			Summary: "Pageview preflight CORS handler",
			Status:  http.StatusNoContent,
		},
		{
			Method:      "GET",
			Path:        "/e/t",
			Summary:     "Record custom event via GET pixel",
			Description: "Records custom event parameters via transparent pixel query data.",
			Auth:        "none (domain-based)",
		},
		{
			Method:      "GET",
			Path:        "/e/h",
			Summary:     "Record visitor heartbeat ping",
			Description: "Records active visitor presence for live count.",
			Auth:        "none (domain-based)",
		},
		{
			Method:       "GET",
			Path:         "/events/{siteId}/live",
			Summary:      "Live pageview stream (SSE)",
			Description:  "Server-Sent Events stream of pageview events for the given site. Sends JSON-encoded PageviewEvent objects. Keepalive comment every 30 seconds.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "siteId", Type: "int", Description: "Site ID"}},
			ResponseBody: PageviewEvent{},
		},
	},
}
