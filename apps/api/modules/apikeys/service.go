package apikeys

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	stderrors "errors"
	"fmt"
	"strconv"
	"time"

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

func (s *Service) createKey(ctx context.Context, ownerID string, name string, scopes string, siteID *int64) (*CreateResponse, error) {
	uid, _ := strconv.ParseInt(ownerID, 10, 64)

	if siteID != nil {
		var siteCount int64
		s.orm.WithContext(ctx).Table("sites").Where("id = ? AND owner_id = ?", *siteID, uid).Count(&siteCount)
		if siteCount == 0 {
			return nil, errors.NotFound("site not found")
		}
	}

	var keyCount int64
	s.orm.WithContext(ctx).Table("api_keys").Where("user_id = ? AND is_active = ?", uid, true).Count(&keyCount)
	if keyCount >= 25 {
		return nil, errors.Invalid("maximum of 25 active API keys reached")
	}

	rawKey, err := generateKey()
	if err != nil {
		return nil, errors.Internal("failed to generate key", err)
	}

	prefix := "vis_ro"
	if scopes == "read,write" {
		prefix = "vis_rw"
	}
	fullKey := prefix + "_" + rawKey
	hint := fullKey[len(fullKey)-4:]
	hash := hashKey(fullKey)

	record := &schemas.APIKey{
		UserID:   uid,
		SiteID:   siteID,
		Name:     name,
		Prefix:   prefix,
		KeyHint:  hint,
		KeyHash:  hash,
		Scopes:   scopes,
		IsActive: true,
	}

	if err := s.orm.WithContext(ctx).Create(record).Error; err != nil {
		return nil, errors.Internal("failed to create api key", err)
	}

	return &CreateResponse{
		Key: fullKey,
		API: *toResponse(record),
	}, nil
}

func (s *Service) listKeys(ctx context.Context, ownerID string) ([]APIKeyResponse, error) {
	uid, _ := strconv.ParseInt(ownerID, 10, 64)

	var records []schemas.APIKey
	if err := s.orm.WithContext(ctx).Where("user_id = ?", uid).Order("created_at desc").Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list api keys", err)
	}

	out := make([]APIKeyResponse, len(records))
	for i := range records {
		out[i] = *toResponse(&records[i])
	}
	return out, nil
}

func (s *Service) revokeKey(ctx context.Context, ownerID string, keyID string) error {
	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	kid, _ := strconv.ParseInt(keyID, 10, 64)

	var record schemas.APIKey
	err := s.orm.WithContext(ctx).Where("id = ? AND user_id = ?", kid, uid).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NotFound("api key not found")
	}
	if err != nil {
		return errors.Internal("failed to read api key", err)
	}

	record.IsActive = false
	if err := s.orm.WithContext(ctx).Save(&record).Error; err != nil {
		return errors.Internal("failed to revoke api key", err)
	}
	return nil
}

func (s *Service) AuthenticateKey(ctx context.Context, rawKey string) (userID string, email string, err error) {
	hash := hashKey(rawKey)

	var record schemas.APIKey
	dbErr := s.orm.WithContext(ctx).Where("key_hash = ? AND is_active = ?", hash, true).First(&record).Error
	if stderrors.Is(dbErr, gorm.ErrRecordNotFound) {
		return "", "", errors.Unauthorized("invalid api key")
	}
	if dbErr != nil {
		return "", "", errors.Internal("failed to validate api key", dbErr)
	}

	if record.ExpiresAt != nil && time.Now().After(*record.ExpiresAt) {
		return "", "", errors.Unauthorized("api key expired")
	}

	go func() {
		now := time.Now()
		s.orm.Model(&schemas.APIKey{}).Where("id = ?", record.ID).Update("last_used_at", now)
	}()

	var user schemas.User
	if dbErr := s.orm.Where("id = ?", record.UserID).First(&user).Error; dbErr != nil {
		return "", "", errors.Internal("failed to load user", dbErr)
	}

	return fmt.Sprintf("%d", record.UserID), user.Email, nil
}

func generateKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func toResponse(record *schemas.APIKey) *APIKeyResponse {
	return &APIKeyResponse{
		ID:         record.ID,
		Name:       record.Name,
		Prefix:     record.Prefix,
		KeyHint:    record.KeyHint,
		Scopes:     record.Scopes,
		SiteID:     record.SiteID,
		IsActive:   record.IsActive,
		LastUsedAt: record.LastUsedAt,
		ExpiresAt:  record.ExpiresAt,
		CreatedAt:  record.CreatedAt,
	}
}
