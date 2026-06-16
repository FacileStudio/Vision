package siteaccess

import (
	"context"

	"gorm.io/gorm"
)

func CanAccess(ctx context.Context, db *gorm.DB, userID int64, siteID int64) bool {
	var count int64
	db.WithContext(ctx).
		Table("workspace_members").
		Joins("JOIN sites ON sites.workspace_id = workspace_members.workspace_id").
		Where("sites.id = ? AND workspace_members.user_id = ?", siteID, userID).
		Count(&count)
	return count > 0
}

func CanWrite(ctx context.Context, db *gorm.DB, userID int64, siteID int64) bool {
	var count int64
	db.WithContext(ctx).
		Table("workspace_members").
		Joins("JOIN sites ON sites.workspace_id = workspace_members.workspace_id").
		Where("sites.id = ? AND workspace_members.user_id = ? AND workspace_members.role IN ?",
			siteID, userID, []string{"owner", "admin", "editor"}).
		Count(&count)
	return count > 0
}
