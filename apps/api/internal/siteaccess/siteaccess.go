package siteaccess

import (
	"context"

	"gorm.io/gorm"
)

// CanAccess reports whether the user belongs to the workspace that owns the
// given site, regardless of role.
func CanAccess(ctx context.Context, db *gorm.DB, userID int64, siteID int64) bool {
	var count int64
	db.WithContext(ctx).
		Table("workspace_members").
		Joins("JOIN sites ON sites.workspace_id = workspace_members.workspace_id").
		Where("sites.id = ? AND workspace_members.user_id = ?", siteID, userID).
		Count(&count)
	return count > 0
}

// CanWrite reports whether the user's role in the site's workspace is owner,
// admin or editor, i.e. able to modify the site.
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
