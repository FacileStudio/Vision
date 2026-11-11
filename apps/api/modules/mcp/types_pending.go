package mcp

import mcpkit "github.com/FacileStudio/tronc/mcp"

// PendingItem is one action waiting on the confirm page.
type PendingItem struct {
	mcpkit.PendingItem
	Target *PendingTarget `json:"target,omitempty"`
}

// PendingTarget is the site or goal an action targets, as it is now. Site names the site
// a goal belongs to.
type PendingTarget struct {
	Type   string `json:"type"`
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain,omitempty"`
	Site   string `json:"site,omitempty"`
	URL    string `json:"url"`
}
