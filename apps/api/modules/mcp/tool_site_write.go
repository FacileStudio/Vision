package mcp

import (
	"context"
	stderrors "errors"
	"strconv"

	"github.com/FacileStudio/Vision/apps/api/modules/sites"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	statusCreated  = "created"
	statusUpdated  = "updated"
	statusUnshared = "unshared"
	installHint    = "Add tracking_snippet to every page of the site, in <head>. Visits show up in vision_get_stats " +
		"once the script runs on the domain."
)

var errNoChange = stderrors.New("pass at least one field to change")

// createSite is vision_create_site, through POST /sites.
func (t *toolset) createSite(ctx context.Context, req *sdk.CallToolRequest, in createSiteInput) (*sdk.CallToolResult, siteOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, siteOutput{}, err
	}
	site, err := t.Sites.Create(ctx, userID, &sites.CreateRequest{Name: in.Name, Domain: in.Domain, WorkspaceID: in.SpaceID})
	if err != nil {
		return nil, siteOutput{}, mcpkit.Failure(err, hintSpace)
	}
	out := t.siteOf(site)
	return nil, siteOutput{Status: statusCreated, Site: &out, TrackingSnippet: t.snippet(), Message: installHint}, nil
}

// updateSite is vision_update_site, through PUT /sites/{id}: fields left out keep their
// current value.
func (t *toolset) updateSite(ctx context.Context, req *sdk.CallToolRequest, in updateSiteInput) (*sdk.CallToolResult, siteOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, siteOutput{}, err
	}
	if in.Name == "" && in.Domain == "" {
		return nil, siteOutput{}, errNoChange
	}
	current, err := t.writableSite(ctx, userID, in.SiteID)
	if err != nil {
		return nil, siteOutput{}, err
	}
	change := sites.UpdateRequest{Name: current.Name, Domain: current.Domain}
	if in.Name != "" {
		change.Name = in.Name
	}
	if in.Domain != "" {
		change.Domain = in.Domain
	}
	site, err := t.Sites.Update(ctx, userID, strconv.FormatInt(in.SiteID, 10), &change)
	if err != nil {
		return nil, siteOutput{}, mcpkit.Failure(err, hintSite)
	}
	out := t.siteOf(site)
	return nil, siteOutput{Status: statusUpdated, Site: &out}, nil
}

// unshareSite is vision_unshare_site, through DELETE /sites/{id}/share. Closing a public
// link only takes access away, so it runs without a confirmation.
func (t *toolset) unshareSite(ctx context.Context, req *sdk.CallToolRequest, in siteInput) (*sdk.CallToolResult, siteOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, siteOutput{}, err
	}
	if _, err := t.writableSite(ctx, userID, in.SiteID); err != nil {
		return nil, siteOutput{}, err
	}
	if err := t.Sites.Unshare(ctx, userID, strconv.FormatInt(in.SiteID, 10)); err != nil {
		return nil, siteOutput{}, mcpkit.Failure(err, hintSite)
	}
	site, err := t.site(ctx, userID, in.SiteID)
	if err != nil {
		return nil, siteOutput{}, err
	}
	out := t.siteOf(site)
	return nil, siteOutput{Status: statusUnshared, Site: &out, Message: "The public link no longer works."}, nil
}
