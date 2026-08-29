package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FacileStudio/Vision/apps/api/internal/database"
	documentation "github.com/FacileStudio/Vision/apps/api/internal/documentation"
	"github.com/FacileStudio/Vision/apps/api/internal/env"
	"github.com/FacileStudio/Vision/apps/api/internal/middleware"
	"github.com/FacileStudio/Vision/apps/api/modules/analytics"
	"github.com/FacileStudio/Vision/apps/api/modules/apikeys"
	"github.com/FacileStudio/Vision/apps/api/modules/auth"
	"github.com/FacileStudio/Vision/apps/api/modules/events"
	"github.com/FacileStudio/Vision/apps/api/modules/goals"
	"github.com/FacileStudio/Vision/apps/api/modules/sites"
	"github.com/FacileStudio/Vision/apps/api/modules/webhooks"
	"github.com/FacileStudio/porte/local"
	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"

	"github.com/FacileStudio/Vision/apps/api/modules/workspaces"
	"github.com/FacileStudio/Vision/apps/api/schemas"
	"gorm.io/gorm"

	"github.com/FacileStudio/Journal/sdk/journal"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/healthcheck"
	"github.com/FacileStudio/tronc/httpx"
	"github.com/FacileStudio/tronc/logger"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
)

// buildAuth constructs porte: one session manager, shared by the OIDC kit and
// the local login, over the identity tables.
//
// One manager and not two: they would each keep their own idea of the clock
// and of whether the cookie is Secure, and porte refuses a kit whose config
// disagrees with its manager's for exactly that reason. Discovery runs here,
// so an unreachable or half-configured issuer fails at boot rather than on
// somebody's first login — a change from what this app did, where a discovery
// failure at route-registration time logged an error and left SSO 404ing until
// the next restart.
//
// ConfigExtra keeps the two keys this app's client reads off /auth/config.
// porte owns that route now and writes sso_only and oidc_enabled over whatever
// the app returns, so an app cannot claim SSO is optional when it is mandatory.
// buildAuth wires porte's local and OIDC kits to Vision's stores.
//
// The local password floor is eight characters, deliberately below porte's
// default of twelve: Vision's floor has always been eight, and raising it here
// would reject a password this app accepted yesterday — a product decision,
// not a migration.
func buildAuth(ctx context.Context, db *gorm.DB, appEnv env.Config, appLogger *slog.Logger) (*session.Manager, *local.Kit, *oidc.Kit, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, nil, err
	}
	store := portepg.New(sqlDB)
	users := auth.NewUserStore(db)
	cfg := appEnv.Porte()

	sessions, err := session.New(cfg, session.Deps{Sessions: store.Sessions(), Logger: appLogger})
	if err != nil {
		return nil, nil, nil, err
	}
	kit, err := oidc.New(ctx, cfg, oidc.Deps{
		Users:      users,
		Identities: store.Identities(),
		Sessions:   sessions,
		Codes:      store.LoginCodes(),
		Logger:     appLogger,
		ConfigExtra: func() map[string]any {
			if appEnv.OIDC == nil {
				return nil
			}
			return map[string]any{
				"oidc_redirect_url": appEnv.OIDC.RedirectURL,
				"oidc_issuer":       appEnv.OIDC.Issuer,
			}
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}
	passwords, err := local.New(local.Config{AllowRegistration: !appEnv.SSOOnly, MinPasswordLength: 8}, local.Deps{
		Users:      users,
		Identities: store.Identities(),
		Sessions:   sessions,
		Logger:     appLogger,
		Count:      users.CountUsers,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return sessions, passwords, kit, nil
}

func main() {
	if healthcheck.Handle(os.Args) {
		return
	}

	os.Exit(run())
}

// run returns the process exit code. Every failure below used to return from
// main, which exits 0 — so a failed migration or an unreachable database looked
// to Docker, Dokploy and any supervisor like a clean shutdown, and a broken
// deploy reported success.
//
// Trusted proxies: behind Traefik and Cloudflare, RemoteAddr is only the
// visitor if both are trusted. Traefik replaces the forwarded chain rather than
// extending it, so the visitor survives in Cf-Connecting-Ip alone, and
// TRUSTED_PROXIES=private,cloudflare fills all three.
//
// The API's routes sit at the root: the client container strips /api before
// proxying, so the server sees /auth/me. With the default prefix every request
// was classified as static and logged at the quiet level, which is why this app
// appeared to log nothing.
//
// A port already bound, or a listener that dies under us, is a failure to
// start. Logging it and falling through used to return 0, so a container that
// never served a request looked healthy.
func run() int {
	appEnv, err := env.Load()
	appLogger := logger.New(logger.Config{})
	if err != nil {
		appLogger.Error("failed to load config", slog.Any("error", err))
		return 1
	}
	var journalClient *journal.Client
	appLogger = logger.New(logger.Config{
		Level: appEnv.LogLevel,
		Wrap: func(handler slog.Handler) slog.Handler {
			if appEnv.JournalURL == "" || appEnv.JournalToken == "" {
				return handler
			}
			journalClient = journal.New(journal.Config{URL: appEnv.JournalURL, Token: appEnv.JournalToken})
			return journal.NewHandler(journalClient, handler)
		},
	})
	if journalClient != nil {
		defer journalClient.Close()
	}

	db, err := database.Open(appEnv.DatabaseURL)
	if err != nil {
		appLogger.Error("failed to open database", slog.Any("error", err))
		return 1
	}

	if err := schemas.MigrateWithIssuer(db, appEnv.IssuerForMigration()); err != nil {
		appLogger.Error("failed to run migrations", slog.Any("error", err))
		return 1
	}
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("failed to access database handle", slog.Any("error", err))
		return 1
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			appLogger.Error("failed to close database", slog.Any("error", err))
		}
	}()

	sessions, passwords, kit, err := buildAuth(context.Background(), db, appEnv, appLogger)
	if err != nil {
		appLogger.Error("failed to build authentication", slog.Any("error", err))
		return 1
	}
	authService := auth.NewService(db, sessions, passwords, appLogger)
	workspaceService := workspaces.NewService(db)
	siteService := sites.NewService(db)
	eventHub := events.NewHub()
	activeTracker := events.NewActiveTracker()
	eventService := events.NewService(db)
	eventService.SetHub(eventHub)
	analyticsService := analytics.NewService(db)
	goalService := goals.NewService(db)
	apiKeyService := apikeys.NewService(db)
	middleware.SetAPIKeyAuthenticator(apiKeyService)
	router := httpx.NewRouter(httpx.Config{
		TrustedProxies: appEnv.TrustedProxies,
		CDNProxies:     appEnv.CDNProxies,
		CDNHeader:      appEnv.CDNHeader,
		APIPrefix:      httpx.RootAPI,
		Logger:         appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins: []string{appEnv.Domain},
		},
	})

	health.Mount(router, health.DB(sqlDB))
	apiref.Mount(router, referenceConfig())

	sessions.Mount(router)
	kit.Mount(router)
	auth.RegisterRoutes(router, authService, appEnv)
	workspaces.RegisterRoutes(router, workspaceService, authService)
	sites.RegisterRoutes(router, siteService, authService)
	events.RegisterRoutes(router, eventService, eventHub, activeTracker, authService, db)
	analytics.RegisterRoutes(router, analyticsService, activeTracker, db, authService)

	goals.RegisterRoutes(router, goalService, authService)
	apikeys.RegisterRoutes(router, apiKeyService, authService)

	webhookService := webhooks.NewService(db)
	webhooks.RegisterRoutes(router, webhookService, analyticsService, authService)
	webhooks.StartScheduler(db, analyticsService)

	addr := ":" + appEnv.Port
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	appLogger.Info("server starting", slog.String("addr", addr))
	select {
	case err := <-serverErrCh:
		if !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("server stopped", slog.Any("error", err))
			return 1
		}
	case <-shutdownSignal.Done():
		appLogger.Info("server shutting down")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			appLogger.Error("server shutdown failed", slog.Any("error", err))
			return 1
		}
		appLogger.Info("server stopped")
	}

	return 0
}

// referenceConfig describes the API reference served at /docs. Registry paths
// are written relative to /api, the one server every documented route hangs off.
func referenceConfig() apiref.Config {
	return apiref.Config{
		Title:       "Vision API",
		Description: "Self-hosted, privacy-friendly web analytics.",
		Servers:     []string{"/api"},
		Registry: documentation.Response{Modules: []documentation.Module{
			auth.Documentation,
			workspaces.Documentation,
			sites.Documentation,
			events.Documentation,
			analytics.Documentation,
			goals.Documentation,
			apikeys.Documentation,
			webhooks.Documentation,
		}},
	}
}
