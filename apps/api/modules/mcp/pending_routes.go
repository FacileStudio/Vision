package mcp

import (
	"context"
	"net/http"
	"strconv"

	"github.com/FacileStudio/porte/mcpauth"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

const webActor = "vision-web"

// pendingUser is the signed-in person of a confirm page request. Only the person's own
// browser session passes: an API key, a bearer, a CLI session or a delegated token must
// never confirm what an agent asked for.
func pendingUser(r *http.Request) (string, error) {
	identity, err := mcpauth.RequireOwnSession(r)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(identity.UserID, 10), nil
}

// describe adds the targeted site or goal, as it is now, to a confirm page item.
func (t *toolset) describe(ctx context.Context, item mcpkit.PendingItem, a mcpkit.Action) PendingItem {
	return PendingItem{PendingItem: item, Target: t.pendingTarget(ctx, a)}
}

// done is the confirm page's answer once an action ran: a deletion, or the public link
// the person just opened.
func (t *toolset) done(ctx context.Context, o mcpkit.Outcome) mcpkit.PendingResult {
	if o.Action.Kind != kindShareSite {
		return mcpkit.PendingResult{Status: statusDeleted}
	}
	result := mcpkit.PendingResult{Status: statusShared}
	if userID, siteID, err := refOf(o.Action); err == nil {
		id, _ := strconv.ParseInt(siteID, 10, 64)
		if site, err := t.site(ctx, userID, id); err == nil {
			result.URL = t.shareURL(site)
		}
	}
	return result
}

// pendingTarget describes what an action targets, or nil once it is gone.
func (t *toolset) pendingTarget(ctx context.Context, a mcpkit.Action) *PendingTarget {
	userID, targetID, err := refOf(a)
	if err != nil {
		return nil
	}
	if a.Kind == kindDeleteGoal {
		goal, err := t.Goals.Own(ctx, userID, targetID)
		if err != nil {
			return nil
		}
		out := &PendingTarget{Type: "goal", ID: goal.ID, Name: goal.Name, URL: t.siteURL(goal.SiteID)}
		if site, err := t.site(ctx, userID, goal.SiteID); err == nil {
			out.Site, out.Domain = site.Name, site.Domain
		}
		return out
	}
	id, _ := strconv.ParseInt(targetID, 10, 64)
	site, err := t.site(ctx, userID, id)
	if err != nil {
		return nil
	}
	return &PendingTarget{Type: "site", ID: site.ID, Name: site.Name, Domain: site.Domain, URL: t.siteURL(site.ID)}
}
