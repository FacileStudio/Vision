package mcp

import (
	"context"
	"strings"

	"github.com/FacileStudio/Vision/apps/api/modules/sites"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const sitesCapped = "More sites match than fit in one answer. Pass space_id or query to narrow the list."

// listSpaces is vision_list_spaces.
func (t *toolset) listSpaces(ctx context.Context, req *sdk.CallToolRequest, _ spacesInput) (*sdk.CallToolResult, spacesOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, spacesOutput{}, err
	}
	spaces, err := t.Workspaces.List(ctx, userID)
	if err != nil {
		return nil, spacesOutput{}, mcpkit.Failure(err, "")
	}
	out := spacesOutput{Spaces: make([]spaceOut, len(spaces))}
	for i, space := range spaces {
		out.Spaces[i] = spaceOut{ID: space.ID, Name: space.Name, Role: space.Role, Sites: space.SiteCount, Members: space.MemberCount}
	}
	return nil, out, nil
}

// listSites is vision_list_sites: the sites of one space, or of every space the person
// belongs to, as the Sites page lists them space by space.
func (t *toolset) listSites(ctx context.Context, req *sdk.CallToolRequest, in sitesInput) (*sdk.CallToolResult, sitesOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, sitesOutput{}, err
	}
	spaces, err := t.Workspaces.List(ctx, userID)
	if err != nil {
		return nil, sitesOutput{}, mcpkit.Failure(err, "")
	}
	query := strings.ToLower(strings.TrimSpace(in.Query))
	found := false
	all := []siteOut{}
	for _, space := range spaces {
		if in.SpaceID != 0 && space.ID != in.SpaceID {
			continue
		}
		found = true
		list, err := t.Sites.List(ctx, userID, space.ID)
		if err != nil {
			return nil, sitesOutput{}, mcpkit.Failure(err, hintSpace)
		}
		all = append(all, t.matching(list, space.Name, query)...)
	}
	if !found {
		return nil, sitesOutput{}, mcpkit.Failure(errNoSpace, hintSpace)
	}
	return nil, capSites(all), nil
}

// matching renders the sites whose name or domain contains query.
func (t *toolset) matching(list []sites.SiteResponse, space, query string) []siteOut {
	out := []siteOut{}
	for i := range list {
		site := &list[i]
		if query != "" && !strings.Contains(strings.ToLower(site.Name+" "+site.Domain), query) {
			continue
		}
		rendered := t.siteOf(site)
		rendered.Space = space
		out = append(out, rendered)
	}
	return out
}

// capSites keeps as many sites as fit in one result.
func capSites(all []siteOut) sitesOutput {
	out := sitesOutput{}
	n := mcpkit.FitCount(len(all), func(n int) int { out.Sites = all[:n]; return mcpkit.EncodedSize(out) })
	out.Sites = all[:n]
	if n < len(all) {
		out.Truncated, out.Message = true, sitesCapped
	}
	return out
}

// siteOf renders a site.
func (t *toolset) siteOf(site *sites.SiteResponse) siteOut {
	return siteOut{
		ID: site.ID, Name: site.Name, Domain: site.Domain, SpaceID: site.WorkspaceID,
		URL: t.siteURL(site.ID), ShareURL: t.shareURL(site),
	}
}
