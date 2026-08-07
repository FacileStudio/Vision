package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/FacileStudio/Vision/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

type Authenticator interface {
	Authenticate(context context.Context, authorization string) (string, any, error)
}

type APIKeyAuthenticator interface {
	AuthenticateKey(ctx context.Context, rawKey string) (userID string, email string, err error)
}

var apiKeyAuth APIKeyAuthenticator

func SetAPIKeyAuthenticator(auth APIKeyAuthenticator) {
	apiKeyAuth = auth
}

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

			userID, rawData, err := authService.Authenticate(request.Context(), authorization)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			data, ok := rawData.(interface{ GetEmail() string })
			if !ok {
				httpjson.WriteError(w, errors.Unauthorized("missing auth"))
				return
			}
			if data == nil {
				httpjson.WriteError(w, errors.Unauthorized("missing auth"))
				return
			}

			authContext := authcontext.WithIdentity(request.Context(), authcontext.Identity{
				UserID: userID,
				Email:  data.GetEmail(),
			})
			next.ServeHTTP(w, request.WithContext(authContext))
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
