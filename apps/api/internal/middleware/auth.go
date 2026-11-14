package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/FacileStudio/Vision/apps/api/internal/authcontext"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// Authenticator is the auth service: porte/kit's session middleware, which
// authenticates the credential and puts the account's kit.Identity on the request.
type Authenticator interface {
	// RequireAuth wraps a handler with porte/kit's session middleware.
	RequireAuth(http.Handler) http.Handler

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
	// AuthenticateKey returns the identity carried by a scoped API key, with the key's
	// scope on Identity.Key.
	AuthenticateKey(ctx context.Context, rawKey string) (authcontext.Identity, error)
}

var apiKeyAuth APIKeyAuthenticator

// SetAPIKeyAuthenticator installs the API-key resolver used by RequireAuth.
func SetAPIKeyAuthenticator(auth APIKeyAuthenticator) {
	apiKeyAuth = auth
}

// RequireAuth gates a handler behind an authenticator. When the request is not
// an API key it is a session: porte/kit verifies the credential and reads the
// account, and the identity is carried on in the string form the controllers take.
func RequireAuth(authService Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			authorization := request.Header.Get("Authorization")

			if apiKeyAuth != nil && isAPIKey(authorization, request) {
				key := extractAPIKey(authorization, request)
				identity, err := apiKeyAuth.AuthenticateKey(request.Context(), key)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				if !identity.Key.Write && !safeMethod(request.Method) {
					httpjson.WriteError(w, errors.Forbidden("api key is read-only"))
					return
				}
				authContext := authcontext.WithIdentity(request.Context(), identity)
				next.ServeHTTP(w, request.WithContext(authContext))
				return
			}

			session := authService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				identity, ok := kit.FromContext(request.Context())
				if !ok {
					httpjson.WriteError(w, errors.Unauthorized("missing auth"))
					return
				}
				authContext := authcontext.WithIdentity(request.Context(), authcontext.Identity{
					UserID: strconv.FormatInt(identity.UserID, 10),
					Email:  identity.Email,
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

func safeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

// RefuseKeys answers 403 to an API key. It guards the routes that manage accounts, keys,
// members and webhooks, which a key never needs to reach.
func RefuseKeys(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if authcontext.KeyFromContext(request.Context()) != nil {
			httpjson.WriteError(w, errors.Forbidden("api keys cannot use this route"))
			return
		}
		next.ServeHTTP(w, request)
	})
}
