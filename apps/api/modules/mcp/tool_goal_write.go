package mcp

import (
	"context"
	"strconv"

	"github.com/FacileStudio/Vision/apps/api/modules/goals"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// createGoal is vision_create_goal, through POST /goals.
func (t *toolset) createGoal(ctx context.Context, req *sdk.CallToolRequest, in createGoalInput) (*sdk.CallToolResult, goalOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, goalOutput{}, err
	}
	goal, err := t.Goals.Create(ctx, userID, &goals.CreateRequest{
		SiteID: in.SiteID, Name: in.Name, GoalType: in.GoalType, EventName: in.EventName, PagePath: in.PagePath, MatchType: in.MatchType,
	})
	if err != nil {
		return nil, goalOutput{}, mcpkit.Failure(err, hintSite)
	}
	out := goalOf(goal)
	return nil, goalOutput{Status: statusCreated, Goal: &out}, nil
}

// updateGoal is vision_update_goal, through PUT /goals/{id}: fields left out keep their
// current value, and switching the type clears the other type's field as the web app does.
func (t *toolset) updateGoal(ctx context.Context, req *sdk.CallToolRequest, in updateGoalInput) (*sdk.CallToolResult, goalOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, goalOutput{}, err
	}
	if in == (updateGoalInput{GoalID: in.GoalID}) {
		return nil, goalOutput{}, errNoChange
	}
	id := strconv.FormatInt(in.GoalID, 10)
	current, err := t.Goals.Own(ctx, userID, id)
	if err != nil {
		return nil, goalOutput{}, mcpkit.Failure(err, hintGoal)
	}
	goal, err := t.Goals.Update(ctx, userID, id, mergedGoal(goalOf(current), in))
	if err != nil {
		return nil, goalOutput{}, mcpkit.Failure(err, hintGoal)
	}
	out := goalOf(goal)
	return nil, goalOutput{Status: statusUpdated, Goal: &out}, nil
}

// mergedGoal is the PUT body: the goal as it is, with the fields given replaced.
func mergedGoal(current goalOut, in updateGoalInput) *goals.UpdateRequest {
	change := goals.UpdateRequest{
		Name: current.Name, GoalType: current.GoalType, EventName: current.EventName,
		PagePath: current.PagePath, MatchType: current.MatchType,
	}
	for field, value := range map[*string]string{
		&change.Name: in.Name, &change.GoalType: in.GoalType, &change.EventName: in.EventName,
		&change.PagePath: in.PagePath, &change.MatchType: in.MatchType,
	} {
		if value != "" {
			*field = value
		}
	}
	return &change
}
