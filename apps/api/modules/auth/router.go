package auth

import (
	"net/http"

	"github.com/FacileStudio/Vision/apps/api/internal/authcontext"
	"github.com/FacileStudio/Vision/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the profile routes. Sign-in itself is porte's OIDC kit and
// there is no register, login or password route: Registre is the only way in.
func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/auth", func(router chi.Router) {
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

			router.With(middleware.RefuseKeys).Put("/me", func(w http.ResponseWriter, request *http.Request) {
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
		})

	})
}
