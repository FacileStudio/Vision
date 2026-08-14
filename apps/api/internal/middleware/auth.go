package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/FacileStudio/Vision/apps/api/internal/authcontext"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// Authenticator is the auth service: porte's session middleware, plus the
// lookup that turns the user id porte resolved into the identity the rest of
// Vision reads.
type Authenticator interface {
	// RequireAuth wraps a handler with porte's session middleware.
	RequireAuth(http.Handler) http.Handler
	// IdentityForUser resolves the user id porte authenticated into the
	// identity the rest of Vision reads.
	IdentityForUser(ctx context.Context, userID int64) (id string, email string, err error)

	// AuthenticateRequest authenticates the live-events stream. EventSource
	// cannot set a request header.
	AuthenticateRequest(w http.ResponseWriter, r *http.Request) (int64, error)
	// AuthenticateToken authenticates a bearer token that arrived in the
	// query string instead of the Authorization header.
	AuthenticateToken(w http.ResponseWriter, r *http.Request, token string) (int64, error)
}

// APIKeyAuthenticator resolves a scoped API key to an identity, for
// machine-to-machine access.
type APIKeyAuthenticator interface {
	// AuthenticateKey returns the identity carried by a scoped API key.
	AuthenticateKey(ctx context.Context, rawKey string) (userID string, email string, err error)
}

var apiKeyAuth APIKeyAuthenticator

// SetAPIKeyAuthenticator installs the API-key resolver used by RequireAuth.
func SetAPIKeyAuthenticator(auth APIKeyAuthenticator) {
	apiKeyAuth = auth
}

// RequireAuth gates a handler behind an authenticator. When the request is not
// an API key it is a session: porte verifies the credential and hands on a user
// id, and the profile the rest of Vision reads is looked up here — porte carries
// no email by design.
func RequireAuth(authService Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			authorization := request.Header.Get("Authorization")

			if apiKeyAuth != nil && isAPIKey(authorization, request) {
				key := extractAPIKey(authorization, request)
				userID, email, err := apiKeyAuth.AuthenticateKey(request.Context(), key)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				authContext := authcontext.WithIdentity(request.Context(), authcontext.Identity{
					UserID: userID,
					Email:  email,
				})
				next.ServeHTTP(w, request.WithContext(authContext))
				return
			}

			session := authService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				authenticated, ok := porte.From(request.Context())
				if !ok {
					httpjson.WriteError(w, errors.Unauthorized("missing auth"))
					return
				}
				userID, email, err := authService.IdentityForUser(request.Context(), authenticated.UserID)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				authContext := authcontext.WithIdentity(request.Context(), authcontext.Identity{
					UserID: userID,
					Email:  email,
				})
				next.ServeHTTP(w, request.WithContext(authContext))
			}))
			session.ServeHTTP(w, request)
		})
	}
}

// isAPIKey and extractAPIKey read the two transports the API reference
// documents: `Authorization: Bearer vis_…` and `X-API-Key`.
//
// An `?api_key=` query parameter used to be accepted as a third. It was never
// documented, the tracker does not use it, and nothing in the suite builds such
// a URL — but an API key in a query string is copied into access logs, Referer
// headers and browser history, and Vision's keys are long-lived. Removed
// 2026-08-07.
//
// This is not the same as the `?token=` on GET /events/{siteId}/live, which
// stays: EventSource cannot set headers, so that one is a real constraint of
// the browser API rather than a convenience.
func isAPIKey(authorization string, r *http.Request) bool {
	if strings.HasPrefix(authorization, "Bearer vis_") {
		return true
	}
	return r.Header.Get("X-API-Key") != ""
}

func extractAPIKey(authorization string, r *http.Request) string {
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}
	if strings.HasPrefix(authorization, "Bearer vis_") {
		return strings.TrimPrefix(authorization, "Bearer ")
	}
	return ""
}
