package main

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Vision/apps/api/internal/env"
	"github.com/FacileStudio/Vision/apps/api/modules/auth"
	"github.com/FacileStudio/Vision/apps/api/modules/workspaces"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// authentication is what buildAuth assembles: porte's session manager, its OIDC kit,
// porte's identity tables, and porte/kit's account store over Vision's users table.
type authentication struct {
	sessions  *session.Manager
	oidc      *oidc.Kit
	porte     *portepg.Store
	users     *kit.Store
	deletions *workspaces.Deletions
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
func buildAuth(ctx context.Context, db *gorm.DB, appEnv env.Config, appLogger *slog.Logger) (authentication, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return authentication{}, err
	}
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
	deletions := newDeletions(db, store, users, appEnv, appLogger)
	deps := oidc.Deps{
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
	}
	if deletions != nil {
		deps.Erased = deletions
	}
	oidcKit, err := oidc.New(ctx, cfg, deps)
	if err != nil {
		return authentication{}, err
	}
	return authentication{sessions: sessions, oidc: oidcKit, porte: store, users: users, deletions: deletions}, nil
}

// newDeletions builds the receiver for Registre's erasures when Vision signs in through
// Registre, and nil otherwise. It exists before the OIDC kit so the kit can refuse a
// subject Vision already erased, and its Client.Issuer is OIDC_ISSUER, the provider
// porte/oidc files identities under, so the re-check after a race matches them.
func newDeletions(db *gorm.DB, store *portepg.Store, users *kit.Store, appEnv env.Config, appLogger *slog.Logger) *workspaces.Deletions {
	if appEnv.OIDC == nil {
		return nil
	}
	client := directory.Client{Issuer: appEnv.OIDC.Issuer, ClientID: appEnv.OIDC.ClientID, ClientSecret: appEnv.OIDC.ClientSecret}
	return workspaces.NewDeletions(db, client, workspaces.Erasers{
		Account:  users.EraseAccount,
		Identity: store.EraseIdentity,
		Holds:    store.HoldsIdentity,
	}, appLogger)
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

// mountInvitations serves GET /me/invitations, the signed-in person's pending Registre space
// invitations for muse's user menu, behind the session, on the /api router, so the public
// path is /api/me/invitations. It asks Registre with the directory credentials.
func mountInvitations(router chi.Router, authn authentication, client directory.Client, appLogger *slog.Logger) error {
	return kit.MountInvitations(router.With(kit.RequireAuth(authn.sessions, authn.users)), kit.InvitationsConfig{
		Directory:  client,
		Identities: authn.porte.Identities(),
		Logger:     appLogger,
	})
}
