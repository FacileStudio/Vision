package workspaces

import (
	"net/http"

	"api/internal/authcontext"
	"api/internal/httpjson"
	"api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/workspaces", func(r chi.Router) {
		r.Use(middleware.RequireAuth(authService))

		r.Post("/", func(w http.ResponseWriter, req *http.Request) {
			identity, _ := authcontext.IdentityFromContext(req.Context())
			var body CreateRequest
			if err := httpjson.DecodeJSON(w, req, &body); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.create(req.Context(), identity.UserID, &body)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			identity, _ := authcontext.IdentityFromContext(req.Context())
			resp, err := service.controller.list(req.Context(), identity.UserID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
			identity, _ := authcontext.IdentityFromContext(req.Context())
			resp, err := service.controller.get(req.Context(), identity.UserID, chi.URLParam(req, "id"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		r.Put("/{id}", func(w http.ResponseWriter, req *http.Request) {
			identity, _ := authcontext.IdentityFromContext(req.Context())
			var body UpdateRequest
			if err := httpjson.DecodeJSON(w, req, &body); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.update(req.Context(), identity.UserID, chi.URLParam(req, "id"), &body)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		r.Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
			identity, _ := authcontext.IdentityFromContext(req.Context())
			if err := service.controller.delete(req.Context(), identity.UserID, chi.URLParam(req, "id")); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		r.Get("/{id}/members", func(w http.ResponseWriter, req *http.Request) {
			identity, _ := authcontext.IdentityFromContext(req.Context())
			resp, err := service.controller.listMembers(req.Context(), identity.UserID, chi.URLParam(req, "id"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		r.Post("/{id}/members", func(w http.ResponseWriter, req *http.Request) {
			identity, _ := authcontext.IdentityFromContext(req.Context())
			var body AddMemberRequest
			if err := httpjson.DecodeJSON(w, req, &body); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.addMember(req.Context(), identity.UserID, chi.URLParam(req, "id"), &body)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		r.Put("/{id}/members/{userId}", func(w http.ResponseWriter, req *http.Request) {
			identity, _ := authcontext.IdentityFromContext(req.Context())
			var body UpdateMemberRequest
			if err := httpjson.DecodeJSON(w, req, &body); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			if err := service.controller.updateMember(req.Context(), identity.UserID, chi.URLParam(req, "id"), chi.URLParam(req, "userId"), &body); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		r.Delete("/{id}/members/{userId}", func(w http.ResponseWriter, req *http.Request) {
			identity, _ := authcontext.IdentityFromContext(req.Context())
			if err := service.controller.removeMember(req.Context(), identity.UserID, chi.URLParam(req, "id"), chi.URLParam(req, "userId")); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
}
