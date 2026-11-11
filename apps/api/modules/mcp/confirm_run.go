package mcp

import (
	"context"

	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// runDeleteSite deletes the targeted site.
func (t *toolset) runDeleteSite(ctx context.Context, a mcpkit.Action) error {
	userID, siteID, err := refOf(a)
	if err != nil {
		return err
	}
	return t.Sites.Delete(ctx, userID, siteID)
}

// runShareSite opens a fresh public link to the targeted site.
func (t *toolset) runShareSite(ctx context.Context, a mcpkit.Action) error {
	userID, siteID, err := refOf(a)
	if err != nil {
		return err
	}
	_, err = t.Sites.Share(ctx, userID, siteID)
	return err
}

// runDeleteGoal deletes the targeted goal.
func (t *toolset) runDeleteGoal(ctx context.Context, a mcpkit.Action) error {
	userID, goalID, err := refOf(a)
	if err != nil {
		return err
	}
	return t.Goals.Delete(ctx, userID, goalID)
}
