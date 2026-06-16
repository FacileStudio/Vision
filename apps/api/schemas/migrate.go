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
		db.Model(&Site{}).Where("owner_id = ? AND workspace_id = 0", u.ID).Update("workspace_id", ws.ID)
	}
	return nil
}
