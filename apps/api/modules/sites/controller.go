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
	return c.service.createSite(ctx, ownerID, name, domain)
}

func (c *Controller) list(ctx context.Context, ownerID string) ([]SiteResponse, error) {
	return c.service.listSites(ctx, ownerID)
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

func (c *Controller) rotateKey(ctx context.Context, ownerID string, siteID string) (*SiteResponse, error) {
	return c.service.rotateAPIKey(ctx, ownerID, siteID)
}
