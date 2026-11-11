package workspaces

import (
	"cmp"
	"context"
	stderrors "errors"
	"strconv"
	"strings"

	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// Service owns workspace and membership persistence.
//
// registre is the Registre issuer when Vision signs in through one, and empty
// otherwise. With it set, shared workspaces are created and managed in Registre, and
// every write to a linked workspace is refused with a pointer there.
type Service struct {
	orm        *gorm.DB
	registre   string
	controller *Controller
}

// NewService returns a workspaces Service backed by the given connection. registre is
// the OIDC issuer, or empty when Vision runs without SSO.
func NewService(orm *gorm.DB, registre string) *Service {
	s := &Service{orm: orm, registre: strings.TrimSuffix(registre, "/")}
	s.controller = newController(s)
	return s
}

func (s *Service) countMembers(ctx context.Context, wsID int64) int64 {
	var c int64
	s.orm.WithContext(ctx).Model(&schemas.WorkspaceMember{}).Where("workspace_id = ?", wsID).Count(&c)
	return c
}

func (s *Service) countSites(ctx context.Context, wsID int64) int64 {
	var c int64
	s.orm.WithContext(ctx).Model(&schemas.Site{}).Where("workspace_id = ?", wsID).Count(&c)
	return c
}

func (s *Service) createWorkspace(ctx context.Context, userID string, name string) (*WorkspaceResponse, error) {
	if s.registre != "" {
		return nil, errors.Conflict("spaces are created in Registre: " + s.registre + "/spaces")
	}
	uid, _ := strconv.ParseInt(userID, 10, 64)

	ws := &schemas.Workspace{Name: name}
	if err := s.orm.WithContext(ctx).Create(ws).Error; err != nil {
		return nil, errors.Internal("failed to create space", err)
	}

	member := &schemas.WorkspaceMember{WorkspaceID: ws.ID, UserID: uid, Role: "owner"}
	if err := s.orm.WithContext(ctx).Create(member).Error; err != nil {
		return nil, errors.Internal("failed to add owner", err)
	}

	return &WorkspaceResponse{
		ID: ws.ID, Name: ws.Name, Role: "owner",
		MemberCount: 1, SiteCount: 0,
		CreatedAt: ws.CreatedAt, UpdatedAt: ws.UpdatedAt,
	}, nil
}

// listWorkspaces lists the caller's workspaces, giving them a personal one first when
// they belong to none. Spaces are created in Registre once it manages them, so this is
// what keeps a newcomer outside every shared space able to add a site.
func (s *Service) listWorkspaces(ctx context.Context, userID string) ([]WorkspaceResponse, error) {
	uid, _ := strconv.ParseInt(userID, 10, 64)
	if err := s.ensurePersonal(ctx, uid); err != nil {
		return nil, errors.Internal("failed to create a personal space", err)
	}
	var members []schemas.WorkspaceMember
	if err := s.orm.WithContext(ctx).Preload("Workspace").
		Where("user_id = ?", uid).Find(&members).Error; err != nil {
		return nil, errors.Internal("failed to list spaces", err)
	}
	out := make([]WorkspaceResponse, len(members))
	for i, m := range members {
		out[i] = s.response(ctx, m.Workspace, m.Role)
	}
	return out, nil
}

// ensurePersonal creates a workspace named after the user, as the boot migration names
// one, when they belong to none. The advisory lock makes two first page loads racing
// each other create one workspace, not two.
func (s *Service) ensurePersonal(ctx context.Context, uid int64) error {
	return s.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user schemas.User
		err := tx.Exec(`SELECT pg_advisory_xact_lock(?)`, uid).Error
		if err == nil {
			err = tx.Where("id = ? AND id NOT IN (SELECT user_id FROM workspace_members)", uid).Limit(1).Find(&user).Error
		}
		if err != nil || user.ID == 0 {
			return err
		}
		ws := schemas.Workspace{Name: cmp.Or(user.Name, user.Email)}
		if err := tx.Create(&ws).Error; err != nil {
			return err
		}
		return tx.Create(&schemas.WorkspaceMember{WorkspaceID: ws.ID, UserID: uid, Role: "owner"}).Error
	})
}

func (s *Service) getWorkspace(ctx context.Context, userID string, wsID string) (*WorkspaceResponse, error) {
	member, err := s.findMembership(ctx, userID, wsID)
	if err != nil {
		return nil, err
	}

	resp := s.response(ctx, member.Workspace, member.Role)
	return &resp, nil
}

func (s *Service) updateWorkspace(ctx context.Context, userID string, wsID string, name string) (*WorkspaceResponse, error) {
	member, err := s.requireRole(ctx, userID, wsID, s.refuseLinked, "owner", "admin")
	if err != nil {
		return nil, err
	}

	ws := member.Workspace
	ws.Name = name
	if err := s.orm.WithContext(ctx).Save(&ws).Error; err != nil {
		return nil, errors.Internal("failed to update space", err)
	}

	resp := s.response(ctx, ws, member.Role)
	return &resp, nil
}

func (s *Service) deleteWorkspace(ctx context.Context, userID string, wsID string) error {
	_, err := s.requireRole(ctx, userID, wsID, s.refuseLinked, "owner")
	if err != nil {
		return err
	}
	wid, _ := strconv.ParseInt(wsID, 10, 64)

	var siteCount int64
	s.orm.WithContext(ctx).Model(&schemas.Site{}).Where("workspace_id = ?", wid).Count(&siteCount)
	if siteCount > 0 {
		return errors.Invalid("space still has sites, remove them first")
	}

	s.orm.WithContext(ctx).Where("workspace_id = ?", wid).Delete(&schemas.WorkspaceMember{})
	return s.orm.WithContext(ctx).Delete(&schemas.Workspace{}, wid).Error
}

func (s *Service) listMembers(ctx context.Context, userID string, wsID string) ([]MemberResponse, error) {
	if _, err := s.findMembership(ctx, userID, wsID); err != nil {
		return nil, err
	}
	wid, _ := strconv.ParseInt(wsID, 10, 64)

	var members []schemas.WorkspaceMember
	if err := s.orm.WithContext(ctx).Preload("User").
		Where("workspace_id = ?", wid).Order("created_at asc").
		Find(&members).Error; err != nil {
		return nil, errors.Internal("failed to list members", err)
	}

	out := make([]MemberResponse, len(members))
	for i, m := range members {
		out[i] = MemberResponse{
			ID: m.ID, UserID: m.UserID,
			Email: m.User.Email, Name: m.User.Name, AvatarURL: m.User.Avatar(),
			Role: m.Role, CreatedAt: m.CreatedAt,
		}
	}
	return out, nil
}

func (s *Service) addMember(ctx context.Context, userID string, wsID string, email string, role string) (*MemberResponse, error) {
	if _, err := s.requireRole(ctx, userID, wsID, s.refuseMembers, "owner", "admin"); err != nil {
		return nil, err
	}
	wid, _ := strconv.ParseInt(wsID, 10, 64)

	var user schemas.User
	if err := s.orm.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("no user found with that email")
		}
		return nil, errors.Internal("failed to find user", err)
	}

	var existing schemas.WorkspaceMember
	if s.orm.WithContext(ctx).Where("workspace_id = ? AND user_id = ?", wid, user.ID).First(&existing).Error == nil {
		return nil, errors.Conflict("user is already a member")
	}

	member := &schemas.WorkspaceMember{WorkspaceID: wid, UserID: user.ID, Role: role}
	if err := s.orm.WithContext(ctx).Create(member).Error; err != nil {
		return nil, errors.Internal("failed to add member", err)
	}

	return &MemberResponse{
		ID: member.ID, UserID: user.ID,
		Email: user.Email, Name: user.Name, AvatarURL: user.Avatar(),
		Role: role, CreatedAt: member.CreatedAt,
	}, nil
}

func (s *Service) updateMemberRole(ctx context.Context, userID string, wsID string, targetUserID string, role string) error {
	if _, err := s.requireRole(ctx, userID, wsID, s.refuseMembers, "owner", "admin"); err != nil {
		return err
	}
	wid, _ := strconv.ParseInt(wsID, 10, 64)
	tuid, _ := strconv.ParseInt(targetUserID, 10, 64)

	var member schemas.WorkspaceMember
	if err := s.orm.WithContext(ctx).Where("workspace_id = ? AND user_id = ?", wid, tuid).First(&member).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("member not found")
		}
		return errors.Internal("failed to find member", err)
	}
	if member.Role == "owner" {
		return errors.Forbidden("cannot change the owner's role")
	}

	member.Role = role
	return s.orm.WithContext(ctx).Save(&member).Error
}

func (s *Service) removeMember(ctx context.Context, userID string, wsID string, targetUserID string) error {
	if _, err := s.requireRole(ctx, userID, wsID, s.refuseMembers, "owner", "admin"); err != nil {
		return err
	}
	wid, _ := strconv.ParseInt(wsID, 10, 64)
	tuid, _ := strconv.ParseInt(targetUserID, 10, 64)

	var member schemas.WorkspaceMember
	if err := s.orm.WithContext(ctx).Where("workspace_id = ? AND user_id = ?", wid, tuid).First(&member).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("member not found")
		}
		return errors.Internal("failed to find member", err)
	}
	if member.Role == "owner" {
		return errors.Forbidden("cannot remove the owner")
	}
	return s.orm.WithContext(ctx).Delete(&member).Error
}

func (s *Service) findMembership(ctx context.Context, userID string, wsID string) (*schemas.WorkspaceMember, error) {
	uid, _ := strconv.ParseInt(userID, 10, 64)
	wid, _ := strconv.ParseInt(wsID, 10, 64)

	var member schemas.WorkspaceMember
	err := s.orm.WithContext(ctx).Preload("Workspace").
		Where("workspace_id = ? AND user_id = ?", wid, uid).First(&member).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Forbidden("access denied")
	}
	if err != nil {
		return nil, errors.Internal("failed to check membership", err)
	}
	return &member, nil
}

// requireRole is the gate on every write: the caller must hold one of roles, and the
// write must be Vision's to make. refuse names the rule that decides the second half.
func (s *Service) requireRole(
	ctx context.Context, userID, wsID string, refuse func(schemas.Workspace) error, roles ...string,
) (*schemas.WorkspaceMember, error) {
	member, err := s.findMembership(ctx, userID, wsID)
	if err != nil {
		return nil, err
	}
	if err := refuse(member.Workspace); err != nil {
		return nil, err
	}
	for _, r := range roles {
		if member.Role == r {
			return member, nil
		}
	}
	return nil, errors.Forbidden("insufficient permissions")
}

func (s *Service) leaveWorkspace(ctx context.Context, userID string, wsID string) error {
	member, err := s.findMembership(ctx, userID, wsID)
	if err != nil {
		return err
	}
	if err := s.refuseMembers(member.Workspace); err != nil {
		return err
	}
	if member.Role == "owner" {
		return errors.Forbidden("owner cannot leave, transfer ownership or delete the space")
	}
	return s.orm.WithContext(ctx).Delete(member).Error
}

// List is GET /workspaces: the workspaces userID belongs to, with a personal one made for
// a person who has none, as the web app does on first load.
func (s *Service) List(ctx context.Context, userID string) ([]WorkspaceResponse, error) {
	return s.controller.list(ctx, userID)
}
