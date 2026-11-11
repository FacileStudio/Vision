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
// the OIDC kit, as a first web login would.
func newMCP(ctx context.Context, appEnv env.Config, authn authentication, deps mcp.Deps, appLogger *slog.Logger) (*mcp.Service, error) {
	if appEnv.OIDC == nil || appEnv.OIDC.MCPResource == "" {
		return nil, nil
	}
	sqlDB, err := deps.DB.DB()
	if err != nil {
		return nil, err
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
	appLogger.Info("mcp server on", slog.String("resource", appEnv.OIDC.MCPResource))
	return mcp.NewService(deps, mcp.Config{
		Resource: appEnv.OIDC.MCPResource,
		Issuer:   appEnv.OIDC.Issuer,
		Version:  version,
		Origins:  []string{appEnv.Domain},
		StateKey: appEnv.OIDC.MCPStateKey(),
		Verifier: verifier,
		Pending:  mcpkit.PostgresPending{DB: sqlDB},
		Logger:   appLogger,
	})
}
