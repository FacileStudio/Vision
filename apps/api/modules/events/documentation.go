package events

import documentation "api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "events",
	Description: "Public event ingestion endpoints. Authenticated by site API key.",
	Routes: []documentation.Route{
		{
			Method:      "POST",
			Path:        "/event/pageview",
			Summary:     "Record a pageview",
			Description: "Records a single pageview event for the site identified by the API key. Called from the tracking script embedded on client websites.",
			Auth:        "api_key (Bearer or ?key= query param)",
			RequestBody: "PageviewRequest",
		},
	},
}
