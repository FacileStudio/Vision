package mcp

import (
	"strconv"

	"github.com/FacileStudio/Vision/apps/api/modules/sites"
)

// shareURL is the public dashboard of a shared site, or empty.
func (t *toolset) shareURL(site *sites.SiteResponse) string {
	if site.ShareToken == nil || *site.ShareToken == "" {
		return ""
	}
	return t.origin + "/share/" + *site.ShareToken
}

// siteURL is the site's dashboard in the web app.
func (t *toolset) siteURL(siteID int64) string {
	return t.origin + "/sites/" + strconv.FormatInt(siteID, 10)
}

// snippet is the script tag that tracks a site.
func (t *toolset) snippet() string {
	return `<script defer src="` + t.origin + `/s.js"></script>`
}
