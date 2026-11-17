package events

import (
	"net/http"

	documentation "github.com/FacileStudio/Vision/apps/api/internal/documentation"
)

var Documentation = documentation.Module{
	Name:        "events",
	Description: "Public event ingestion and realtime streaming. Site resolved from Origin/Referer header for ingestion — only registered domains are accepted. Requests from bots, crawlers and automated browsers (classified from the User-Agent and the payload's `b` flags: bit 1 navigator.webdriver, bit 2 a 0x0 outer window) never count as pageviews; they are counted per day in bot_hits instead.",
	Routes: []documentation.Route{
		{
			Method:      "GET",
			Path:        "/e/p",
			Summary:     "Record a pageview via GET pixel",
			Description: "Records a pageview for the site matching the request Origin via transparent pixel query data (a JSON PageviewRequest in ?data=, including the `b` bot flags). A bot's request is counted in bot_hits instead. With ?type=perf it stores the timings of the visitor's last pageview, never for a bot.",
			Auth:        "none (domain-based)",
		},
		{
			Method:      "POST",
			Path:        "/e/p",
			Summary:     "Record a pageview",
			Description: "Records a pageview for the site matching the request Origin. Rejects unregistered domains. A bot's request answers the same 204 but is counted in bot_hits instead.",
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
			Description: "Records custom event parameters via transparent pixel query data (a JSON CustomEventRequest in ?data=, including the `b` bot flags). A bot's event is dropped.",
			Auth:        "none (domain-based)",
		},
		{
			Method:      "GET",
			Path:        "/e/h",
			Summary:     "Record visitor heartbeat ping",
			Description: "Records active visitor presence for live count. ?data= carries a JSON HeartbeatRequest: hostname, visitor_id, `b` bot flags and `interacted`, which marks the visitor's open session as having seen a real gesture. A bot's heartbeat is dropped.",
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
