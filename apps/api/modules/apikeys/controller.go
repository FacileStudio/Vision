package apikeys

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

func (c *Controller) create(ctx context.Context, ownerID string, req *CreateRequest) (*CreateResponse, error) {
	name := strings.TrimSpace(req.Name)
	scopes := strings.TrimSpace(strings.ToLower(req.Scopes))

	if name == "" {
		return nil, errors.Invalid("name is required")
	}
	if scopes == "" {
		scopes = "read"
	}
	if scopes != "read" && scopes != "read,write" {
		return nil, errors.Invalid("scopes must be read or read,write")
	}

	return c.service.createKey(ctx, ownerID, name, scopes, req.SiteID)
}

func (c *Controller) list(ctx context.Context, ownerID string) ([]APIKeyResponse, error) {
	return c.service.listKeys(ctx, ownerID)
}

func (c *Controller) revoke(ctx context.Context, ownerID string, keyID string) error {
	return c.service.revokeKey(ctx, ownerID, keyID)
}
