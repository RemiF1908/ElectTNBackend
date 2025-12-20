package handlers

import (
	"elect-tn-backend/internal/api"
	"elect-tn-backend/internal/auth"
	"elect-tn-backend/internal/config"
	"elect-tn-backend/internal/services"
)

// Server implements the API ServerInterface
var _ api.ServerInterface = (*ServerHandler)(nil)

type ServerHandler struct {
	electionService  *services.ElectionService
	userService      *services.UserService
	voteService      *services.VoteService
	candidacyService *services.CandidacyService
	config           *config.Config
	sessionStore     *auth.SessionStore
}

// NewServer creates a new server instance
func NewServer(
	electionService *services.ElectionService,
	userService *services.UserService,
	voteService *services.VoteService,
	candidacyService *services.CandidacyService,
	cfg *config.Config,
) *ServerHandler {
	sessionStore := auth.NewSessionStore(cfg.SessionSecret, cfg.Mode)

	return &ServerHandler{
		electionService:  electionService,
		userService:      userService,
		voteService:      voteService,
		candidacyService: candidacyService,
		config:           cfg,
		sessionStore:     sessionStore,
	}
}

// GetSessionStore returns the session store for authorization middleware
func (s *ServerHandler) GetSessionStore() *auth.SessionStore {
	return s.sessionStore
}
