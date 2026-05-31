package middleware

import (
	"context"
	"net/http"
	"strings"

	"api/internal/authcontext"
	"api/internal/errors"
	"api/internal/httpjson"
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

func isAPIKey(authorization string, r *http.Request) bool {
	if strings.HasPrefix(authorization, "Bearer vis_") {
		return true
	}
	if r.Header.Get("X-API-Key") != "" {
		return true
	}
	if r.URL.Query().Get("api_key") != "" {
		return true
	}
	return false
}

func extractAPIKey(authorization string, r *http.Request) string {
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}
	if strings.HasPrefix(authorization, "Bearer vis_") {
		return strings.TrimPrefix(authorization, "Bearer ")
	}
	if key := r.URL.Query().Get("api_key"); key != "" {
		return key
	}
	return ""
}
