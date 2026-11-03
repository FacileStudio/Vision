package workspaces

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// Directory keeps Vision's shared workspaces equal to the spaces Registre owns.
//
// It imports the workspaces Vision already shared once, then applies Registre's
// snapshot at boot, every five minutes and shortly after every nudge. The import has to
// land before the first sync: a snapshot that predates it does not yet carry Vision's
// members, and applying it would remove them.
type Directory struct {
	orm        *gorm.DB
	issuer     string
	client     directory.Client
	syncer     *directory.Syncer
	logger     *slog.Logger
	minBackoff time.Duration
	maxBackoff time.Duration
}

// NewDirectory returns the sync for the Registre at issuer, authenticated with Vision's
// own OIDC client credentials.
func NewDirectory(orm *gorm.DB, issuer, clientID, clientSecret string, logger *slog.Logger) *Directory {
	d := &Directory{
		orm:        orm,
		issuer:     issuer,
		client:     directory.Client{Issuer: issuer, ClientID: clientID, ClientSecret: clientSecret},
		logger:     logger,
		minBackoff: 10 * time.Second,
		maxBackoff: 5 * time.Minute,
	}
	d.syncer = &directory.Syncer{Client: d.client, Apply: d.Apply, Logger: logger}
	return d
}

// Handler answers Registre's nudge. It takes no credentials because a nudge only makes
// the next authenticated pull happen sooner.
func (d *Directory) Handler() http.Handler { return d.syncer.Handler() }

// Run imports, retrying with a capped backoff while Registre cannot take it, then keeps
// syncing until ctx ends. A Registre without the directory API (404) is retried the
// same way, so Vision keeps serving its own workspaces until it is deployed.
func (d *Directory) Run(ctx context.Context) {
	if !d.importOnce(ctx) {
		return
	}
	d.syncer.Run(ctx)
}

// importOnce sends the import until Registre accepts it, reporting false only when ctx
// ended first. Each failure is logged once per attempt, and attempts are at most
// maxBackoff apart, so an absent Registre costs one warning every few minutes.
func (d *Directory) importOnce(ctx context.Context) bool {
	wait := d.minBackoff
	for {
		spaces, err := d.sharedWorkspaces(ctx)
		if err == nil {
			_, err = d.client.Import(ctx, spaces)
		}
		if err == nil {
			d.logger.Info("directory: imported shared workspaces into registre", slog.Int("spaces", len(spaces)))
			return true
		}
		d.logger.Warn("directory: import failed, retrying", slog.Any("error", err), slog.String("retry_in", wait.String()))
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return false
		case <-timer.C:
		}
		wait = min(wait*2, d.maxBackoff)
	}
}

// importRow is one member of one shared workspace, as the import query reads it.
type importRow struct {
	WorkspaceID int64
	Name        string
	Role        string
	Email       string
	Subject     string
}

// sharedWorkspaces lists every unlinked workspace with more than one member, with its
// members at their Registre rank.
//
// A single-member workspace is personal and never leaves Vision. Viewers are left out
// on purpose: Registre has no read-only rank, so sending one as a member would make
// them an editor here and a full member in every other app. They stay Vision's own.
// Subject is sent when the member has signed in through this issuer, the email
// otherwise.
func (d *Directory) sharedWorkspaces(ctx context.Context) ([]directory.ImportSpace, error) {
	var rows []importRow
	err := d.orm.WithContext(ctx).Raw(`
		SELECT w.id AS workspace_id, w.name, wm.role, u.email, coalesce(i.subject, '') AS subject
		  FROM workspaces w
		  JOIN workspace_members wm ON wm.workspace_id = w.id
		  JOIN users u ON u.id = wm.user_id
		  LEFT JOIN porte_identities i ON i.user_id = u.id AND i.provider = ?
		 WHERE w.registre_space_id IS NULL
		   AND wm.role <> 'viewer'
		   AND (SELECT count(*) FROM workspace_members c WHERE c.workspace_id = w.id) > 1
		 ORDER BY w.id, wm.id`, d.issuer).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	spaces := []directory.ImportSpace{}
	index := map[int64]int{}
	for _, row := range rows {
		at, seen := index[row.WorkspaceID]
		if !seen {
			at = len(spaces)
			index[row.WorkspaceID] = at
			spaces = append(spaces, directory.ImportSpace{Name: row.Name})
		}
		member := directory.ImportMember{Subject: row.Subject, Role: registreRole(row.Role)}
		if row.Subject == "" {
			member.Email = row.Email
		}
		spaces[at].Members = append(spaces[at].Members, member)
	}
	return spaces, nil
}

// registreRole is a Vision role at its Registre rank. Editor is Vision's ordinary
// member; viewer never reaches here.
func registreRole(role string) string {
	switch role {
	case "owner":
		return directory.RoleOwner
	case "admin":
		return directory.RoleAdmin
	default:
		return directory.RoleMember
	}
}

// visionRole is a Registre rank in Vision's vocabulary. A member is an editor, the
// ordinary member who can add and change sites.
func visionRole(role string) string {
	switch role {
	case directory.RoleOwner:
		return "owner"
	case directory.RoleAdmin:
		return "admin"
	default:
		return "editor"
	}
}

// manageURL is where a workspace's name and members are managed, or empty when Vision
// manages it itself.
func (s *Service) manageURL(ws schemas.Workspace) string {
	if ws.RegistreSpaceID == nil || s.registre == "" {
		return ""
	}
	return s.registre + "/spaces/" + url.PathEscape(*ws.RegistreSpaceID)
}

// response is a workspace as the caller with role sees it.
func (s *Service) response(ctx context.Context, ws schemas.Workspace, role string) WorkspaceResponse {
	return WorkspaceResponse{
		ID: ws.ID, Name: ws.Name, Role: role,
		MemberCount: s.countMembers(ctx, ws.ID),
		SiteCount:   s.countSites(ctx, ws.ID),
		ManageURL:   s.manageURL(ws),
		CreatedAt:   ws.CreatedAt, UpdatedAt: ws.UpdatedAt,
	}
}

// refuseLinked stops a rename or a deletion of a workspace Registre owns, naming where
// to make it.
func (s *Service) refuseLinked(ws schemas.Workspace) error {
	if manage := s.manageURL(ws); manage != "" {
		return errors.Conflict("this space is managed in Registre: " + manage)
	}
	return nil
}

// refuseMembers stops every membership write once Registre manages spaces, personal
// workspaces included: inviting somebody into one would make a shared space Registre
// never hears of.
func (s *Service) refuseMembers(ws schemas.Workspace) error {
	if s.registre == "" {
		return nil
	}
	manage := s.manageURL(ws)
	if manage == "" {
		manage = s.registre + "/spaces"
	}
	return errors.Conflict("members are managed in Registre: " + manage)
}
