package config

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2"
)

func TestParseScopes(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"openid,email,profile", []string{"openid", "email", "profile"}},
		{"openid email  profile", []string{"openid", "email", "profile"}},
		{" openid , email ", []string{"openid", "email"}},
	}

	for _, tt := range tests {
		got := parseScopes(tt.input)
		if len(got) != len(tt.want) {
			t.Fatalf("parseScopes(%q) = %v, want %v", tt.input, got, tt.want)
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Fatalf("parseScopes(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	}
}

func TestParseBoolDefault(t *testing.T) {
	if !parseBoolDefault("true", false) || !parseBoolDefault("1", false) || !parseBoolDefault("yes", false) {
		t.Fatal("expected truthy values to parse to true")
	}
	if parseBoolDefault("false", true) || parseBoolDefault("0", true) || parseBoolDefault("off", true) {
		t.Fatal("expected falsy values to parse to false")
	}
	if !parseBoolDefault("", true) || parseBoolDefault("", false) {
		t.Fatal("expected empty value to use the fallback")
	}
}

func TestParseAuthStyle(t *testing.T) {
	if got := parseAuthStyle("in_header"); got != oauth2.AuthStyleInHeader {
		t.Fatalf("got %v, want AuthStyleInHeader", got)
	}
	if got := parseAuthStyle("in_params"); got != oauth2.AuthStyleInParams {
		t.Fatalf("got %v, want AuthStyleInParams", got)
	}
	if got := parseAuthStyle(""); got != oauth2.AuthStyleAutoDetect {
		t.Fatalf("got %v, want AuthStyleAutoDetect", got)
	}
}

func TestDiscoverOIDC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"issuer": "https://id.example.com",
			"authorization_endpoint": "https://id.example.com/auth",
			"token_endpoint": "https://id.example.com/token",
			"userinfo_endpoint": "https://id.example.com/userinfo"
		}`))
	}))
	defer server.Close()

	doc, err := discoverOIDC(server.URL)
	if err != nil {
		t.Fatalf("discoverOIDC returned error: %v", err)
	}
	if doc.Issuer != "https://id.example.com" {
		t.Fatalf("unexpected issuer: %q", doc.Issuer)
	}
	if doc.AuthorizationEndpoint != "https://id.example.com/auth" {
		t.Fatalf("unexpected authorization endpoint: %q", doc.AuthorizationEndpoint)
	}
	if doc.UserinfoEndpoint != "https://id.example.com/userinfo" {
		t.Fatalf("unexpected userinfo endpoint: %q", doc.UserinfoEndpoint)
	}
}

func TestLoadOAuthSettingsGeneric(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"issuer": "https://id.example.com",
			"authorization_endpoint": "https://id.example.com/auth",
			"token_endpoint": "https://id.example.com/token",
			"userinfo_endpoint": "https://id.example.com/userinfo"
		}`))
	}))
	defer server.Close()

	t.Setenv("OIDC_ISSUER", "https://id.example.com")
	t.Setenv("OIDC_DISCOVERY_URL", server.URL)
	t.Setenv("OIDC_CLIENT_ID", "client-id")
	t.Setenv("OIDC_CLIENT_SECRET", "client-secret")
	t.Setenv("OIDC_SCOPES", "")

	settings, err := loadOAuthSettings("development", "8080")
	if err != nil {
		t.Fatalf("loadOAuthSettings returned error: %v", err)
	}

	if settings.clientID != "client-id" || settings.clientSecret != "client-secret" {
		t.Fatal("unexpected client credentials")
	}
	if settings.authURL != "https://id.example.com/auth" {
		t.Fatalf("unexpected auth URL: %q", settings.authURL)
	}
	if settings.tokenURL != "https://id.example.com/token" {
		t.Fatalf("unexpected token URL: %q", settings.tokenURL)
	}
	if settings.userInfoURL != "https://id.example.com/userinfo" {
		t.Fatalf("unexpected userinfo URL: %q", settings.userInfoURL)
	}
	if len(settings.scopes) != 3 || settings.scopes[0] != "openid" {
		t.Fatalf("unexpected scopes: %v", settings.scopes)
	}
}

func TestLoadOAuthSettingsLegacyGoogle(t *testing.T) {
	t.Setenv("OIDC_ISSUER", "")
	t.Setenv("GOOGLE_CLIENT_ID", "google-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")
	t.Setenv("OIDC_SCOPES", "")

	settings, err := loadOAuthSettings("development", "8080")
	if err != nil {
		t.Fatalf("loadOAuthSettings returned error: %v", err)
	}

	if settings.providerName != "Google" {
		t.Fatalf("unexpected provider name: %q", settings.providerName)
	}
	if settings.authURL != "https://accounts.google.com/o/oauth2/auth" {
		t.Fatalf("unexpected auth URL: %q", settings.authURL)
	}
	if settings.userInfoURL != "https://www.googleapis.com/oauth2/v2/userinfo" {
		t.Fatalf("unexpected userinfo URL: %q", settings.userInfoURL)
	}
	if !settings.requireVerifiedEmail {
		t.Fatal("legacy Google should require verified emails")
	}
}
