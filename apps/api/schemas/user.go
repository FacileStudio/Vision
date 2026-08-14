package schemas

import "time"

// User is an account that can sign in and own analytics data.
type User struct {
	ID               int64     `gorm:"column:id;primaryKey"`
	Email            string    `gorm:"column:email;uniqueIndex"`
	Name             string    `gorm:"column:name"`
	PasswordHash     string    `gorm:"column:password_hash"`
	AvatarURL        string    `gorm:"column:avatar_url"`
	AvatarSource     string    `gorm:"column:avatar_source"`
	OIDCPictureURL   string    `gorm:"column:oidc_picture_url"`
	OIDCSubject      *string   `gorm:"column:oidc_subject;uniqueIndex"`
	OIDCAccessToken  string    `gorm:"column:oidc_access_token"`
	OIDCRefreshToken string    `gorm:"column:oidc_refresh_token"`
	OIDCTokenExpiry  time.Time `gorm:"column:oidc_token_expiry"`
	ProfileSyncedAt  time.Time `gorm:"column:profile_synced_at"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (User) TableName() string { return "users" }

// Avatar is the picture to render. Vision has no upload, so the IdP's URL is the only
// source there is — nothing is copied locally, which is why there is no file to serve and
// no storage directory to lose. An empty value means the client draws initials.
func (u User) Avatar() string { return u.OIDCPictureURL }

// AvatarOrigin exists so the settings page can say where the picture comes from without
// the client re-deriving it from the URL's shape.
func (u User) AvatarOrigin() string {
	if u.OIDCPictureURL != "" {
		return "oidc"
	}
	return ""
}
