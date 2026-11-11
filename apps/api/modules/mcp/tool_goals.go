package mcp

import (
	"context"
	"strconv"
	"time"

	"github.com/FacileStudio/Vision/apps/api/modules/goals"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// listGoals is vision_list_goals: a site's goals with their conversions in the range, as
// the dashboard's goals panel shows them.
func (t *toolset) listGoals(ctx context.Context, req *sdk.CallToolRequest, in goalsInput) (*sdk.CallToolResult, goalsOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, goalsOutput{}, err
	}
	span, err := windowOf(in.From, in.To, time.Now())
	if err != nil {
		return nil, goalsOutput{}, err
	}
	site, err := t.site(ctx, userID, in.SiteID)
	if err != nil {
		return nil, goalsOutput{}, err
	}
	siteID := strconv.FormatInt(site.ID, 10)
	list, err := t.Goals.List(ctx, userID, siteID)
	if err != nil {
		return nil, goalsOutput{}, mcpkit.Failure(err, hintSite)
	}
	conversions, err := t.Goals.Conversions(ctx, userID, siteID, span.from, span.to)
	if err != nil {
		return nil, goalsOutput{}, mcpkit.Failure(err, hintSite)
	}
	out := goalsOutput{Site: t.siteOf(site), Range: span.out(), TotalVisitors: conversions.TotalVisitors}
	out.Goals = withConversions(list, conversions.Goals)
	return nil, out, nil
}

// withConversions renders the goals in the list's order with their counts.
func withConversions(list []goals.GoalResponse, conversions []goals.GoalConversion) []goalOut {
	counted := map[int64]goals.GoalConversion{}
	for _, c := range conversions {
		counted[c.ID] = c
	}
	out := make([]goalOut, len(list))
	for i := range list {
		out[i] = goalOf(&list[i])
		out[i].Conversions, out[i].ConversionRate = counted[list[i].ID].Conversions, counted[list[i].ID].ConversionRate
	}
	return out
}

// goalOf renders a goal without its conversions.
func goalOf(g *goals.GoalResponse) goalOut {
	out := goalOut{ID: g.ID, Name: g.Name, GoalType: g.GoalType}
	if g.EventName != nil {
		out.EventName = *g.EventName
	}
	if g.PagePath != nil {
		out.PagePath, out.MatchType = *g.PagePath, g.MatchType
	}
	return out
}
