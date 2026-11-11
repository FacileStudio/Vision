package mcp

import (
	"context"
	"strconv"
	"time"

	"github.com/FacileStudio/Vision/apps/api/modules/goals"
	"github.com/FacileStudio/Vision/apps/api/modules/sites"
	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// The sensitive actions: deleting a site or a goal cannot be undone, and a public link
// shows a site's dashboard to anyone who has it.
const (
	kindDeleteSite = "delete_site"
	kindShareSite  = "share_site"
	kindDeleteGoal = "delete_goal"
)

var errTarget = errors.NotFound("this action names no site or goal")

// kinds are the sensitive actions. Each fingerprint covers the target as it was when the
// person was asked, so a change in between closes the action instead of running it on
// something the person did not see.
func (t *toolset) kinds() map[string]mcpkit.Kind {
	return map[string]mcpkit.Kind{
		kindDeleteSite: {Tool: toolDeleteSite, Fingerprint: t.siteFingerprint, Run: t.runDeleteSite},
		kindShareSite:  {Tool: toolShare, Fingerprint: t.siteFingerprint, Run: t.runShareSite},
		kindDeleteGoal: {Tool: toolDeleteGoal, Fingerprint: t.goalFingerprint, Run: t.runDeleteGoal},
	}
}

// refOf parses an action's user and target into the decimal ids the services take.
func refOf(a mcpkit.Action) (string, string, error) {
	userID, targetID, err := mcpkit.TargetIDs(a, errTarget)
	if err != nil {
		return "", "", err
	}
	return strconv.FormatInt(userID, 10), strconv.FormatInt(targetID, 10), nil
}

// siteFingerprint digests the targeted site as it is now, through the reach of the write
// routes: a site the person can change.
func (t *toolset) siteFingerprint(ctx context.Context, a mcpkit.Action) (string, error) {
	userID, siteID, err := refOf(a)
	if err != nil {
		return "", err
	}
	id, _ := strconv.ParseInt(siteID, 10, 64)
	site, err := t.writableSite(ctx, userID, id)
	if err != nil {
		return "", err
	}
	return siteDigest(site), nil
}

// siteDigest fingerprints a site, its public link included.
func siteDigest(s *sites.SiteResponse) string {
	shared := ""
	if s.ShareToken != nil {
		shared = *s.ShareToken
	}
	return mcpkit.Digest(strconv.FormatInt(s.ID, 10), s.Name, s.Domain, strconv.FormatInt(s.WorkspaceID, 10),
		shared, s.UpdatedAt.UTC().Format(time.RFC3339Nano))
}

// goalFingerprint digests the targeted goal as it is now: one the person created.
func (t *toolset) goalFingerprint(ctx context.Context, a mcpkit.Action) (string, error) {
	userID, goalID, err := refOf(a)
	if err != nil {
		return "", err
	}
	goal, err := t.Goals.Own(ctx, userID, goalID)
	if err != nil {
		return "", err
	}
	return goalDigest(goal), nil
}

// goalDigest fingerprints a goal.
func goalDigest(g *goals.GoalResponse) string {
	out := goalOf(g)
	return mcpkit.Digest(strconv.FormatInt(g.ID, 10), strconv.FormatInt(g.SiteID, 10), out.Name, out.GoalType,
		out.PagePath, out.MatchType, out.EventName)
}
