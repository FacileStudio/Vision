package events

import documentation "api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "events",
	Description: "Public event ingestion. Site resolved from Origin/Referer header — only registered domains are accepted.",
	Routes: []documentation.Route{
		{
			Method:      "POST",
			Path:        "/event/pageview",
			Summary:     "Record a pageview",
			Description: "Records a pageview for the site matching the request Origin. Rejects unregistered domains.",
			Auth:        "none (domain-based)",
			RequestBody: "PageviewRequest",
		},
	},
}
