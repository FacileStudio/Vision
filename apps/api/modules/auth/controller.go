package auth

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/tronc/errors"
)

// Controller mediates the profile routes.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) getMe(context context.Context, userID string) (*ProfileResponse, error) {
	user, err := controller.service.getUserByString(context, userID)
	if err != nil {
		return nil, err
	}
	return &ProfileResponse{
		ID:           strconv.FormatInt(user.ID, 10),
		Email:        user.Email,
		Name:         user.Name,
		AvatarURL:    user.Avatar(),
		AvatarSource: user.AvatarOrigin(),
		Locale:       user.Locale,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (controller *Controller) updateMe(context context.Context, userID string, req *UpdateProfileRequest) (*ProfileResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.Invalid("invalid email")
	}
	name := strings.TrimSpace(req.Name)

	user, err := controller.service.updateProfileByString(context, userID, name, email)
	if err != nil {
		return nil, err
	}
	return &ProfileResponse{
		ID:           strconv.FormatInt(user.ID, 10),
		Email:        user.Email,
		Name:         user.Name,
		AvatarURL:    user.Avatar(),
		AvatarSource: user.AvatarOrigin(),
		Locale:       user.Locale,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	}, nil
}
