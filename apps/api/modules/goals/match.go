package goals

import (
	"time"

	"github.com/FacileStudio/Vision/apps/api/schemas"

	"gorm.io/gorm"
)

// MatchQuery selects the rows that count toward goal on one site between from and to,
// both included: custom events with the goal's name, or pageviews of a visitor whose
// path matches. It returns nil for a goal with no target. Callers count distinct
// visitor_id over it.
func MatchQuery(orm *gorm.DB, goal schemas.Goal, siteID int64, from time.Time, to time.Time) *gorm.DB {
	if goal.GoalType == "event" && goal.EventName != nil {
		return orm.Table("custom_events").
			Where("site_id = ? AND created_at >= ? AND created_at <= ? AND name = ?", siteID, from, to, *goal.EventName)
	}
	if goal.PagePath == nil {
		return nil
	}
	query := orm.Table("pageviews").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND visitor_id != ''", siteID, from, to)
	escaped := escapeLike(*goal.PagePath)
	switch goal.MatchType {
	case "starts_with":
		return query.Where("path LIKE ?", escaped+"%")
	case "contains":
		return query.Where("path LIKE ?", "%"+escaped+"%")
	default:
		return query.Where("path = ?", *goal.PagePath)
	}
}
