package auth

import (
	"net/http"

	"github.com/FacileStudio/Vision/apps/api/internal/authcontext"
	"github.com/FacileStudio/Vision/apps/api/internal/env"
	"github.com/FacileStudio/Vision/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the auth routes. Local register and login are only
// mounted when SSO is not enforced.
func RegisterRoutes(router chi.Router, service *Service, appEnv env.Config) {
	router.Route("/auth", func(router chi.Router) {
		if !appEnv.SSOOnly {
			router.Post("/register", func(w http.ResponseWriter, request *http.Request) {
				var req RegisterRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				resp, err := service.controller.register(w, request, &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusCreated, resp)
			})

			router.Post("/login", func(w http.ResponseWriter, request *http.Request) {
				var req LoginRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				resp, err := service.controller.login(w, request, &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})
		}

		router.Group(func(router chi.Router) {
			router.Use(middleware.RequireAuth(service))

			router.Get("/me", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				resp, err := service.controller.getMe(request.Context(), identity.UserID)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Put("/me", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				var req UpdateProfileRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				resp, err := service.controller.updateMe(request.Context(), identity.UserID, &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Put("/password", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				var req ChangePasswordRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				resp, err := service.controller.changePassword(w, request, identity.UserID, &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})
		})

	})
}
