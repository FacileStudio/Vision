package main

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/FacileStudio/Vision/apps/api/internal/env"
	"github.com/FacileStudio/Vision/apps/api/modules/auth"
	"github.com/FacileStudio/Vision/apps/api/modules/workspaces"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/porte/oidc"
	"github.com/FacileStudio/porte/session"
)

// authentication is what buildAuth assembles: porte's session manager, its OIDC kit, and
// porte/kit's account store over Vision's users table.
type authentication struct {
	sessions *session.Manager
	oidc     *oidc.Kit
	users    *kit.Store
}

// buildAuth constructs porte: one session manager, the OIDC kit, and porte/kit's account
// store over Vision's users table.
//
// Sign-in goes through Registre only. porte mounts no password route while an issuer is
// configured, so Vision builds no local kit and keeps no register or login route.
// Discovery runs here, so an unreachable or half-configured issuer fails at boot rather
// than on somebody's first login.
//
// ConfigExtra keeps the two keys this app's client reads off /auth/config.
// porte owns that route and writes sso_only and oidc_enabled itself.
func buildAuth(ctx context.Context, sqlDB *sql.DB, appEnv env.Config, appLogger *slog.Logger) (authentication, error) {
	store, err := newPorteStore(sqlDB)
	if err != nil {
		return authentication{}, err
	}
	resealInBackground(ctx, store, appLogger)
	users, err := auth.NewStore(ctx, sqlDB, workspaces.EraseData)
	if err != nil {
		return authentication{}, err
	}
	backfillPictures(ctx, users, appEnv, appLogger)
	cfg := appEnv.Porte()

	sessions, err := session.New(cfg, session.Deps{Sessions: store.Sessions(), Logger: appLogger})
	if err != nil {
		return authentication{}, err
	}
	oidcKit, err := oidc.New(ctx, cfg, oidc.Deps{
		Users:      users,
		Identities: store.Identities(),
		Sessions:   sessions,
		Codes:      store.LoginCodes(),
		Logger:     appLogger,
		ConfigExtra: func() map[string]any {
			if appEnv.OIDC == nil {
				return nil
			}
			return map[string]any{
				"oidc_redirect_url": appEnv.OIDC.RedirectURL,
				"oidc_issuer":       appEnv.OIDC.Issuer,
			}
		},
	})
	if err != nil {
		return authentication{}, err
	}
	return authentication{sessions: sessions, oidc: oidcKit, users: users}, nil
}

// backfillPictures gives every SSO account the picture URL its next sign-in would store,
// <issuer>/avatars/<subject>, so a member who has not signed in since Registre began
// sending the claim still renders with a photo in a member list. It fills empty rows only.
func backfillPictures(ctx context.Context, users *kit.Store, appEnv env.Config, appLogger *slog.Logger) {
	if appEnv.OIDC == nil {
		return
	}
	filled, err := users.BackfillPictures(ctx, appEnv.OIDC.Issuer)
	if err != nil {
		appLogger.Warn("picture backfill failed", slog.Any("error", err))
		return
	}
	if filled > 0 {
		appLogger.Info("picture backfill complete", slog.Int64("filled", filled))
	}
}
