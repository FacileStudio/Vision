package workspaces

import (
	"context"
	"strings"

	"github.com/FacileStudio/tronc/errors"
)

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

var validRoles = map[string]bool{"owner": true, "admin": true, "editor": true, "viewer": true}

func assignableRole(role string) bool { return role != "owner" && validRoles[role] }

func (c *Controller) create(ctx context.Context, userID string, req *CreateRequest) (*WorkspaceResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.Invalid("name is required")
	}
	return c.service.createWorkspace(ctx, userID, name)
}

func (c *Controller) list(ctx context.Context, userID string) ([]WorkspaceResponse, error) {
	return c.service.listWorkspaces(ctx, userID)
}

func (c *Controller) get(ctx context.Context, userID string, wsID string) (*WorkspaceResponse, error) {
	return c.service.getWorkspace(ctx, userID, wsID)
}

func (c *Controller) update(ctx context.Context, userID string, wsID string, req *UpdateRequest) (*WorkspaceResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.Invalid("name is required")
	}
	return c.service.updateWorkspace(ctx, userID, wsID, name)
}

func (c *Controller) delete(ctx context.Context, userID string, wsID string) error {
	return c.service.deleteWorkspace(ctx, userID, wsID)
}

func (c *Controller) listMembers(ctx context.Context, userID string, wsID string) ([]MemberResponse, error) {
	return c.service.listMembers(ctx, userID, wsID)
}

func (c *Controller) addMember(ctx context.Context, userID string, wsID string, req *AddMemberRequest) (*MemberResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	role := strings.TrimSpace(strings.ToLower(req.Role))
	if email == "" {
		return nil, errors.Invalid("email is required")
	}
	if !assignableRole(role) {
		return nil, errors.Invalid("role must be admin, editor, or viewer")
	}
	return c.service.addMember(ctx, userID, wsID, email, role)
}

func (c *Controller) updateMember(ctx context.Context, userID string, wsID string, targetUserID string, req *UpdateMemberRequest) error {
	role := strings.TrimSpace(strings.ToLower(req.Role))
	if !assignableRole(role) {
		return errors.Invalid("role must be admin, editor, or viewer")
	}
	return c.service.updateMemberRole(ctx, userID, wsID, targetUserID, role)
}

func (c *Controller) removeMember(ctx context.Context, userID string, wsID string, targetUserID string) error {
	return c.service.removeMember(ctx, userID, wsID, targetUserID)
}

func (c *Controller) leave(ctx context.Context, userID string, wsID string) error {
	return c.service.leaveWorkspace(ctx, userID, wsID)
}
