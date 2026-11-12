package workspaces

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/kit"

	"gorm.io/gorm"
)

// Apply makes Vision's linked workspaces match the snapshot, in one transaction, and is
// idempotent: applying the same snapshot twice writes nothing the second time.
//
// Only spaces present in the snapshot are touched. A linked workspace Registre no longer
// lists is left as it is and logged, because Registre has no space deletion yet and a
// workspace holds sites. Personal and unlinked workspaces are never touched at all.
func (d *Directory) Apply(ctx context.Context, snapshot directory.Snapshot) error {
	return d.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		listed := []string{""}
		for _, space := range snapshot.Spaces {
			if err := d.applySpace(tx, space); err != nil {
				return fmt.Errorf("directory: apply space %s: %w", space.ID, err)
			}
			listed = append(listed, space.ID)
		}
		var orphans []string
		if err := tx.Model(&schemas.Workspace{}).
			Where("registre_space_id IS NOT NULL AND registre_space_id NOT IN ?", listed).
			Pluck("registre_space_id", &orphans).Error; err != nil {
			return err
		}
		if len(orphans) > 0 {
			d.logger.Info("directory: linked workspaces missing from the snapshot were kept",
				slog.Any("registre_space_ids", orphans))
		}
		return nil
	})
}

// applySpace links one Registre space to a workspace, then sets its members.
func (d *Directory) applySpace(tx *gorm.DB, space directory.Space) error {
	workspace, err := linkWorkspace(tx, space)
	if err != nil {
		return err
	}
	kept := []int64{0}
	for _, member := range space.Members {
		userID, err := d.resolveMember(tx, member)
		if err != nil {
			return err
		}
		if userID == 0 {
			continue
		}
		if err := setMembership(tx, workspace.ID, userID, visionRole(member.Role)); err != nil {
			return err
		}
		kept = append(kept, userID)
	}
	return d.removeDeparted(tx, workspace.ID, kept)
}

// linkWorkspace finds the workspace a Registre space mirrors: the one already linked
// to it, else an unlinked shared workspace with the same name ignoring case, else a new
// one. A single-member workspace is personal and never linked by name, so a Registre
// space that happens to share somebody's name cannot open their sites to a team.
func linkWorkspace(tx *gorm.DB, space directory.Space) (*schemas.Workspace, error) {
	var found []schemas.Workspace
	if err := tx.Where("registre_space_id = ?", space.ID).Limit(1).Find(&found).Error; err != nil {
		return nil, err
	}
	if len(found) == 0 {
		if err := tx.Where(`registre_space_id IS NULL AND lower(name) = lower(?)
			AND (SELECT count(*) FROM workspace_members m WHERE m.workspace_id = workspaces.id) > 1`, space.Name).
			Order("id").Limit(1).Find(&found).Error; err != nil {
			return nil, err
		}
	}
	workspace := schemas.Workspace{}
	if len(found) == 1 {
		workspace = found[0]
	}
	if workspace.ID != 0 && workspace.Name == space.Name &&
		workspace.RegistreSpaceID != nil && *workspace.RegistreSpaceID == space.ID {
		return &workspace, nil
	}
	id := space.ID
	workspace.Name = space.Name
	workspace.RegistreSpaceID = &id
	return &workspace, tx.Save(&workspace).Error
}

// resolveMember returns the Vision user a Registre member is, creating one if needed.
//
// The porte identity under this issuer is the key. Failing that, a user with the same
// email is linked to it. Failing that, the user is pre-provisioned with the identity
// row already in place, so their first sign-in lands on the account that already holds
// their workspaces. A member with neither subject nor email, or one Vision already
// erased, resolves to 0 and is skipped.
func (d *Directory) resolveMember(tx *gorm.DB, member directory.Member) (int64, error) {
	if erased, err := d.isErased(tx.Statement.Context, member.Subject); err != nil || erased {
		return 0, err
	}
	var userIDs []int64
	if err := tx.Raw(`SELECT user_id FROM porte_identities WHERE provider = ? AND subject = ?`,
		d.issuer, member.Subject).Scan(&userIDs).Error; err != nil || len(userIDs) > 0 {
		return firstOrZero(userIDs), err
	}
	email := strings.ToLower(strings.TrimSpace(member.Email))
	if member.Subject == "" || email == "" {
		return 0, nil
	}
	var users []schemas.User
	if err := tx.Where("lower(email) = ?", email).Limit(1).Find(&users).Error; err != nil {
		return 0, err
	}
	user := schemas.User{Email: email, Name: member.Name, OIDCPictureURL: kit.PictureURL(member.Picture)}
	if len(users) == 1 {
		user = users[0]
	} else if err := tx.Create(&user).Error; err != nil {
		return 0, err
	}
	return user.ID, tx.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)
		ON CONFLICT (provider, subject) DO NOTHING`, user.ID, d.issuer, member.Subject).Error
}

// isErased asks the deletion receiver whether subject was erased under this issuer.
func (d *Directory) isErased(ctx context.Context, subject string) (bool, error) {
	if d.erased == nil || subject == "" {
		return false, nil
	}
	return d.erased.Erased(ctx, d.issuer, subject)
}

// firstOrZero is the first id, or 0 for none.
func firstOrZero(ids []int64) int64 {
	if len(ids) == 0 {
		return 0
	}
	return ids[0]
}

// setMembership puts the user in the workspace at role, writing only when something
// differs.
func setMembership(tx *gorm.DB, workspaceID, userID int64, role string) error {
	return tx.Exec(`INSERT INTO workspace_members (workspace_id, user_id, role, created_at) VALUES (?, ?, ?, now())
		ON CONFLICT (workspace_id, user_id) DO UPDATE SET role = EXCLUDED.role
		WHERE workspace_members.role <> EXCLUDED.role`, workspaceID, userID, role).Error
}

// removeDeparted drops the members Registre no longer lists in this workspace.
//
// Only members with a Registre identity can depart: a local password account was never
// Registre's to remove. Viewers are spared too, because viewer is a Vision-only rank
// Registre never sees. kept always carries 0, so NOT IN never meets an empty list.
func (d *Directory) removeDeparted(tx *gorm.DB, workspaceID int64, kept []int64) error {
	return tx.Exec(`DELETE FROM workspace_members wm
		 WHERE wm.workspace_id = ? AND wm.role <> 'viewer' AND wm.user_id NOT IN ?
		   AND EXISTS (SELECT 1 FROM porte_identities i WHERE i.user_id = wm.user_id AND i.provider = ?)`,
		workspaceID, kept, d.issuer).Error
}
