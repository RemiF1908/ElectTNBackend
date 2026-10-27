package handlers

import (
	"context"
	"crypto/rand"
	"elect-tn-backend/internal/api"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime/types"
)

// OIDCUserInfo represents the subset of user information returned by an OIDC
// userinfo endpoint. It supports both the standard OIDC claims
// (email_verified) and the Google-specific claim (verified_email).
type OIDCUserInfo struct {
	Subject           string `json:"sub"`
	Email             string `json:"email"`
	EmailVerified     *bool  `json:"email_verified"`
	VerifiedEmail     *bool  `json:"verified_email"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Picture           string `json:"picture"`
	Locale            string `json:"locale"`
}

// isEmailVerified reports the verification status and whether the provider
// sent the information at all.
func (u OIDCUserInfo) isEmailVerified() (verified bool, present bool) {
	if u.EmailVerified != nil {
		return *u.EmailVerified, true
	}
	if u.VerifiedEmail != nil {
		return *u.VerifiedEmail, true
	}
	return false, false
}

// displayName returns the best available human-readable name for the user.
func (u OIDCUserInfo) displayName() string {
	for _, candidate := range []string{u.Name, u.PreferredUsername, u.GivenName, u.Email} {
		if candidate != "" {
			return candidate
		}
	}
	return ""
}

// generateSecureState generates a secure random state token for OAuth
func generateSecureState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Auth endpoints
func (s *ServerHandler) GetAuthLogin(c echo.Context) error {
	// Check if OAuth is configured
	if s.config.OAuthConfig.ClientID == "" {
		return api.GetAuthLogin500JSONResponse(api.Error{
			Error:   stringPtr("oauth_not_configured"),
			Message: stringPtr(fmt.Sprintf("%s authentication is not configured", s.config.ProviderName)),
		}).VisitGetAuthLoginResponse(c.Response())
	}

	// Generate secure state token
	state, err := generateSecureState()
	if err != nil {
		return api.GetAuthLogin500JSONResponse(api.Error{
			Error:   stringPtr("state_generation_failed"),
			Message: stringPtr("Failed to generate secure state token"),
		}).VisitGetAuthLoginResponse(c.Response())
	}

	// Store state and timestamp in session
	session, err := s.sessionStore.GetSession(c)
	if err != nil {
		return api.GetAuthLogin500JSONResponse(api.Error{
			Error:   stringPtr("session_error"),
			Message: stringPtr("Failed to get session"),
		}).VisitGetAuthLoginResponse(c.Response())
	}

	session.Values["oauth_state"] = state
	session.Values["oauth_timestamp"] = time.Now().Unix()
	if err := s.sessionStore.SaveSession(c, session); err != nil {
		return api.GetAuthLogin500JSONResponse(api.Error{
			Error:   stringPtr("session_save_failed"),
			Message: stringPtr("Failed to save session"),
		}).VisitGetAuthLoginResponse(c.Response())
	}

	// Generate OAuth URL with state
	authURL := s.config.OAuthConfig.AuthCodeURL(state, s.config.AuthCodeOptions...)
	return c.Redirect(http.StatusFound, authURL)
}

func (s *ServerHandler) GetAuthCallback(c echo.Context, params api.GetAuthCallbackParams) error {
	// Check for OAuth errors from the provider
	if params.Error != nil {
		errorMsg := "OAuth authentication failed"
		if params.ErrorDescription != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *params.ErrorDescription)
		}
		fmt.Printf("OAuth error: %s - %s\n", *params.Error, errorMsg)
		return api.GetAuthCallback400JSONResponse(api.Error{
			Error:   params.Error,
			Message: &errorMsg,
		}).VisitGetAuthCallbackResponse(c.Response())
	}

	// Validate required parameters
	if params.Code == "" || params.State == "" {
		return api.GetAuthCallback400JSONResponse(api.Error{
			Error:   stringPtr("invalid_request"),
			Message: stringPtr("Missing required OAuth parameters"),
		}).VisitGetAuthCallbackResponse(c.Response())
	}

	// Retrieve session and validate state
	session, err := s.sessionStore.GetSession(c)
	if err != nil {
		return api.GetAuthCallback500JSONResponse(api.Error{
			Error:   stringPtr("internal_error"),
			Message: stringPtr("Failed to get session"),
		}).VisitGetAuthCallbackResponse(c.Response())
	}

	storedState, stateExists := session.Values["oauth_state"].(string)
	storedTimestamp, timestampExists := session.Values["oauth_timestamp"].(int64)

	// Clear OAuth data from session
	delete(session.Values, "oauth_state")
	delete(session.Values, "oauth_timestamp")
	_ = s.sessionStore.SaveSession(c, session) // Ignore save error for cleanup

	// Validate state
	if !stateExists || !timestampExists || storedState != params.State {
		return api.GetAuthCallback400JSONResponse(api.Error{
			Error:   stringPtr("invalid_state"),
			Message: stringPtr("Invalid or expired OAuth state"),
		}).VisitGetAuthCallbackResponse(c.Response())
	}

	// Check state expiration (5 minutes)
	if time.Now().Unix()-storedTimestamp > 300 {
		return api.GetAuthCallback400JSONResponse(api.Error{
			Error:   stringPtr("expired_state"),
			Message: stringPtr("OAuth state has expired"),
		}).VisitGetAuthCallbackResponse(c.Response())
	}

	// Exchange authorization code for tokens
	token, err := s.config.OAuthConfig.Exchange(context.Background(), params.Code)
	if err != nil {
		fmt.Printf("Failed to exchange code for token: %v\n", err)
		return api.GetAuthCallback500JSONResponse(api.Error{
			Error:   stringPtr("token_exchange_failed"),
			Message: stringPtr("Failed to exchange authorization code for tokens"),
		}).VisitGetAuthCallbackResponse(c.Response())
	}

	// Get user info from the configured OIDC userinfo endpoint
	client := s.config.OAuthConfig.Client(context.Background(), token)
	userInfoResp, err := client.Get(s.config.UserInfoURL)
	if err != nil {
		fmt.Printf("Failed to get user info: %v\n", err)
		return api.GetAuthCallback500JSONResponse(api.Error{
			Error:   stringPtr("user_info_failed"),
			Message: stringPtr(fmt.Sprintf("Failed to retrieve user information from %s", s.config.ProviderName)),
		}).VisitGetAuthCallbackResponse(c.Response())
	}
	defer userInfoResp.Body.Close()

	if userInfoResp.StatusCode != http.StatusOK {
		fmt.Printf("User info API returned status: %d\n", userInfoResp.StatusCode)
		return api.GetAuthCallback500JSONResponse(api.Error{
			Error:   stringPtr("user_info_error"),
			Message: stringPtr("User info API returned an error"),
		}).VisitGetAuthCallbackResponse(c.Response())
	}

	// Parse user info
	var userInfo OIDCUserInfo
	if err := json.NewDecoder(userInfoResp.Body).Decode(&userInfo); err != nil {
		fmt.Printf("Failed to parse user info: %v\n", err)
		return api.GetAuthCallback500JSONResponse(api.Error{
			Error:   stringPtr("parse_error"),
			Message: stringPtr("Failed to parse user information"),
		}).VisitGetAuthCallbackResponse(c.Response())
	}

	// Validate email. When the provider reports verification status, an
	// explicitly unverified email is always rejected. Providers that omit the
	// claim are accepted unless OIDC_REQUIRE_VERIFIED_EMAIL=true.
	if userInfo.Email == "" {
		return api.GetAuthCallback400JSONResponse(api.Error{
			Error:   stringPtr("invalid_email"),
			Message: stringPtr("Email is required"),
		}).VisitGetAuthCallbackResponse(c.Response())
	}
	if verified, present := userInfo.isEmailVerified(); (present && !verified) || (!present && s.config.RequireVerifiedEmail) {
		return api.GetAuthCallback400JSONResponse(api.Error{
			Error:   stringPtr("invalid_email"),
			Message: stringPtr("Email must be verified"),
		}).VisitGetAuthCallbackResponse(c.Response())
	}

	// Create or update user in database
	user, err := s.userService.CreateOrUpdateUser(userInfo.Email, userInfo.displayName(), token.AccessToken)
	if err != nil {
		fmt.Printf("Failed to create/update user: %v\n", err)
		return api.GetAuthCallback500JSONResponse(api.Error{
			Error:   stringPtr("user_creation_failed"),
			Message: stringPtr("Failed to create or update user"),
		}).VisitGetAuthCallbackResponse(c.Response())
	}

	// Set user session
	if err := s.sessionStore.SetUser(c, user.UUID.String(), user.Email, user.Name, user.IsAdmin); err != nil {
		fmt.Printf("Failed to create session: %v\n", err)
		return api.GetAuthCallback500JSONResponse(api.Error{
			Error:   stringPtr("session_creation_failed"),
			Message: stringPtr("Failed to create user session"),
		}).VisitGetAuthCallbackResponse(c.Response())
	}

	// Redirect to application after successful authentication
	redirectURL := s.config.AppURL
	return c.Redirect(http.StatusFound, redirectURL)
}

func (s *ServerHandler) GetAuthLogout(c echo.Context) error {
	// Clear the session
	_ = s.sessionStore.ClearSession(c) // Ignore errors in session clearing

	return api.GetAuthLogout200Response(api.GetAuthLogout200Response{}).VisitGetAuthLogoutResponse(c.Response())
}

func (s *ServerHandler) GetAuthMe(c echo.Context) error {
	// Get user info from session (authentication already validated by oapi-codegen middleware)
	userIDStr, email, name, isAdmin, _ := s.sessionStore.GetUser(c)

	// Parse UUID (should be valid since middleware validated authentication)
	userUUID, _ := uuid.Parse(userIDStr)

	// Get user from database to get complete info
	if s.userService != nil {
		user, err := s.userService.GetUserByUUID(userUUID)
		if err == nil {
			// Update session info if database has different values
			email = user.Email
			name = user.Name
			isAdmin = user.IsAdmin
		}
	}

	// Convert to API response
	uuidVal := types.UUID(userUUID)
	emailVal := types.Email(email)
	namePtr := stringPtr(name)
	now := time.Now()

	return api.GetAuthMe200JSONResponse(api.User{
		Uuid:      &uuidVal,
		Email:     &emailVal,
		Name:      namePtr,
		IsAdmin:   &isAdmin,
		CreatedAt: &now,
		UpdatedAt: &now,
	}).VisitGetAuthMeResponse(c.Response())
}
