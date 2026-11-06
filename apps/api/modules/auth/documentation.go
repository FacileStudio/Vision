package auth

import (
	documentation "github.com/FacileStudio/Vision/apps/api/internal/documentation"
)

type ConfigResponse struct {
	SSOOnly     bool   `json:"sso_only"`
	OIDCEnabled bool   `json:"oidc_enabled"`
	OIDCIssuer  string `json:"oidc_issuer,omitempty"`
}

type ExchangeRequest struct {
	Code string `json:"code" validate:"required"`
}

type SyncProfileResponse struct {
	Success bool `json:"success"`
}

var Documentation = documentation.Module{
	Name:        "auth",
	Description: "Authentication routes.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/auth/config",
			Summary:      "Describe the auth methods on offer",
			Description:  "Returns sso_only (always true while Registre is configured) and oidc_enabled, plus this app's oidc_issuer and oidc_redirect_url. Served by porte.",
			Auth:         "public",
			ResponseBody: ConfigResponse{},
		},
		{
			Method:      "POST",
			Path:        "/auth/logout",
			Summary:     "End the session",
			Description: "Revokes the session and clears the cookie. Idempotent: a caller whose session has already expired still gets 200, because that is exactly when somebody needs to log out.",
			Auth:        "public",
		},
		{
			Method:      "GET",
			Path:        "/auth/oidc",
			Summary:     "Start the SSO login",
			Description: "Redirects to the identity provider with PKCE and a nonce. Registered only when OIDC_ISSUER is set.",
			Auth:        "public",
		},
		{
			Method:      "GET",
			Path:        "/auth/oidc/callback",
			Summary:     "Complete the SSO login",
			Description: "Verifies the ID token, upserts the account and issues a session. A failure redirects to the login page carrying ?error= — a refusal and a success are both a 302, so read Location and not the status.",
			Auth:        "public",
		},
		{
			Method:       "POST",
			Path:         "/auth/oidc/exchange",
			Summary:      "Exchange a CLI login code for a token",
			Description:  "Consumes the one-time code handed to a ?flow=cli login. Single use: a replay finds nothing.",
			Auth:         "public",
			RequestBody:  ExchangeRequest{},
			ResponseBody: AuthResponse{},
		},
		{
			Method:       "POST",
			Path:         "/auth/sync-profile",
			Summary:      "Refresh the profile from the provider",
			Description:  "Calls UserInfo with the stored refresh token and updates the name and photo. Rate-limited server-side.",
			Auth:         "bearer",
			ResponseBody: SyncProfileResponse{},
		},
		{
			Method:      "POST",
			Path:        "/auth/backchannel-logout",
			Summary:     "Revoke sessions on the provider's behalf",
			Description: "Called by the identity provider, not the client. Validates the logout token and deletes that user's sessions.",
			Auth:        "signed logout token",
		},
		{
			Method:       "GET",
			Path:         "/auth/me",
			Summary:      "Get current user profile",
			Description:  "Returns the authenticated user's profile.",
			Auth:         "bearer",
			ResponseBody: ProfileResponse{},
		},
		{
			Method:       "PUT",
			Path:         "/auth/me",
			Summary:      "Update current user profile",
			Description:  "Updates the authenticated user's name. The email field is ignored: the address comes from Registre.",
			Auth:         "bearer",
			RequestBody:  UpdateProfileRequest{},
			ResponseBody: ProfileResponse{},
		},
	},
}
