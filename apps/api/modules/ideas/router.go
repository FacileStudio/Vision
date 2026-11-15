// Package ideas serves Bulle to Vision's own client: the per-app Ideas API and the
// explanation of one anomaly.
//
// The routes call Bulle with a token exchanged for the signed-in user. The package decides
// the two things Bulle cannot know: which Registre space the request is for, taken from a
// workspace the caller belongs to, and whether that space or person has ai.access.
package ideas

import (
	"context"
	"net/http"
	"time"

	"github.com/FacileStudio/Vision/apps/api/modules/analytics"
	"github.com/FacileStudio/tronc/bulle"

	"github.com/go-chi/chi/v5"
)

// App is Vision's name in Bulle's idea enum.
const App = "vision"

const completeTimeout = 90 * time.Second

// Config wires the routes to Bulle. Plans decides ai.access; a nil Plans leaves the
// decision to Bulle.
type Config struct {
	URL        string
	UpgradeURL string
	Token      bulle.TokenSource
	Forget     func(context.Context)
	Plans      Plans
	Analytics  *analytics.Service
	HTTPClient *http.Client
}

// Authenticator is the session middleware the routes sit behind.
type Authenticator interface {
	RequireAuth(http.Handler) http.Handler
}

// RegisterRoutes serves /ideas and /anomalies/{id}/explain behind a signed-in session on
// router. An API key never reaches them: AI is the person's own plan.
func RegisterRoutes(router chi.Router, service *Service, authService Authenticator) error {
	explainer, err := service.explainer()
	if err != nil {
		return err
	}
	router.Group(func(group chi.Router) {
		group.Use(authService.RequireAuth)
		group.Post("/anomalies/{id}/explain", explainer.handle)
	})
	var mountErr error
	router.Group(func(group chi.Router) {
		group.Use(authService.RequireAuth)
		group.Use(service.resolve)
		mountErr = bulle.MountIdeas(group, service.options())
	})
	return mountErr
}

func (s *Service) options() bulle.Options {
	return bulle.Options{
		App:         App,
		URL:         s.cfg.URL,
		TokenSource: s.cfg.Token,
		Forget:      s.cfg.Forget,
		Space:       func(r *http.Request) string { return scopeOf(r).space },
		AIAccess:    func(r *http.Request) bool { return scopeOf(r).aiAccess },
		UpgradeURL:  s.cfg.UpgradeURL,
	}
}
