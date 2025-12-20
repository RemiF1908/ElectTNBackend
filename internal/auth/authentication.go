package auth

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3filter"
	middleware "github.com/oapi-codegen/echo-middleware"
)

// SessionAuthenticator provides session-based authentication for oapi-codegen
type SessionAuthenticator struct {
	sessionStore *SessionStore
}

// NewSessionAuthenticator creates a new session authenticator
func NewSessionAuthenticator(sessionStore *SessionStore) *SessionAuthenticator {
	return &SessionAuthenticator{
		sessionStore: sessionStore,
	}
}

// NewAuthenticator creates the oapi-codegen authentication function
func NewAuthenticator(sessionStore *SessionStore) openapi3filter.AuthenticationFunc {
	authenticator := NewSessionAuthenticator(sessionStore)
	return authenticator.Authenticate
}

// Authenticate validates sessions and checks security requirements
// This function is called during the oapi-codegen validation phase,
// AFTER oapi-codegen wrappers set context values but BEFORE handlers are called
func (a *SessionAuthenticator) Authenticate(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	// Only handle our security schemes
	if input.SecuritySchemeName != "SessionAuth" && input.SecuritySchemeName != "AdminAuth" {
		return fmt.Errorf("unsupported security scheme: %s", input.SecuritySchemeName)
	}

	// Get Echo context to access sessions
	eCtx := middleware.GetEchoContext(ctx)
	if eCtx == nil {
		return fmt.Errorf("echo context not found")
	}

	// Validate session and get user info (no need to set in context since handlers use session directly)
	_, _, _, isAdmin, loggedIn := a.sessionStore.GetUser(eCtx)

	// Check if user is authenticated
	if !loggedIn {
		return fmt.Errorf("user not authenticated")
	}

	// Check admin requirements for AdminAuth
	if len(input.Scopes) > 0 && input.Scopes[0] == "admin" && !isAdmin {
		return fmt.Errorf("admin access required")
	}

	// Authentication successful - handlers can access user data from session when needed
	return nil
}
