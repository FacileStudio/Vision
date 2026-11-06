package auth

import (
	"context"
	"strconv"
	"strings"
	"time"
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
	name := strings.TrimSpace(req.Name)

	user, err := controller.service.updateProfileByString(context, userID, name)
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
