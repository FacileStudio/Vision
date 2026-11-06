package auth

import (
	"context"
	stderrors "errors"
	"strings"

	"github.com/FacileStudio/Vision/apps/api/internal/oidcavatar"
	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// UserStore resolves a login to a row in Vision's own users table.
//
// It is porte's escape hatch, taken deliberately. porte/pg ships a porte_users
// table and Vision could have moved onto it, but workspaces, workspace
// members, sites, goals and api_keys all hang off users(id). Moving it would
// be a rewrite of everything except authentication.
//
// Vision's roles live in workspace_members, not here and not in porte: what a
// role may do is the app's business, which is exactly why porte asks the app
// to own this method rather than owning the write itself.
type UserStore struct {
	orm *gorm.DB
}

// NewUserStore returns a UserStore backed by the given GORM connection.
func NewUserStore(orm *gorm.DB) *UserStore {
	return &UserStore{orm: orm}
}

var _ porte.UserStore = (*UserStore)(nil)

// UpsertFromOIDC matches on (provider, subject) first, falls back to a verified
// email, and creates the account when neither finds anything.
//
// porte has already done the matching it can: it looks up the identity row and
// only calls this when it needs a user id. The email fallback here is what
// links a pre-existing account to a subject signing in for the first time, and
// it follows porte/pg's linking rules because porte cannot enforce them over
// Vision's own users table: the address must be verified by the provider, and
// the account must not already hold another subject of the same issuer, so a
// recycled address is never handed to a newcomer.
//
// The identity provider is the source of truth for the profile, but an absent
// claim asserts nothing: a provider that stops sending a name should not blank
// it here. oidc_picture_url is written unconditionally because this app derives
// the whole avatar from it.
func (s *UserStore) UpsertFromOIDC(ctx context.Context, claims porte.Claims) (int64, error) {
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if email == "" {
		return 0, errors.Invalid("the identity provider returned no email")
	}

	var user schemas.User
	err := s.orm.WithContext(ctx).
		Joins("JOIN porte_identities i ON i.user_id = users.id AND i.provider = ? AND i.subject = ?", claims.Provider, claims.Subject).
		First(&user).Error
	switch {
	case err == nil:
	case stderrors.Is(err, gorm.ErrRecordNotFound):
		err = s.orm.WithContext(ctx).Where("lower(email) = ?", email).Order("id").First(&user).Error
		switch {
		case err == nil:
			if err := s.checkLink(ctx, user.ID, claims); err != nil {
				return 0, err
			}
		case stderrors.Is(err, gorm.ErrRecordNotFound):
			created, createErr := s.create(ctx, email, claims.DisplayName())
			if createErr != nil {
				return 0, createErr
			}
			user = *created
		default:
			return 0, errors.Internal("failed to look up the account", err)
		}
	default:
		return 0, errors.Internal("failed to resolve the identity", err)
	}

	updates := map[string]any{"email": email, "oidc_picture_url": oidcavatar.PhotoURL(claims.Picture)}
	if displayName := claims.DisplayName(); displayName != "" {
		updates["name"] = displayName
	}
	if claims.Locale != "" {
		updates["locale"] = claims.Locale
	}
	if err := s.orm.WithContext(ctx).Model(&schemas.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return 0, errors.Internal("failed to update the account", err)
	}
	return user.ID, nil
}

// checkLink decides whether a verified-email match may link this subject to an
// existing account. The provider must have verified the address, and the account
// must hold no identity under the same issuer with a different subject.
func (s *UserStore) checkLink(ctx context.Context, userID int64, claims porte.Claims) error {
	if !claims.EmailVerified {
		return errors.Conflict("an account with this email already exists and the identity provider did not verify the address")
	}
	var others int64
	err := s.orm.WithContext(ctx).Table("porte_identities").
		Where("user_id = ? AND provider = ? AND subject <> ?", userID, claims.Provider, claims.Subject).
		Count(&others).Error
	if err != nil {
		return errors.Internal("failed to check the account's identities", err)
	}
	if others > 0 {
		return errors.Conflict("an account with this email is already linked to another identity")
	}
	return nil
}

// create is the one place a Vision account comes into existence, whichever
// way the human arrived.
func (s *UserStore) create(ctx context.Context, email, name string) (*schemas.User, error) {
	user := schemas.User{Email: email, Name: name}
	if err := s.orm.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, errors.Internal("failed to create the account", err)
	}
	return &user, nil
}
