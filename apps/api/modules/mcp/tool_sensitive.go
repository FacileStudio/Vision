package mcp

import (
	"context"
	"fmt"
	"strconv"

	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	statusDeleted = "deleted"
	statusShared  = "shared"
)

// deleteSite is vision_delete_site: the site goes once the user confirms.
func (t *toolset) deleteSite(ctx context.Context, req *sdk.CallToolRequest, in siteInput) (*sdk.CallToolResult, siteOutput, error) {
	return t.askSite(ctx, req, in, kindDeleteSite, "Delete the site %s (%s) and its dashboard in Vision? This cannot be undone.")
}

// shareSite is vision_share_site: a public link opens once the user confirms.
func (t *toolset) shareSite(ctx context.Context, req *sdk.CallToolRequest, in siteInput) (*sdk.CallToolResult, siteOutput, error) {
	return t.askSite(ctx, req, in, kindShareSite, "Open a public link to the analytics of %s (%s)? Anyone with the link "+
		"can see its dashboard, and an older link stops working.")
}

// askSite asks before a sensitive site action, or resumes it with the user's answer.
func (t *toolset) askSite(ctx context.Context, req *sdk.CallToolRequest, in siteInput, kind, question string) (*sdk.CallToolResult, siteOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, siteOutput{}, err
	}
	if mcpkit.Retrying(req) {
		return t.resumeSite(ctx, req, userID, kind)
	}
	site, err := t.writableSite(ctx, userID, in.SiteID)
	if err != nil {
		return nil, siteOutput{}, err
	}
	result, outcome, err := t.confirm.Request(ctx, req, kind, target(site.ID), fmt.Sprintf(question, site.Name, site.Domain))
	if err != nil {
		return nil, siteOutput{}, mcpkit.Failure(err, hintSite)
	}
	if result != nil {
		return result, siteOutput{}, nil
	}
	return nil, siteOutput{Status: outcome.Status, SiteID: site.ID, ConfirmURL: outcome.ConfirmURL, Message: outcome.Message}, nil
}

// resumeSite handles the retry that carries the user's answer about a site.
func (t *toolset) resumeSite(ctx context.Context, req *sdk.CallToolRequest, userID, kind string) (*sdk.CallToolResult, siteOutput, error) {
	outcome, err := t.confirm.Resume(ctx, req, kind)
	if err != nil {
		return nil, siteOutput{}, mcpkit.Failure(err, "Call the tool again to ask the user afresh.")
	}
	if outcome.Status == mcpkit.StatusDeclined {
		return nil, siteOutput{Status: mcpkit.StatusDeclined, Message: outcome.Message}, nil
	}
	_, siteID, _ := refOf(outcome.Action)
	id, _ := strconv.ParseInt(siteID, 10, 64)
	if kind == kindDeleteSite {
		return nil, siteOutput{Status: statusDeleted, SiteID: id, Message: "The user confirmed and the site is deleted."}, nil
	}
	site, err := t.site(ctx, userID, id)
	if err != nil {
		return nil, siteOutput{}, err
	}
	out := t.siteOf(site)
	return nil, siteOutput{Status: statusShared, Site: &out, Message: "The user confirmed. share_url is the public dashboard."}, nil
}

// deleteGoal is vision_delete_goal: the goal goes once the user confirms.
func (t *toolset) deleteGoal(ctx context.Context, req *sdk.CallToolRequest, in goalInput) (*sdk.CallToolResult, goalOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, goalOutput{}, err
	}
	if mcpkit.Retrying(req) {
		return t.resumeGoal(ctx, req)
	}
	goal, err := t.Goals.Own(ctx, userID, strconv.FormatInt(in.GoalID, 10))
	if err != nil {
		return nil, goalOutput{}, mcpkit.Failure(err, hintGoal)
	}
	question := fmt.Sprintf("Delete the goal %q in Vision? This cannot be undone.", goal.Name)
	result, outcome, err := t.confirm.Request(ctx, req, kindDeleteGoal, target(goal.ID), question)
	if err != nil {
		return nil, goalOutput{}, mcpkit.Failure(err, hintGoal)
	}
	if result != nil {
		return result, goalOutput{}, nil
	}
	return nil, goalOutput{Status: outcome.Status, GoalID: goal.ID, ConfirmURL: outcome.ConfirmURL, Message: outcome.Message}, nil
}

// resumeGoal handles the retry that carries the user's answer about a goal.
func (t *toolset) resumeGoal(ctx context.Context, req *sdk.CallToolRequest) (*sdk.CallToolResult, goalOutput, error) {
	outcome, err := t.confirm.Resume(ctx, req, kindDeleteGoal)
	if err != nil {
		return nil, goalOutput{}, mcpkit.Failure(err, "Call vision_delete_goal again to ask the user afresh.")
	}
	if outcome.Status == mcpkit.StatusDeclined {
		return nil, goalOutput{Status: mcpkit.StatusDeclined, Message: outcome.Message}, nil
	}
	_, goalID, _ := refOf(outcome.Action)
	id, _ := strconv.ParseInt(goalID, 10, 64)
	return nil, goalOutput{Status: statusDeleted, GoalID: id, Message: "The user confirmed and the goal is deleted."}, nil
}
