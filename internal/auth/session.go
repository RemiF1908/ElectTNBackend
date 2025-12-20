package auth

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

const (
	SessionName = "election-session"
	UserIDKey   = "user_id"
	AdminKey    = "is_admin"
	EmailKey    = "email"
	NameKey     = "name"
)

// SessionStore manages user sessions
type SessionStore struct {
	store sessions.Store
}

// NewSessionStore creates a new session store
func NewSessionStore(sessionSecret string, appMode string) *SessionStore {
	// Use cookie store for session management
	store := sessions.NewCookieStore([]byte(sessionSecret))

	// Configure session options
	secure := appMode == "production" // Only secure in production
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}

	return &SessionStore{store: store}
}

// GetSession retrieves the current session
func (s *SessionStore) GetSession(c echo.Context) (*sessions.Session, error) {
	req := c.Request()
	return s.store.Get(req, SessionName)
}

// SaveSession saves the session to the response
func (s *SessionStore) SaveSession(c echo.Context, session *sessions.Session) error {
	req := c.Request()
	resp := c.Response()
	return session.Save(req, resp)
}

// SetUser sets user information in the session
func (s *SessionStore) SetUser(c echo.Context, userID string, email, name string, isAdmin bool) error {
	session, err := s.GetSession(c)
	if err != nil {
		return err
	}

	session.Values[UserIDKey] = userID
	session.Values[EmailKey] = email
	session.Values[NameKey] = name
	session.Values[AdminKey] = isAdmin

	return s.SaveSession(c, session)
}

// GetUser retrieves user information from the session
func (s *SessionStore) GetUser(c echo.Context) (userID string, email, name string, isAdmin bool, loggedIn bool) {
	session, err := s.GetSession(c)
	if err != nil {
		return
	}

	userIDInterface, exists := session.Values[UserIDKey]
	if !exists {
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		return
	}

	email, _ = session.Values[EmailKey].(string)
	name, _ = session.Values[NameKey].(string)
	isAdmin, _ = session.Values[AdminKey].(bool)
	loggedIn = true
	return
}

// ClearSession clears the user session (logout)
func (s *SessionStore) ClearSession(c echo.Context) error {
	session, err := s.GetSession(c)
	if err != nil {
		return err
	}

	// Clear all session values
	for key := range session.Values {
		delete(session.Values, key)
	}

	// Set max age to -1 to delete the cookie
	session.Options.MaxAge = -1

	return s.SaveSession(c, session)
}
