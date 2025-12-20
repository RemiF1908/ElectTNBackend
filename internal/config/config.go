package config

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
)

// Config holds all application configuration
type Config struct {
	Mode          string
	Port          string
	DatabaseURL   string
	SessionSecret string
	AppURL        string
	OAuthConfig   *oauth2.Config
}

// LoadConfig loads configuration from environment variables
func Load() (*Config, error) {
	mode := os.Getenv("APP_MODE")
	if mode == "" {
		mode = "development"
	}

	// Get session secret - required in production, optional in development
	sessionSecret := os.Getenv("SESSION_SECRET")
	if mode == "production" && sessionSecret == "" {
		return nil, fmt.Errorf("SESSION_SECRET environment variable is required in production mode")
	}
	if mode == "development" && sessionSecret == "" {
		sessionSecret = "dev-secret-key-change-in-production"
		fmt.Println("WARNING: Using default session secret for development. Do NOT use this in production!")
	}

	// Database URL
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		if mode == "production" {
			return nil, fmt.Errorf("DATABASE_URL environment variable is required in production mode")
		}
		// Default for development - matches docker-compose setup
		databaseURL = "postgres://electTN:devpassword@localhost/electTN?sslmode=disable"
		fmt.Println("WARNING: Using default database URL for development. Do NOT use this in production!")
	}

	// Server port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Google OAuth configuration
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if mode == "production" {
		if googleClientID == "" {
			return nil, fmt.Errorf("GOOGLE_CLIENT_ID environment variable is required in production mode")
		}
		if googleClientSecret == "" {
			return nil, fmt.Errorf("GOOGLE_CLIENT_SECRET environment variable is required in production mode")
		}
	}

	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = fmt.Sprintf("http://localhost:%s/auth/callback", port)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	AppURL := os.Getenv("APP_URL")
	if AppURL == "" {
		AppURL = fmt.Sprintf("http://localhost:%s/auth/me", port)
	}

	config := &Config{
		Mode:          mode,
		Port:          port,
		AppURL:        AppURL,
		DatabaseURL:   databaseURL,
		SessionSecret: sessionSecret,
		OAuthConfig:   oauthConfig,
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Mode != "development" && c.Mode != "production" {
		return fmt.Errorf("APP_MODE must be 'development' or 'production', got: %s", c.Mode)
	}

	if c.Mode == "production" {
		if c.SessionSecret == "dev-secret-key-change-in-production" {
			return fmt.Errorf("cannot use default development secret in production mode")
		}
		if c.DatabaseURL == "postgres://user:password@localhost/elections_dev?sslmode=disable" {
			return fmt.Errorf("cannot use default development database URL in production mode")
		}
		if c.OAuthConfig.ClientID == "" {
			return fmt.Errorf("Google OAuth Client ID is required in production mode")
		}
		if c.OAuthConfig.ClientSecret == "" {
			return fmt.Errorf("Google OAuth Client Secret is required in production mode")
		}
		if c.OAuthConfig.RedirectURL == "" {
			return fmt.Errorf("Google OAuth Redirect URL is required in production mode")
		}
		if c.AppURL == "" {
			return fmt.Errorf("APP_URL is required in production mode")
		}
	}

	return nil
}

// IsDevelopment returns true if the application is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Mode == "development"
}

// IsProduction returns true if the application is running in production mode
func (c *Config) IsProduction() bool {
	return c.Mode == "production"
}

// GetServerAddress returns the server address in host:port format
func (c *Config) GetServerAddress() string {
	return ":" + c.Port
}
