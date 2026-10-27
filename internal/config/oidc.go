package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// oidcDiscovery holds the subset of the OpenID Connect discovery document we need
type oidcDiscovery struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
}

// oauthSettings groups the values required to build an oauth2.Config
type oauthSettings struct {
	clientID             string
	clientSecret         string
	redirectURL          string
	scopes               []string
	authURL              string
	tokenURL             string
	userInfoURL          string
	authStyle            oauth2.AuthStyle
	providerName         string
	issuer               string
	requireVerifiedEmail bool
	authCodeOptions      []oauth2.AuthCodeOption
}

// loadOAuthSettings builds the OAuth/OIDC configuration from the environment.
//
// When OIDC_ISSUER is set, the provider is configured generically using OpenID
// Connect discovery. Otherwise the legacy Google configuration (GOOGLE_*) is
// used, so existing deployments keep working unchanged.
func loadOAuthSettings(mode, port string) (*oauthSettings, error) {
	issuer := strings.TrimRight(strings.TrimSpace(os.Getenv("OIDC_ISSUER")), "/")
	generic := issuer != ""

	clientID := os.Getenv("OIDC_CLIENT_ID")
	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	if !generic {
		clientID = os.Getenv("GOOGLE_CLIENT_ID")
		clientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	}

	redirectURL := firstNonEmpty(
		os.Getenv("OIDC_REDIRECT_URL"),
		os.Getenv("GOOGLE_REDIRECT_URL"),
		fmt.Sprintf("http://localhost:%s/auth/callback", port),
	)

	scopes := parseScopes(os.Getenv("OIDC_SCOPES"))
	if len(scopes) == 0 {
		if generic {
			scopes = []string{"openid", "email", "profile"}
		} else {
			scopes = []string{"email", "profile"}
		}
	}

	providerName := strings.TrimSpace(os.Getenv("OIDC_PROVIDER_NAME"))
	if providerName == "" {
		if generic {
			providerName = "OIDC"
		} else {
			providerName = "Google"
		}
	}

	settings := &oauthSettings{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		scopes:       scopes,
		providerName: providerName,
		issuer:       issuer,
		authStyle:    parseAuthStyle(os.Getenv("OIDC_AUTH_STYLE")),
	}

	if generic {
		discoveryURL := os.Getenv("OIDC_DISCOVERY_URL")
		if discoveryURL == "" {
			discoveryURL = issuer + "/.well-known/openid-configuration"
		}

		doc, err := discoverOIDC(discoveryURL)
		if err != nil {
			if os.Getenv("OIDC_AUTHORIZATION_URL") == "" || os.Getenv("OIDC_TOKEN_URL") == "" || os.Getenv("OIDC_USERINFO_URL") == "" {
				return nil, fmt.Errorf("OIDC discovery failed for %q: %w", discoveryURL, err)
			}
			doc = &oidcDiscovery{}
		} else if doc.Issuer != "" && issuer != "" && strings.TrimRight(doc.Issuer, "/") != issuer {
			return nil, fmt.Errorf("OIDC issuer mismatch: configured %q, discovery reports %q", issuer, doc.Issuer)
		}

		settings.authURL = firstNonEmpty(os.Getenv("OIDC_AUTHORIZATION_URL"), doc.AuthorizationEndpoint)
		settings.tokenURL = firstNonEmpty(os.Getenv("OIDC_TOKEN_URL"), doc.TokenEndpoint)
		settings.userInfoURL = firstNonEmpty(os.Getenv("OIDC_USERINFO_URL"), doc.UserinfoEndpoint)

		if settings.authURL == "" || settings.tokenURL == "" {
			return nil, fmt.Errorf("OIDC endpoints are incomplete: authorization_url=%q token_url=%q", settings.authURL, settings.tokenURL)
		}
		if settings.userInfoURL == "" {
			return nil, fmt.Errorf("OIDC userinfo endpoint is missing: set OIDC_USERINFO_URL or ensure the discovery document exposes userinfo_endpoint")
		}

		settings.requireVerifiedEmail = parseBoolDefault(os.Getenv("OIDC_REQUIRE_VERIFIED_EMAIL"), false)
	} else {
		// Legacy Google configuration: preserve the previous behaviour.
		settings.authURL = firstNonEmpty(os.Getenv("GOOGLE_AUTHORIZATION_URL"), "https://accounts.google.com/o/oauth2/auth")
		settings.tokenURL = firstNonEmpty(os.Getenv("GOOGLE_TOKEN_URL"), "https://oauth2.googleapis.com/token")
		settings.userInfoURL = firstNonEmpty(os.Getenv("GOOGLE_USERINFO_URL"), "https://www.googleapis.com/oauth2/v2/userinfo")
		settings.requireVerifiedEmail = parseBoolDefault(os.Getenv("OIDC_REQUIRE_VERIFIED_EMAIL"), true)
		settings.authCodeOptions = []oauth2.AuthCodeOption{
			oauth2.AccessTypeOffline,
			oauth2.SetAuthURLParam("prompt", "consent"),
		}
	}

	if mode == "production" {
		if settings.clientID == "" {
			return nil, fmt.Errorf("OIDC_CLIENT_ID (or GOOGLE_CLIENT_ID) environment variable is required in production mode")
		}
		if settings.clientSecret == "" {
			return nil, fmt.Errorf("OIDC_CLIENT_SECRET (or GOOGLE_CLIENT_SECRET) environment variable is required in production mode")
		}
	}

	return settings, nil
}

// discoverOIDC fetches and parses an OpenID Connect discovery document
func discoverOIDC(discoveryURL string) (*oidcDiscovery, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(discoveryURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var doc oidcDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("invalid discovery document: %w", err)
	}

	return &doc, nil
}

func parseScopes(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})

	scopes := make([]string, 0, len(fields))
	for _, scope := range fields {
		scope = strings.TrimSpace(scope)
		if scope != "" {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}

func parseAuthStyle(raw string) oauth2.AuthStyle {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "in_header", "header", "basic":
		return oauth2.AuthStyleInHeader
	case "in_params", "params", "post":
		return oauth2.AuthStyleInParams
	default:
		return oauth2.AuthStyleAutoDetect
	}
}

func parseBoolDefault(raw string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
