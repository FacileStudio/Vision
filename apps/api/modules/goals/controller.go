package goals

import (
	"context"
	"strings"
	"time"

	"github.com/FacileStudio/Vision/apps/api/internal/errors"
)

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) create(ctx context.Context, ownerID string, req *CreateRequest) (*GoalResponse, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.GoalType = strings.TrimSpace(strings.ToLower(req.GoalType))
	req.EventName = strings.TrimSpace(req.EventName)
	req.PagePath = strings.TrimSpace(req.PagePath)
	req.MatchType = strings.TrimSpace(strings.ToLower(req.MatchType))

	if req.Name == "" {
		return nil, errors.Invalid("name is required")
	}
	if req.SiteID == 0 {
		return nil, errors.Invalid("site_id is required")
	}
	if req.GoalType != "pageview" && req.GoalType != "event" {
		return nil, errors.Invalid("goal_type must be pageview or event")
	}
	if req.GoalType == "event" && req.EventName == "" {
		return nil, errors.Invalid("event_name is required for event goals")
	}
	if req.GoalType == "pageview" && req.PagePath == "" {
		return nil, errors.Invalid("page_path is required for pageview goals")
	}
	if req.GoalType == "pageview" && req.MatchType == "" {
		req.MatchType = "exact"
	}
	if req.GoalType == "pageview" && req.MatchType != "exact" && req.MatchType != "starts_with" && req.MatchType != "contains" {
		return nil, errors.Invalid("match_type must be exact, starts_with, or contains")
	}

	return c.service.createGoal(ctx, ownerID, req)
}

func (c *Controller) list(ctx context.Context, ownerID string, siteID string) ([]GoalResponse, error) {
	return c.service.listGoals(ctx, ownerID, siteID)
}

func (c *Controller) update(ctx context.Context, ownerID string, goalID string, req *UpdateRequest) (*GoalResponse, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.GoalType = strings.TrimSpace(strings.ToLower(req.GoalType))
	req.EventName = strings.TrimSpace(req.EventName)
	req.PagePath = strings.TrimSpace(req.PagePath)
	req.MatchType = strings.TrimSpace(strings.ToLower(req.MatchType))

	if req.Name == "" {
		return nil, errors.Invalid("name is required")
	}
	if req.GoalType != "pageview" && req.GoalType != "event" {
		return nil, errors.Invalid("goal_type must be pageview or event")
	}
	if req.GoalType == "event" && req.EventName == "" {
		return nil, errors.Invalid("event_name is required for event goals")
	}
	if req.GoalType == "pageview" && req.PagePath == "" {
		return nil, errors.Invalid("page_path is required for pageview goals")
	}
	if req.GoalType == "pageview" && req.MatchType == "" {
		req.MatchType = "exact"
	}
	if req.GoalType == "pageview" && req.MatchType != "exact" && req.MatchType != "starts_with" && req.MatchType != "contains" {
		return nil, errors.Invalid("match_type must be exact, starts_with, or contains")
	}

	return c.service.updateGoal(ctx, ownerID, goalID, req)
}

func (c *Controller) delete(ctx context.Context, ownerID string, goalID string) error {
	return c.service.deleteGoal(ctx, ownerID, goalID)
}

func (c *Controller) conversions(ctx context.Context, ownerID string, siteID string, from time.Time, to time.Time) (*ConversionsResponse, error) {
	return c.service.conversions(ctx, ownerID, siteID, from, to)
}
