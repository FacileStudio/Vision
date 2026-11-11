package auth

import (
	"context"
	"database/sql"

	"github.com/FacileStudio/porte/kit"
)

// EraseData deletes what an erased person owns in Vision, inside porte/kit's erase
// transaction, before the kit deletes the users row.
type EraseData func(ctx context.Context, tx *sql.Tx, userID int64) error

// NewStore builds porte/kit's account store over Vision's users table. The picture lives in
// oidc_picture_url, and Vision has no admin flag: roles live in workspace_members.
//
// GORM created the table, so the text and timestamp columns the kit does not name have no
// database default. fillNewAccount writes into them what GORM wrote, so an account the kit
// creates reads like every older one.
func NewStore(ctx context.Context, sqlDB *sql.DB, erase EraseData) (*kit.Store, error) {
	return kit.NewStore(ctx, kit.Store{
		DB:        sqlDB,
		Columns:   kit.Columns{Picture: "oidc_picture_url"},
		OnCreate:  fillNewAccount,
		EraseData: erase,
	})
}

const newAccountDefaults = `UPDATE users SET
	password_hash = '', avatar_url = '', avatar_source = '',
	oidc_access_token = '', oidc_refresh_token = '',
	oidc_token_expiry = '0001-01-01 00:00:00+00', profile_synced_at = '0001-01-01 00:00:00+00'
	WHERE id = $1`

func fillNewAccount(ctx context.Context, tx *sql.Tx, userID int64) error {
	_, err := tx.ExecContext(ctx, newAccountDefaults, userID)
	return err
}
