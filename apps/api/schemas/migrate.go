package schemas

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Vision/apps/api/migrations"
	"gorm.io/gorm"
)

// Migrate applies the migrations and then the boot data steps, with no OIDC
// issuer. The tests build their schemas through it, so they run exactly the
// SQL production runs.
func Migrate(db *gorm.DB) error {
	return MigrateWithIssuer(context.Background(), db, "", slog.Default())
}

// MigrateWithIssuer applies the ordered migrations under migrations/ and then
// Reconcile. The issuer is what the identity backfill needs: porte matches an
// account on (provider, subject) and the provider is the issuer, so
// backfilling with a placeholder would leave every existing SSO user
// unmatched and quietly fall through to the email path.
func MigrateWithIssuer(ctx context.Context, db *gorm.DB, issuer string, logger *slog.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if err := migrations.Up(ctx, migrations.Config(sqlDB, logger)); err != nil {
		return err
	}
	return Reconcile(db, issuer)
}

// Reconcile runs the idempotent data steps that ran at every boot before the
// schema moved to migrations, in the same order. It changes rows, never the
// schema: the placeholder pictures, one personal workspace per user, and the
// credentials porte reads. AdoptPorte runs last because it repoints
// identities at users.
func Reconcile(db *gorm.DB, issuer string) error {
	if err := normalizeOIDCPictureURL(db); err != nil {
		return err
	}
	if err := seedPersonalWorkspaces(db); err != nil {
		return err
	}
	return AdoptPorte(db, issuer)
}

// normalizeOIDCPictureURL empties the rows that recorded a placeholder instead of a photo.
//
// The old sync stored the picture claim verbatim, and Authentik never omits it: a user with
// no photo gets `data:image/svg+xml;base64,…`, its own drawing of their initials. The column
// now means "there is a photo in Porte", so a leftover data: blob would render Authentik's
// initials where the client draws its own. The next sync would fix it anyway — this makes
// the first page load after the deploy correct too.
func normalizeOIDCPictureURL(db *gorm.DB) error {
	return db.Exec(
		`UPDATE users SET oidc_picture_url = ''
		 WHERE coalesce(oidc_picture_url, '') <> ''
		   AND oidc_picture_url NOT LIKE 'https://%'`).Error
}

func seedPersonalWorkspaces(db *gorm.DB) error {
	db.Exec(`
		UPDATE sites SET workspace_id = (
			SELECT wm.workspace_id FROM workspace_members wm
			WHERE wm.user_id = sites.owner_id
			LIMIT 1
		)
		WHERE (workspace_id IS NULL OR workspace_id = 0)
		AND owner_id IN (SELECT user_id FROM workspace_members)
	`)

	var users []User
	db.Where("id NOT IN (SELECT user_id FROM workspace_members)").Find(&users)
	if len(users) == 0 {
		return nil
	}

	for _, u := range users {
		name := u.Name
		if name == "" {
			name = u.Email
		}
		ws := &Workspace{Name: name}
		if err := db.Create(ws).Error; err != nil {
			return err
		}
		if err := db.Create(&WorkspaceMember{
			WorkspaceID: ws.ID,
			UserID:      u.ID,
			Role:        "owner",
		}).Error; err != nil {
			return err
		}
		db.Model(&Site{}).Where("owner_id = ? AND (workspace_id IS NULL OR workspace_id = 0)", u.ID).Update("workspace_id", ws.ID)
	}
	return nil
}
