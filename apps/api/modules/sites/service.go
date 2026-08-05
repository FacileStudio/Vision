package sites

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	stderrors "errors"
	"strconv"

	"github.com/FacileStudio/Vision/apps/api/internal/siteaccess"
	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

type Service struct {
	orm        *gorm.DB
	controller *Controller
}

func NewService(orm *gorm.DB) *Service {
	service := &Service{orm: orm}
	service.controller = newController(service)
	return service
}

func (s *Service) createSite(ctx context.Context, ownerID string, name string, domain string, workspaceID int64) (*SiteResponse, error) {
	uid, _ := strconv.ParseInt(ownerID, 10, 64)

	var count int64
	s.orm.WithContext(ctx).Model(&schemas.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ? AND role IN ?", workspaceID, uid, []string{"owner", "admin", "editor"}).
		Count(&count)
	if count == 0 {
		return nil, errors.Forbidden("no write access to this space")
	}

	record := &schemas.Site{
		Name:        name,
		Domain:      domain,
		OwnerID:     uid,
		WorkspaceID: workspaceID,
	}
	if err := s.orm.WithContext(ctx).Create(record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.Conflict("domain already registered")
		}
		return nil, errors.Internal("failed to create site", err)
	}

	return toResponse(record), nil
}

func (s *Service) listSites(ctx context.Context, userID string, workspaceID int64) ([]SiteResponse, error) {
	uid, _ := strconv.ParseInt(userID, 10, 64)

	if workspaceID > 0 {
		var count int64
		s.orm.WithContext(ctx).Model(&schemas.WorkspaceMember{}).
			Where("workspace_id = ? AND user_id = ?", workspaceID, uid).
			Count(&count)
		if count == 0 {
			return []SiteResponse{}, nil
		}

		var records []schemas.Site
		if err := s.orm.WithContext(ctx).Where("workspace_id = ?", workspaceID).
			Order("created_at desc").Find(&records).Error; err != nil {
			return nil, errors.Internal("failed to list sites", err)
		}

		out := make([]SiteResponse, len(records))
		for i := range records {
			out[i] = *toResponse(&records[i])
		}
		return out, nil
	}

	var records []schemas.Site
	if err := s.orm.WithContext(ctx).Where("owner_id = ?", uid).
		Order("created_at desc").Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list sites", err)
	}

	out := make([]SiteResponse, len(records))
	for i := range records {
		out[i] = *toResponse(&records[i])
	}
	return out, nil
}

func (s *Service) getSite(ctx context.Context, userID string, siteID string) (*SiteResponse, error) {
	record, err := s.findAccessible(ctx, userID, siteID)
	if err != nil {
		return nil, err
	}
	return toResponse(record), nil
}

func (s *Service) updateSite(ctx context.Context, userID string, siteID string, name string, domain string) (*SiteResponse, error) {
	record, err := s.findWritable(ctx, userID, siteID)
	if err != nil {
		return nil, err
	}

	record.Name = name
	record.Domain = domain
	if err := s.orm.WithContext(ctx).Save(record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.Conflict("domain already registered")
		}
		return nil, errors.Internal("failed to update site", err)
	}
	return toResponse(record), nil
}

func (s *Service) deleteSite(ctx context.Context, userID string, siteID string) error {
	record, err := s.findWritable(ctx, userID, siteID)
	if err != nil {
		return err
	}
	if err := s.orm.WithContext(ctx).Delete(record).Error; err != nil {
		return errors.Internal("failed to delete site", err)
	}
	return nil
}

func (s *Service) findAccessible(ctx context.Context, userID string, siteID string) (*schemas.Site, error) {
	uid, _ := strconv.ParseInt(userID, 10, 64)
	sid, _ := strconv.ParseInt(siteID, 10, 64)

	var record schemas.Site
	if err := s.orm.WithContext(ctx).First(&record, sid).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("site not found")
		}
		return nil, errors.Internal("failed to read site", err)
	}

	if !siteaccess.CanAccess(ctx, s.orm, uid, sid) {
		return nil, errors.NotFound("site not found")
	}
	return &record, nil
}

func (s *Service) findWritable(ctx context.Context, userID string, siteID string) (*schemas.Site, error) {
	uid, _ := strconv.ParseInt(userID, 10, 64)
	sid, _ := strconv.ParseInt(siteID, 10, 64)

	var record schemas.Site
	if err := s.orm.WithContext(ctx).First(&record, sid).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("site not found")
		}
		return nil, errors.Internal("failed to read site", err)
	}

	if !siteaccess.CanWrite(ctx, s.orm, uid, sid) {
		return nil, errors.Forbidden("access denied")
	}
	return &record, nil
}

func (s *Service) generateShareToken(ctx context.Context, userID string, siteID string) (*SiteResponse, error) {
	record, err := s.findWritable(ctx, userID, siteID)
	if err != nil {
		return nil, err
	}

	token := generateToken()
	record.ShareToken = &token
	if err := s.orm.WithContext(ctx).Save(record).Error; err != nil {
		return nil, errors.Internal("failed to generate share token", err)
	}
	return toResponse(record), nil
}

func (s *Service) revokeShareToken(ctx context.Context, userID string, siteID string) error {
	record, err := s.findWritable(ctx, userID, siteID)
	if err != nil {
		return err
	}

	record.ShareToken = nil
	return s.orm.WithContext(ctx).Save(record).Error
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func toResponse(record *schemas.Site) *SiteResponse {
	return &SiteResponse{
		ID:          record.ID,
		Name:        record.Name,
		Domain:      record.Domain,
		OwnerID:     record.OwnerID,
		WorkspaceID: record.WorkspaceID,
		ShareToken:  record.ShareToken,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}
