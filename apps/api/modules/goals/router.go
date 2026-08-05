package goals

import (
	"net/http"
	"time"

	"github.com/FacileStudio/Vision/apps/api/internal/authcontext"
	"github.com/FacileStudio/Vision/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/goals", func(router chi.Router) {
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
			siteID := request.URL.Query().Get("site_id")
			if siteID == "" {
				httpjson.WriteError(w, errors.Invalid("site_id is required"))
				return
			}
			resp, err := service.controller.list(request.Context(), identity.UserID, siteID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Put("/{id}", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			var req UpdateRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.update(request.Context(), identity.UserID, chi.URLParam(request, "id"), &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			err := service.controller.delete(request.Context(), identity.UserID, chi.URLParam(request, "id"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		router.Get("/{siteId}/conversions", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			siteID := chi.URLParam(request, "siteId")

			from, to := parseDateRange(request)
			resp, err := service.controller.conversions(request.Context(), identity.UserID, siteID, from, to)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	})
}

func parseDateRange(request *http.Request) (time.Time, time.Time) {
	now := time.Now().UTC()
	from := now.AddDate(0, 0, -30)
	to := now

	if v := request.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = t
		}
	}
	if v := request.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			to = t.Add(24*time.Hour - time.Nanosecond)
		}
	}
	return from, to
}
