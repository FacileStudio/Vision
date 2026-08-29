package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Vision/apps/api/internal/env"
	"github.com/FacileStudio/Vision/apps/api/modules/analytics"
	"github.com/FacileStudio/Vision/apps/api/modules/apikeys"
	"github.com/FacileStudio/Vision/apps/api/modules/auth"
	"github.com/FacileStudio/Vision/apps/api/modules/events"
	"github.com/FacileStudio/Vision/apps/api/modules/goals"
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

	appEnv := env.Config{Domain: "localhost"}
	sessions, _ := session.New(appEnv.Porte(), session.Deps{})
	authService := auth.NewService(nil, sessions, nil, nil)

	auth.RegisterRoutes(router, authService, appEnv)
	workspaces.RegisterRoutes(router, nil, authService)
	sites.RegisterRoutes(router, nil, authService)
	events.RegisterRoutes(router, nil, nil, nil, authService, nil)
	analytics.RegisterRoutes(router, nil, nil, nil, authService)
	goals.RegisterRoutes(router, nil, authService)
	apikeys.RegisterRoutes(router, nil, authService)
	webhooks.RegisterRoutes(router, nil, nil, authService)

	return router
}

func TestEveryRouteIsDocumented(t *testing.T) {
	router := buildTestRouter()
	if missing := apiref.Undocumented(router, referenceConfig()); len(missing) > 0 {
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
