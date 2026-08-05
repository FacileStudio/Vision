package webhooks

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

type Service struct {
	orm *gorm.DB
}

func NewService(orm *gorm.DB) *Service {
	return &Service{orm: orm}
}

func (s *Service) create(ctx context.Context, ownerID int64, req *CreateWebhookRequest) (*WebhookResponse, error) {
	if req.URL == "" {
		return nil, errors.Invalid("url is required")
	}
	if req.IntervalHours < 1 {
		return nil, errors.Invalid("interval_hours must be at least 1")
	}

	record := &schemas.Webhook{
		OwnerID:       ownerID,
		WorkspaceID:   req.WorkspaceID,
		URL:           req.URL,
		Secret:        req.Secret,
		Period:        periodLabel(req.IntervalHours),
		IntervalHours: req.IntervalHours,
		Enabled:       true,
	}
	if err := s.orm.WithContext(ctx).Create(record).Error; err != nil {
		return nil, errors.Internal("failed to create webhook", err)
	}

	return toResponse(record), nil
}

func (s *Service) list(ctx context.Context, ownerID int64, workspaceID int64) ([]WebhookResponse, error) {
	var records []schemas.Webhook
	q := s.orm.WithContext(ctx).Where("owner_id = ?", ownerID)
	if workspaceID > 0 {
		q = q.Where("workspace_id = ?", workspaceID)
	} else {
		q = q.Where("workspace_id IS NULL")
	}
	if err := q.Order("created_at desc").Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list webhooks", err)
	}

	out := make([]WebhookResponse, len(records))
	for i := range records {
		out[i] = *toResponse(&records[i])
	}
	return out, nil
}

func (s *Service) get(ctx context.Context, ownerID int64, webhookID int64) (*WebhookResponse, error) {
	record, err := s.findWebhook(ctx, ownerID, webhookID)
	if err != nil {
		return nil, err
	}
	return toResponse(record), nil
}

func (s *Service) update(ctx context.Context, ownerID int64, webhookID int64, req *UpdateWebhookRequest) (*WebhookResponse, error) {
	record, err := s.findWebhook(ctx, ownerID, webhookID)
	if err != nil {
		return nil, err
	}

	if req.URL == "" {
		return nil, errors.Invalid("url is required")
	}
	if req.IntervalHours < 1 {
		return nil, errors.Invalid("interval_hours must be at least 1")
	}

	record.URL = req.URL
	if req.Secret != "" {
		record.Secret = req.Secret
	}
	record.IntervalHours = req.IntervalHours
	record.Period = periodLabel(req.IntervalHours)
	record.Enabled = req.Enabled

	if err := s.orm.WithContext(ctx).Save(record).Error; err != nil {
		return nil, errors.Internal("failed to update webhook", err)
	}
	return toResponse(record), nil
}

func (s *Service) delete(ctx context.Context, ownerID int64, webhookID int64) error {
	record, err := s.findWebhook(ctx, ownerID, webhookID)
	if err != nil {
		return err
	}

	if err := s.orm.WithContext(ctx).Delete(record).Error; err != nil {
		return errors.Internal("failed to delete webhook", err)
	}
	return nil
}

func (s *Service) findWebhook(ctx context.Context, ownerID int64, webhookID int64) (*schemas.Webhook, error) {
	var record schemas.Webhook
	err := s.orm.WithContext(ctx).Where("id = ? AND owner_id = ?", webhookID, ownerID).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("webhook not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to read webhook", err)
	}
	return &record, nil
}

func getIntervalHours(wh *schemas.Webhook) int {
	if wh.IntervalHours > 0 {
		return wh.IntervalHours
	}
	switch wh.Period {
	case "hourly":
		return 1
	case "daily":
		return 24
	case "weekly":
		return 168
	case "monthly":
		return 720
	}
	return 24
}

func toResponse(w *schemas.Webhook) *WebhookResponse {
	hours := getIntervalHours(w)
	resp := &WebhookResponse{
		ID:            w.ID,
		URL:           w.URL,
		Period:        periodLabel(hours),
		IntervalHours: hours,
		Enabled:       w.Enabled,
		WorkspaceID:   w.WorkspaceID,
		CreatedAt:     w.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     w.UpdatedAt.Format(time.RFC3339),
	}
	if w.LastSentAt != nil {
		formatted := w.LastSentAt.Format(time.RFC3339)
		resp.LastSentAt = &formatted
	}
	return resp
}
