-- +goose Up
-- Rows written outside GORM, porte/kit's sign-up insert first among them, got NULL
-- where GORM used to fill the column itself. The database now fills the timestamps,
-- and the users columns kit.Store does not name get the values modules/auth's
-- fillNewAccount writes.

ALTER TABLE api_keys ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE custom_events ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE goals ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE pageviews ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE sites ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE sites ALTER COLUMN updated_at SET DEFAULT now();
ALTER TABLE visitor_sessions ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE webhooks ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE webhooks ALTER COLUMN updated_at SET DEFAULT now();
ALTER TABLE workspace_members ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE workspaces ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE workspaces ALTER COLUMN updated_at SET DEFAULT now();

ALTER TABLE users
    ALTER COLUMN created_at SET DEFAULT now(),
    ALTER COLUMN password_hash SET DEFAULT '',
    ALTER COLUMN avatar_url SET DEFAULT '',
    ALTER COLUMN avatar_source SET DEFAULT '',
    ALTER COLUMN oidc_access_token SET DEFAULT '',
    ALTER COLUMN oidc_refresh_token SET DEFAULT '',
    ALTER COLUMN oidc_token_expiry SET DEFAULT '0001-01-01 00:00:00+00',
    ALTER COLUMN profile_synced_at SET DEFAULT '0001-01-01 00:00:00+00';

-- +goose Down
ALTER TABLE users
    ALTER COLUMN created_at DROP DEFAULT,
    ALTER COLUMN password_hash DROP DEFAULT,
    ALTER COLUMN avatar_url DROP DEFAULT,
    ALTER COLUMN avatar_source DROP DEFAULT,
    ALTER COLUMN oidc_access_token DROP DEFAULT,
    ALTER COLUMN oidc_refresh_token DROP DEFAULT,
    ALTER COLUMN oidc_token_expiry DROP DEFAULT,
    ALTER COLUMN profile_synced_at DROP DEFAULT;

ALTER TABLE workspaces ALTER COLUMN updated_at DROP DEFAULT;
ALTER TABLE workspaces ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE workspace_members ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE webhooks ALTER COLUMN updated_at DROP DEFAULT;
ALTER TABLE webhooks ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE visitor_sessions ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE sites ALTER COLUMN updated_at DROP DEFAULT;
ALTER TABLE sites ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE pageviews ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE goals ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE custom_events ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE api_keys ALTER COLUMN created_at DROP DEFAULT;
