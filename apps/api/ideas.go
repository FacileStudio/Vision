package main

import (
	"context"
	"log/slog"
	"strings"

	"github.com/FacileStudio/Vision/apps/api/internal/env"
	"github.com/FacileStudio/Vision/apps/api/internal/plans"
	"github.com/FacileStudio/Vision/apps/api/modules/analytics"
	"github.com/FacileStudio/Vision/apps/api/modules/ideas"
	"github.com/FacileStudio/porte/exchange"

	"gorm.io/gorm"
)

// newIdeas builds the Bulle service and its plan gate when BULLE_URL is set and SSO is on,
// and nothing otherwise. Vision acts as the signed-in user through a Registre token
// exchange for audience bulle, over the identity store that seals the Registre tokens.
// The caller feeds every directory snapshot to the gate, which loads the last saved
// plans first.
func newIdeas(
	ctx context.Context, db *gorm.DB, appEnv env.Config, authn authentication,
	analyticsService *analytics.Service, appLogger *slog.Logger,
) (*ideas.Service, *plans.Gate, error) {
	if appEnv.OIDC == nil || strings.TrimSpace(appEnv.BulleURL) == "" {
		return nil, nil, nil
	}
	exchanger, err := exchange.New(exchange.Config{
		Issuer:       appEnv.OIDC.Issuer,
		ClientID:     appEnv.OIDC.ClientID,
		ClientSecret: appEnv.OIDC.ClientSecret,
		Identities:   authn.porte.Identities(),
		Logger:       appLogger,
	})
	if err != nil {
		return nil, nil, err
	}
	gate := plans.New(db, appEnv.OIDC.Issuer, appLogger)
	if err := gate.Load(ctx); err != nil {
		return nil, nil, err
	}
	source := exchanger.Source(exchange.Target{Audience: "bulle", Scopes: []string{"bulle.read", "bulle.write"}})
	return ideas.NewService(db, ideas.Config{
		URL:        strings.TrimSpace(appEnv.BulleURL),
		UpgradeURL: strings.TrimRight(appEnv.OIDC.Issuer, "/") + "/spaces",
		Token:      source.Token,
		Forget:     source.Forget,
		Plans:      gate,
		Analytics:  analyticsService,
	}), gate, nil
}
