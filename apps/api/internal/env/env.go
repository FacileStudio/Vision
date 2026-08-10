package env

import (
	"github.com/FacileStudio/porte"
	troncenv "github.com/FacileStudio/tronc/env"

	"fmt"
	"net/netip"
	"strconv"
	"strings"
)

type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	SuccessURL   string
}

type Config struct {
	// TrustedProxies, CDNProxies and CDNHeader decide what RemoteAddr is by
	// the time a rate limiter sees it. All three come from TRUSTED_PROXIES;
	// see tronc's configuration docs.
	TrustedProxies []netip.Prefix
	CDNProxies     []netip.Prefix
	CDNHeader      string

	DatabaseURL  string
	Port         string
	Domain       string
	LogLevel     string
	OIDC         *OIDCConfig
	SSOOnly      bool
	JournalURL   string
	JournalToken string
}

func Load() (Config, error) {
	trustedProxies, err := troncenv.TrustedProxies()
	if err != nil {
		return Config{}, err
	}
	cdnProxies, cdnHeader := troncenv.CDN()

	env := Config{
		TrustedProxies: trustedProxies,
		CDNProxies:     cdnProxies,
		CDNHeader:      cdnHeader,
		DatabaseURL:    valueOrDefault("DATABASE_URL", "postgres://postgres:postgres@db:5432/vision?sslmode=disable"),
		Port:           valueOrDefault("PORT", "4000"),
		Domain:         valueOrDefault("DOMAIN", "http://localhost:5173"),
		LogLevel:       valueOrDefault("LOG_LEVEL", "info"),
	}

	port, err := strconv.Atoi(env.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a valid TCP port")
	}
	if err := validateLogLevel(env.LogLevel); err != nil {
		return Config{}, err
	}

	env.SSOOnly = strings.ToLower(envGet("SSO_ONLY")) == "true"
	env.JournalURL = envGet("JOURNAL_URL")
	env.JournalToken = envGet("JOURNAL_TOKEN")

	if issuer := envGet("OIDC_ISSUER"); issuer != "" {
		clientID := envGet("OIDC_CLIENT_ID")
		clientSecret := envGet("OIDC_CLIENT_SECRET")
		redirectURL := envGet("OIDC_REDIRECT_URL")
		if clientID == "" || clientSecret == "" || redirectURL == "" {
			return Config{}, fmt.Errorf("OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and OIDC_REDIRECT_URL are required when OIDC_ISSUER is set")
		}
		successURL := envGet("OIDC_SUCCESS_URL")
		if successURL == "" {
			successURL = env.Domain
		}
		env.OIDC = &OIDCConfig{
			Issuer:       issuer,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			SuccessURL:   successURL,
		}
	}

	return env, nil
}

func valueOrDefault(key string, fallback string) string {
	if value := envGet(key); value != "" {
		return value
	}
	return fallback
}

func validateLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error")
	}
}

// Porte is the one configuration porte's session manager, OIDC kit and local
// login are all built from. They share it because porte refuses at boot a kit
// whose config disagrees with its manager's — a mismatch would otherwise
// change silently whether the session cookie is Secure.
//
// AcceptLegacyCookie is on even though this app has always been bearer-only:
// it costs nothing when no such cookie exists, and porte now sets one, so the
// setting describes the transport rather than the migration.
func (c Config) Porte() porte.Config {
	cfg := porte.Config{SSOOnly: c.SSOOnly, AcceptLegacyCookie: true}
	if c.OIDC == nil {
		return cfg
	}
	cfg.Issuer = c.OIDC.Issuer
	cfg.ClientID = c.OIDC.ClientID
	cfg.ClientSecret = c.OIDC.ClientSecret
	cfg.RedirectURL = c.OIDC.RedirectURL
	cfg.SuccessURL = c.OIDC.SuccessURL
	return cfg
}

// IssuerForMigration is the issuer the identity backfill keys on, or empty
// when SSO is not configured. It exists so the migration cannot be handed a
// placeholder: an identity row written under the wrong provider matches
// nothing and degrades to the email fallback in silence.
func (c Config) IssuerForMigration() string {
	if c.OIDC == nil {
		return ""
	}
	return c.OIDC.Issuer
}
