package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Vision/apps/api/internal/env"
	"github.com/FacileStudio/Vision/apps/api/modules/analytics"
	"github.com/FacileStudio/Vision/apps/api/modules/apikeys"
	"github.com/FacileStudio/Vision/apps/api/modules/auth"
	"github.com/FacileStudio/Vision/apps/api/modules/events"
	"github.com/FacileStudio/Vision/apps/api/modules/goals"
	"github.com/FacileStudio/Vision/apps/api/modules/mcp"
	"github.com/FacileStudio/Vision/apps/api/modules/sites"
	"github.com/FacileStudio/Vision/apps/api/modules/webhooks"
	"github.com/FacileStudio/Vision/apps/api/modules/workspaces"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/go-chi/chi/v5"
)

func buildTestRouter() chi.Router {
	router := chi.NewRouter()
	apiref.Mount(router, referenceConfig())
	router.Method(http.MethodGet, "/profile", http.RedirectHandler("/settings", http.StatusPermanentRedirect))

	appEnv := env.Config{Domain: "localhost"}
	sessions, _ := session.New(appEnv.Porte(), session.Deps{})
	authService := auth.NewService(nil, sessions, nil, nil)

	api := chi.NewRouter()
	api.Get("/version", versionHandler)
	apiref.Mount(api, referenceConfig())
	auth.RegisterRoutes(api, authService)
	workspaces.RegisterRoutes(api, nil, authService)
	sites.RegisterRoutes(api, nil, authService)
	events.RegisterRoutes(api, nil, nil, nil, authService, nil)
	analytics.RegisterRoutes(api, nil, nil, nil, authService)
	goals.RegisterRoutes(api, nil, authService)
	apikeys.RegisterRoutes(api, nil, authService)
	webhooks.RegisterRoutes(api, nil, nil, authService)
	mcp.RegisterPendingRoutes(api, nil, authService)
	router.Mount("/api", api)

	return router
}

func TestEveryRouteIsDocumented(t *testing.T) {
	router := buildTestRouter()
	if missing := apiref.Undocumented(router, referenceConfig(), "/profile", "/api/version", "/api/docs"); len(missing) > 0 {
		t.Errorf("routes missing from the API registry: %v", missing)
	}
}

func TestRegistryIsComplete(t *testing.T) {
	if issues := apiref.Incomplete(
		referenceConfig(),
		"/e/p",
		"/e/t",
		"/e/h",
		"/auth/logout",
		"/auth/oidc",
		"/auth/oidc/callback",
		"/auth/backchannel-logout",
		"/auth/sync-profile",
		"/analytics/{siteId}/export",
		"/workspaces/{id}/leave",
		"/webhooks/{id}/test",
		"/sites/{id}/share",
	); len(issues) > 0 {
		t.Errorf("incomplete documentation routes: %v", issues)
	}
}

func TestEveryAPIRouteHangsOffTheAPIPrefix(t *testing.T) {
	root := map[string]bool{"/docs": true, "/docs/openapi.json": true, "/profile": true}
	_ = chi.Walk(buildTestRouter(), func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if !root[route] && !strings.HasPrefix(route, "/api/") {
			t.Errorf("%s %s is mounted outside /api, where the client catch-all would shadow it", method, route)
		}
		return nil
	})
}

func TestReferenceIsServedAtDocs(t *testing.T) {
	router := buildTestRouter()

	page := httptest.NewRecorder()
	router.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if page.Code != http.StatusOK {
		t.Fatalf("GET /docs = %d, want 200", page.Code)
	}

	spec := httptest.NewRecorder()
	router.ServeHTTP(spec, httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil))
	if spec.Code != http.StatusOK {
		t.Fatalf("GET /docs/openapi.json = %d, want 200", spec.Code)
	}
	var document struct {
		OpenAPI string         `json:"openapi"`
		Paths   map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(spec.Body.Bytes(), &document); err != nil {
		t.Fatalf("spec is not JSON: %v", err)
	}
	if document.OpenAPI == "" || len(document.Paths) == 0 {
		t.Fatalf("spec is empty: %+v", document)
	}
}
