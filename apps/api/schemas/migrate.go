package schemas

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&User{}, &Session{}, &Workspace{}, &WorkspaceMember{},
		&Site{}, &Pageview{}, &Webhook{}, &VisitorSession{},
		&CustomEvent{}, &Goal{}, &APIKey{},
	); err != nil {
		return err
	}
	return seedPersonalWorkspaces(db)
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
