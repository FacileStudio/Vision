package main

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Vision/apps/api/internal/env"
	"github.com/FacileStudio/Vision/apps/api/modules/mcp"
	"github.com/FacileStudio/porte/mcpauth"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// newMCP builds the MCP server over deps when SSO is on and MCP_RESOURCE is set, and
// nothing otherwise. The verifier runs Registre discovery here, so a wrong issuer fails at
// boot like the OIDC kit does. A person's first MCP call creates their account through
// the OIDC kit, as a first web login would. Pending actions resolved or expired a month ago
// are pruned every hour until ctx ends.
func newMCP(ctx context.Context, appEnv env.Config, authn authentication, deps mcp.Deps, appLogger *slog.Logger) (*mcp.Service, error) {
	if appEnv.OIDC == nil || appEnv.OIDC.MCPResource == "" {
		return nil, nil
	}
	verifier, err := mcpauth.Verifier(ctx, mcpauth.Config{
		Issuer:      appEnv.OIDC.Issuer,
		Resource:    appEnv.OIDC.MCPResource,
		ScopePrefix: "vision",
		Identities:  authn.porte.Identities(),
		Provision:   true,
		Accounts:    authn.oidc,
		Logger:      appLogger,
	})
	if err != nil {
		return nil, err
	}
	pending, err := keepPending(ctx, deps, appLogger)
	if err != nil {
		return nil, err
	}
	appLogger.Info("mcp server on", slog.String("resource", appEnv.OIDC.MCPResource))
	return mcp.NewService(deps, mcp.Config{
		Resource: appEnv.OIDC.MCPResource,
		Issuer:   appEnv.OIDC.Issuer,
		Version:  version,
		Origins:  []string{appEnv.Domain},
		StateKey: mcpkit.StateKey("vision", appEnv.OIDC.ClientSecret),
		Verifier: verifier,
		Pending:  pending,
		Logger:   appLogger,
	})
}

// keepPending returns the Postgres pending store and prunes it every hour until ctx ends.
func keepPending(ctx context.Context, deps mcp.Deps, appLogger *slog.Logger) (mcpkit.PostgresPending, error) {
	sqlDB, err := deps.DB.DB()
	if err != nil {
		return mcpkit.PostgresPending{}, err
	}
	pending := mcpkit.PostgresPending{DB: sqlDB}
	go mcpkit.KeepPruned(ctx, pending, 0, 0, appLogger)
	return pending, nil
}
