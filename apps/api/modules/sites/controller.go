package sites

import (
	"context"
	"strings"

	"api/internal/errors"
)

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) create(ctx context.Context, ownerID string, req *CreateRequest) (*SiteResponse, error) {
	name := strings.TrimSpace(req.Name)
	domain := strings.TrimSpace(strings.ToLower(req.Domain))
	if name == "" {
		return nil, errors.Invalid("name is required")
	}
	if domain == "" {
		return nil, errors.Invalid("domain is required")
	}
	if req.WorkspaceID == 0 {
		return nil, errors.Invalid("workspace_id is required")
	}
	return c.service.createSite(ctx, ownerID, name, domain, req.WorkspaceID)
}

func (c *Controller) list(ctx context.Context, ownerID string, workspaceID int64) ([]SiteResponse, error) {
	return c.service.listSites(ctx, ownerID, workspaceID)
}

func (c *Controller) get(ctx context.Context, ownerID string, siteID string) (*SiteResponse, error) {
	return c.service.getSite(ctx, ownerID, siteID)
}

func (c *Controller) update(ctx context.Context, ownerID string, siteID string, req *UpdateRequest) (*SiteResponse, error) {
	name := strings.TrimSpace(req.Name)
	domain := strings.TrimSpace(strings.ToLower(req.Domain))
	if name == "" {
		return nil, errors.Invalid("name is required")
	}
	if domain == "" {
		return nil, errors.Invalid("domain is required")
	}
	return c.service.updateSite(ctx, ownerID, siteID, name, domain)
}

func (c *Controller) delete(ctx context.Context, ownerID string, siteID string) error {
	return c.service.deleteSite(ctx, ownerID, siteID)
}

func (c *Controller) generateShare(ctx context.Context, ownerID string, siteID string) (*SiteResponse, error) {
	return c.service.generateShareToken(ctx, ownerID, siteID)
}

func (c *Controller) revokeShare(ctx context.Context, ownerID string, siteID string) error {
	return c.service.revokeShareToken(ctx, ownerID, siteID)
}
