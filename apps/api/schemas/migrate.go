package schemas

import "gorm.io/gorm"

// Migrate brings the schema up to date and then hands authentication to
// porte. AdoptPorte runs last because it repoints foreign keys at users(id)
// and reads the columns AutoMigrate has just guaranteed exist.
func Migrate(db *gorm.DB) error {
	return MigrateWithIssuer(db, "")
}

// MigrateWithIssuer is Migrate with the OIDC issuer, which the identity
// backfill needs: porte matches an account on (provider, subject) and the
// provider is the issuer, so backfilling with a placeholder would leave every
// existing SSO user unmatched and quietly fall through to the email path.
func MigrateWithIssuer(db *gorm.DB, issuer string) error {
	if err := db.AutoMigrate(
		&User{}, &Workspace{}, &WorkspaceMember{},
		&Site{}, &Pageview{}, &Webhook{}, &VisitorSession{},
		&CustomEvent{}, &Goal{}, &APIKey{},
	); err != nil {
		return err
	}
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
