package auth

import (
	"context"
	stderrors "errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"
)

// Controller mediates registration, login, profile and password routes.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) register(w http.ResponseWriter, r *http.Request, req *RegisterRequest) (*AuthResponse, error) {
	context := r.Context()
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.Invalid("invalid email")
	}
	if len(req.Password) < 8 {
		return nil, errors.Invalid("password must be at least 8 characters")
	}

	userID, token, err := controller.service.Register(context, w, r, email, req.Password)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{UserID: userID, Token: token}, nil
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
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (controller *Controller) changePassword(w http.ResponseWriter, r *http.Request, userID string, req *ChangePasswordRequest) (*PasswordResponse, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Internal("failed to parse user id", err)
	}
	if req.CurrentPassword == "" {
		if err := controller.service.SetPassword(r.Context(), id, req.NewPassword); err != nil {
			return nil, passwordError(err)
		}
		return &PasswordResponse{Status: "ok"}, nil
	}
	token, _, err := controller.service.ChangePassword(w, r, id, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return nil, passwordError(err)
	}
	return &PasswordResponse{Status: "ok", Token: token}, nil
}

// passwordError maps porte's sentinels onto the answers the suite agreed on.
//
// The length floor is porte's alone: it holds the argon2 parameters and the
// MinPasswordLength this app configures, so a second check here is a number
// that drifts. Only ErrPasswordSet is re-coded. porte answers 409 because from
// inside the kit it is a race, but from here it is a caller that left
// current_password out of the body, which is a 400 naming the field.
func passwordError(err error) error {
	switch {
	case stderrors.Is(err, porte.ErrPasswordSet):
		return errors.Invalid("current_password is required to replace an existing password")
	case stderrors.Is(err, porte.ErrWrongPassword):
		return errors.Unauthorized("current password is incorrect")
	case stderrors.Is(err, porte.ErrNoPassword):
		return errors.Invalid("this account has no password; omit current_password to set one")
	case stderrors.Is(err, porte.ErrWeakPassword):
		return errors.Invalid("new password must be at least 8 characters")
	}
	return err
}

func (controller *Controller) login(w http.ResponseWriter, r *http.Request, req *LoginRequest) (*AuthResponse, error) {
	context := r.Context()
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		return nil, errors.Invalid("email and password required")
	}

	userID, token, err := controller.service.Login(context, w, r, email, req.Password)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{UserID: userID, Token: token}, nil
}
