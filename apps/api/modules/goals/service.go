package goals

import (
	"context"
	stderrors "errors"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Vision/apps/api/internal/siteaccess"
	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

type Service struct {
	orm        *gorm.DB
	controller *Controller
}

func NewService(orm *gorm.DB) *Service {
	service := &Service{orm: orm}
	service.controller = newController(service)
	return service
}

func (s *Service) createGoal(ctx context.Context, ownerID string, req *CreateRequest) (*GoalResponse, error) {
	uid, _ := strconv.ParseInt(ownerID, 10, 64)

	if !siteaccess.CanWrite(ctx, s.orm, uid, req.SiteID) {
		return nil, errors.NotFound("site not found")
	}

	record := &schemas.Goal{
		SiteID:   req.SiteID,
		OwnerID:  uid,
		Name:     req.Name,
		GoalType: req.GoalType,
	}

	if req.GoalType == "event" {
		record.EventName = &req.EventName
	} else {
		record.PagePath = &req.PagePath
		record.MatchType = req.MatchType
	}

	if err := s.orm.WithContext(ctx).Create(record).Error; err != nil {
		return nil, errors.Internal("failed to create goal", err)
	}
	return toResponse(record), nil
}

func (s *Service) listGoals(ctx context.Context, ownerID string, siteID string) ([]GoalResponse, error) {
	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	sid, _ := strconv.ParseInt(siteID, 10, 64)

	if !siteaccess.CanAccess(ctx, s.orm, uid, sid) {
		return nil, errors.NotFound("site not found")
	}

	var records []schemas.Goal
	if err := s.orm.WithContext(ctx).Where("site_id = ?", sid).Order("created_at desc").Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list goals", err)
	}

	out := make([]GoalResponse, len(records))
	for i := range records {
		out[i] = *toResponse(&records[i])
	}
	return out, nil
}

func (s *Service) updateGoal(ctx context.Context, ownerID string, goalID string, req *UpdateRequest) (*GoalResponse, error) {
	record, err := s.findOwned(ctx, ownerID, goalID)
	if err != nil {
		return nil, err
	}

	record.Name = req.Name
	record.GoalType = req.GoalType
	if req.GoalType == "event" {
		record.EventName = &req.EventName
		record.PagePath = nil
	} else {
		record.PagePath = &req.PagePath
		record.MatchType = req.MatchType
		record.EventName = nil
	}

	if err := s.orm.WithContext(ctx).Save(record).Error; err != nil {
		return nil, errors.Internal("failed to update goal", err)
	}
	return toResponse(record), nil
}

func (s *Service) deleteGoal(ctx context.Context, ownerID string, goalID string) error {
	record, err := s.findOwned(ctx, ownerID, goalID)
	if err != nil {
		return err
	}
	if err := s.orm.WithContext(ctx).Delete(record).Error; err != nil {
		return errors.Internal("failed to delete goal", err)
	}
	return nil
}

func (s *Service) conversions(ctx context.Context, ownerID string, siteID string, from time.Time, to time.Time) (*ConversionsResponse, error) {
	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	sid, _ := strconv.ParseInt(siteID, 10, 64)

	if !siteaccess.CanAccess(ctx, s.orm, uid, sid) {
		return nil, errors.NotFound("site not found")
	}

	var totalVisitors int64
	s.orm.WithContext(ctx).Table("pageviews").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND visitor_id != ''", sid, from, to).
		Distinct("visitor_id").
		Count(&totalVisitors)

	var goals []schemas.Goal
	if err := s.orm.WithContext(ctx).Where("site_id = ?", sid).Find(&goals).Error; err != nil {
		return nil, errors.Internal("failed to load goals", err)
	}

	result := make([]GoalConversion, 0, len(goals))
	for _, g := range goals {
		var conversions int64

		if g.GoalType == "event" && g.EventName != nil {
			s.orm.WithContext(ctx).Table("custom_events").
				Where("site_id = ? AND created_at >= ? AND created_at <= ? AND name = ?", sid, from, to, *g.EventName).
				Distinct("visitor_id").
				Count(&conversions)
		} else if g.PagePath != nil {
			q := s.orm.WithContext(ctx).Table("pageviews").
				Where("site_id = ? AND created_at >= ? AND created_at <= ? AND visitor_id != ''", sid, from, to)

			escaped := escapeLike(*g.PagePath)
			switch g.MatchType {
			case "starts_with":
				q = q.Where("path LIKE ?", escaped+"%")
			case "contains":
				q = q.Where("path LIKE ?", "%"+escaped+"%")
			default:
				q = q.Where("path = ?", *g.PagePath)
			}

			q.Distinct("visitor_id").Count(&conversions)
		}

		var cr float64
		if totalVisitors > 0 {
			cr = float64(conversions) / float64(totalVisitors) * 100
		}

		result = append(result, GoalConversion{
			ID:             g.ID,
			Name:           g.Name,
			GoalType:       g.GoalType,
			Conversions:    conversions,
			ConversionRate: cr,
		})
	}

	return &ConversionsResponse{
		Goals:         result,
		TotalVisitors: totalVisitors,
	}, nil
}

func (s *Service) findOwned(ctx context.Context, ownerID string, goalID string) (*schemas.Goal, error) {
	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	gid, _ := strconv.ParseInt(goalID, 10, 64)

	var record schemas.Goal
	err := s.orm.WithContext(ctx).Where("id = ? AND owner_id = ?", gid, uid).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("goal not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to read goal", err)
	}
	return &record, nil
}

func toResponse(record *schemas.Goal) *GoalResponse {
	return &GoalResponse{
		ID:        record.ID,
		SiteID:    record.SiteID,
		Name:      record.Name,
		GoalType:  record.GoalType,
		EventName: record.EventName,
		PagePath:  record.PagePath,
		MatchType: record.MatchType,
		CreatedAt: record.CreatedAt,
	}
}
