package main

import (
	"net/http"

	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/go-chi/chi/v5"
)

// shareFrameAncestors lists who may frame a public shared dashboard: Vision
// itself and Docs, which embeds it with ?embed=1.
var shareFrameAncestors = []string{"'self'", "https://docs.facile.studio"}

// mountClient serves the built client on every path the API left free. Only
// the public dashboards under /share/ may be framed; every other page keeps
// the chassis's frame-ancestors 'none' and X-Frame-Options: DENY.
func mountClient(router chi.Router, client http.Handler) {
	framable := troncmiddleware.SecurityOverride(troncmiddleware.SecurityOverrideConfig{
		CSPSources: map[string][]string{"frame-ancestors": shareFrameAncestors},
	})
	router.Handle("/share/*", framable(client))
	router.Handle("/*", client)
}
