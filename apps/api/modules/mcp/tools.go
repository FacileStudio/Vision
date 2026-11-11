package mcp

import (
	"context"
	stderrors "errors"
	"strconv"
	"time"

	"github.com/FacileStudio/Vision/apps/api/internal/siteaccess"
	"github.com/FacileStudio/Vision/apps/api/modules/sites"
	"github.com/FacileStudio/porte/mcpauth"
	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	toolSpaces     = "vision_list_spaces"
	toolSites      = "vision_list_sites"
	toolStats      = "vision_get_stats"
	toolGoals      = "vision_list_goals"
	toolCreateSite = "vision_create_site"
	toolUpdateSite = "vision_update_site"
	toolDeleteSite = "vision_delete_site"
	toolShare      = "vision_share_site"
	toolUnshare    = "vision_unshare_site"
	toolCreateGoal = "vision_create_goal"
	toolUpdateGoal = "vision_update_goal"
	toolDeleteGoal = "vision_delete_goal"
	hintSite       = "Call vision_list_sites for the ids of the sites you can see."
	hintSpace      = "Call vision_list_spaces for the ids of your spaces."
	hintGoal       = "Call vision_list_goals with the site_id for its goal ids; only the person who created a goal can change it."
	dayLayout      = "2006-01-02"
	defaultDays    = 30
)

var (
	errNoCaller = stderrors.New("this request carries no valid access token. Reconnect Vision")
	errDate     = stderrors.New("from and to must be dates such as 2026-11-01")
	errRange    = stderrors.New("from must not be after to")
	errNoSpace  = errors.NotFound("space not found")
	errReadOnly = errors.Forbidden("you can read this site but not change it: only owners, admins and editors of its space can")
)

// toolset holds what the tool handlers share.
type toolset struct {
	Deps
	confirm *mcpkit.Confirmations
	origin  string
}

// userOf reads the Vision user id from the verified token, in the decimal form the
// services take, refused the way RequireAuth refuses it: a deleted account never acts.
func (t *toolset) userOf(ctx context.Context, req *sdk.CallToolRequest) (string, error) {
	identity, ok := mcpauth.Identity(req.Extra.TokenInfo)
	if !ok {
		return "", errNoCaller
	}
	if _, err := t.Accounts.Identity(ctx, identity.UserID); err != nil {
		return "", errNoCaller
	}
	return strconv.FormatInt(identity.UserID, 10), nil
}

// site reads a site the person can see, as GET /sites/{id} does.
func (t *toolset) site(ctx context.Context, userID string, siteID int64) (*sites.SiteResponse, error) {
	site, err := t.Sites.Get(ctx, userID, strconv.FormatInt(siteID, 10))
	if err != nil {
		return nil, mcpkit.Failure(err, hintSite)
	}
	return site, nil
}

// writableSite reads a site the person can change, refusing a viewer before anything is
// asked or written.
func (t *toolset) writableSite(ctx context.Context, userID string, siteID int64) (*sites.SiteResponse, error) {
	site, err := t.site(ctx, userID, siteID)
	if err != nil {
		return nil, err
	}
	uid, _ := strconv.ParseInt(userID, 10, 64)
	if !siteaccess.CanWrite(ctx, t.DB, uid, siteID) {
		return nil, mcpkit.Failure(errReadOnly, "")
	}
	return site, nil
}

// window is the date range of a report, whole days in UTC, as the dashboard reads it:
// the last 30 days by default, and to includes that whole day.
type window struct {
	from, to time.Time
}

// windowOf parses from and to.
func windowOf(from, to string, now time.Time) (window, error) {
	end := now.UTC()
	start := end.AddDate(0, 0, -defaultDays)
	if from != "" {
		day, err := time.Parse(dayLayout, from)
		if err != nil {
			return window{}, errDate
		}
		start = day
	}
	if to != "" {
		day, err := time.Parse(dayLayout, to)
		if err != nil {
			return window{}, errDate
		}
		end = day.Add(24*time.Hour - time.Nanosecond)
	}
	if start.After(end) {
		return window{}, errRange
	}
	return window{from: start, to: end}, nil
}

// out renders a window as the days it covers.
func (w window) out() rangeOut {
	return rangeOut{From: w.from.Format(dayLayout), To: w.to.Format(dayLayout)}
}

// target is the pending-action reference of a site or a goal.
func target(id int64) string {
	return strconv.FormatInt(id, 10)
}
