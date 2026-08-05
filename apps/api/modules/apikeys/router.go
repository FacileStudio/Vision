package apikeys

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/Vision/apps/api/internal/authcontext"
	"github.com/FacileStudio/Vision/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/api-keys", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Post("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			var req CreateRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.create(request.Context(), identity.UserID, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		router.Get("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			var workspaceID int64
			if raw := request.URL.Query().Get("workspace_id"); raw != "" {
				workspaceID, _ = strconv.ParseInt(raw, 10, 64)
			}
			resp, err := service.controller.list(request.Context(), identity.UserID, workspaceID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			err := service.controller.revoke(request.Context(), identity.UserID, chi.URLParam(request, "id"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
}
