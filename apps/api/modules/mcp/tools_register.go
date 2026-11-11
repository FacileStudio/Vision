package mcp

import (
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// list is the curated tool set. Reports are open world: visitors of the tracked sites
// write the paths, referrers and event names. Deleting and sharing are the sensitive tools.
func (t *toolset) list() []mcpkit.Tool {
	return append(append(t.reads(), t.writes()...), t.sensitive()...)
}

// reads are the four read-only tools.
func (t *toolset) reads() []mcpkit.Tool {
	return []mcpkit.Tool{
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolSpaces, Title: "List spaces",
			Description: "List the spaces (teams) the user belongs to, with the user's role (owner, admin, editor or " +
				"viewer) and how many sites and members each has. Use a space id to create a site or narrow vision_list_sites.",
			Annotations: &mcpkit.Annotations{ReadOnly: true, Idempotent: true},
		}, t.listSpaces),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolSites, Title: "List sites",
			Description: "List the tracked sites the user can see, newest first within each space, with their id, " +
				"domain, space, dashboard url and public share_url when one is open. Default is every space; pass " +
				"space_id or query (name or domain) to narrow it.",
			Annotations: &mcpkit.Annotations{ReadOnly: true, Idempotent: true},
		}, t.listSites),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolStats, Title: "Get site analytics",
			Description: "Report one site's traffic for a range of whole UTC days (default the last 30): pageviews, " +
				"unique visitors, the same for the previous period of equal length, bounce rate (percent), average " +
				"session length, pages per session, visitors on the site right now, and the top 10 pages, referrers, " +
				"countries, devices, browsers and custom events. Filters (country, browser, os, device, path, referrer) " +
				"narrow every number. response_format detailed adds the time series at granularity, operating systems, " +
				"screens, entry and exit pages, UTM sources, mediums and campaigns, and page-load timings.",
			Annotations: &mcpkit.Annotations{ReadOnly: true, Idempotent: true, OpenWorld: true},
		}, t.getStats),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolGoals, Title: "List goals and conversions",
			Description: "List one site's goals with how many distinct visitors converted in a range of whole UTC days " +
				"(default the last 30) and the conversion rate in percent of the range's visitors. A pageview goal " +
				"matches a path; an event goal matches a custom event name.",
			Annotations: &mcpkit.Annotations{ReadOnly: true, Idempotent: true, OpenWorld: true},
		}, t.listGoals),
	}
}

// writes are the tools that add or change sites and goals without losing any data.
func (t *toolset) writes() []mcpkit.Tool {
	return []mcpkit.Tool{
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolCreateSite, Title: "Add a site",
			Description: "Start tracking a site in a space where the user is owner, admin or editor. Returns the site " +
				"and tracking_snippet, the script tag to put in the <head> of every page. A domain can be tracked once " +
				"across Vision.",
			Annotations: &mcpkit.Annotations{},
		}, t.createSite),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolUpdateSite, Title: "Edit a site",
			Description: "Rename a site or change its domain. Fields left out stay as they are. Changing the domain " +
				"means visits from the old domain stop counting.",
			Annotations: &mcpkit.Annotations{Destructive: true, Idempotent: true},
		}, t.updateSite),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolUnshare, Title: "Close a public link",
			Description: "Close a site's public dashboard link, so nobody can open it any more. Sharing again makes " +
				"a new link. With no link open it changes nothing.",
			Annotations: &mcpkit.Annotations{Destructive: true, Idempotent: true},
		}, t.unshareSite),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolCreateGoal, Title: "Add a goal",
			Description: "Add a conversion goal to a site the user can change. goal_type pageview needs page_path " +
				"(match_type exact, starts_with or contains); goal_type event needs event_name, the name the site's " +
				"script sends. Conversions count from the data already collected.",
			Annotations: &mcpkit.Annotations{},
		}, t.createGoal),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolUpdateGoal, Title: "Edit a goal",
			Description: "Change one of the user's own goals: its name, type, path, match type or event name. Fields " +
				"left out stay as they are.",
			Annotations: &mcpkit.Annotations{Destructive: true, Idempotent: true},
		}, t.updateGoal),
	}
}

// sensitive are the tools the user confirms.
func (t *toolset) sensitive() []mcpkit.Tool {
	return []mcpkit.Tool{
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolDeleteSite, Title: "Delete a site",
			Description: "Stop tracking a site and delete it for good. The user always confirms: in your client when " +
				"it can ask, otherwise the result has status pending_confirmation and a confirm_url to give the user. " +
				"Nothing is deleted before they confirm.",
			Annotations: &mcpkit.Annotations{Destructive: true, Idempotent: true},
		}, t.deleteSite),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolShare, Title: "Open a public link",
			Description: "Open a public link to a site's dashboard that anyone can view without signing in; an " +
				"older link stops working. The user always confirms, as for vision_delete_site; once confirmed the " +
				"result carries share_url.",
			Annotations: &mcpkit.Annotations{Destructive: true},
		}, t.shareSite),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolDeleteGoal, Title: "Delete a goal",
			Description: "Delete one of the user's own goals for good. The user always confirms, as for " +
				"vision_delete_site. Nothing is deleted before they confirm.",
			Annotations: &mcpkit.Annotations{Destructive: true, Idempotent: true},
		}, t.deleteGoal),
	}
}
