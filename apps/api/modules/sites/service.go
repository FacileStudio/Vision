package sites

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	stderrors "errors"
	"strconv"

	"api/internal/errors"
	"api/schemas"

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

func (s *Service) createSite(ctx context.Context, ownerID string, name string, domain string) (*SiteResponse, error) {
	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, errors.Internal("failed to generate API key", err)
	}

	record := &schemas.Site{
		Name:    name,
		Domain:  domain,
		APIKey:  apiKey,
		OwnerID: uid,
	}
	if err := s.orm.WithContext(ctx).Create(record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.Conflict("domain already registered")
		}
		return nil, errors.Internal("failed to create site", err)
	}

	return toResponse(record), nil
}

func (s *Service) listSites(ctx context.Context, ownerID string) ([]SiteResponse, error) {
	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	var records []schemas.Site
	if err := s.orm.WithContext(ctx).Where("owner_id = ?", uid).Order("created_at desc").Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list sites", err)
	}

	out := make([]SiteResponse, len(records))
	for i := range records {
		out[i] = *toResponse(&records[i])
	}
	return out, nil
}

func (s *Service) getSite(ctx context.Context, ownerID string, siteID string) (*SiteResponse, error) {
	record, err := s.findOwned(ctx, ownerID, siteID)
	if err != nil {
		return nil, err
	}
	return toResponse(record), nil
}

func (s *Service) updateSite(ctx context.Context, ownerID string, siteID string, name string, domain string) (*SiteResponse, error) {
	record, err := s.findOwned(ctx, ownerID, siteID)
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

func (s *Service) deleteSite(ctx context.Context, ownerID string, siteID string) error {
	record, err := s.findOwned(ctx, ownerID, siteID)
	if err != nil {
		return err
	}
	if err := s.orm.WithContext(ctx).Delete(record).Error; err != nil {
		return errors.Internal("failed to delete site", err)
	}
	return nil
}

func (s *Service) rotateAPIKey(ctx context.Context, ownerID string, siteID string) (*SiteResponse, error) {
	record, err := s.findOwned(ctx, ownerID, siteID)
	if err != nil {
		return nil, err
	}

	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, errors.Internal("failed to generate API key", err)
	}
	record.APIKey = apiKey
	if err := s.orm.WithContext(ctx).Save(record).Error; err != nil {
		return nil, errors.Internal("failed to rotate API key", err)
	}
	return toResponse(record), nil
}

func (s *Service) findOwned(ctx context.Context, ownerID string, siteID string) (*schemas.Site, error) {
	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	sid, _ := strconv.ParseInt(siteID, 10, 64)

	var record schemas.Site
	err := s.orm.WithContext(ctx).Where("id = ? AND owner_id = ?", sid, uid).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("site not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to read site", err)
	}
	return &record, nil
}

func toResponse(record *schemas.Site) *SiteResponse {
	return &SiteResponse{
		ID:        record.ID,
		Name:      record.Name,
		Domain:    record.Domain,
		APIKey:    record.APIKey,
		OwnerID:   record.OwnerID,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

func generateAPIKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "vs_" + hex.EncodeToString(b), nil
}
