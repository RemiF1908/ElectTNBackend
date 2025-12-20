package app

import (
	"log"

	"elect-tn-backend/internal/api"
	"elect-tn-backend/internal/auth"
	"elect-tn-backend/internal/config"
	"elect-tn-backend/internal/handlers"
	"elect-tn-backend/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server represents the main application server
type Server struct {
	config   *config.Config
	echo     *echo.Echo
	handlers *handlers.ServerHandler
}

// NewServer creates a new server instance
func NewServer(
	electionService *services.ElectionService,
	userService *services.UserService,
	voteService *services.VoteService,
	candidacyService *services.CandidacyService,
	cfg *config.Config,
) *Server {
	e := echo.New()

	// Initialize handlers
	serverHandler := handlers.NewServer(
		electionService,
		userService,
		voteService,
		candidacyService,
		cfg,
	)

	// Middleware
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Authentication and validation middleware (replaces old auth middleware)
	// This integrates with oapi-codegen to handle authentication during request validation
	authValidationMiddleware, err := auth.CreateAuthMiddleware(serverHandler.GetSessionStore())
	if err != nil {
		log.Fatalf("Failed to create auth validation middleware: %v", err)
	}
	e.Use(authValidationMiddleware...)

	return &Server{
		config:   cfg,
		echo:     e,
		handlers: serverHandler,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Register routes
	s.registerRoutes()

	// Start server
	log.Printf("Starting server on %s in %s mode", s.config.GetServerAddress(), s.config.Mode)
	return s.echo.Start(s.config.GetServerAddress())
}

// registerRoutes registers all HTTP routes
func (s *Server) registerRoutes() {
	// Register generated API routes (includes health check)
	api.RegisterHandlers(s.echo, s.handlers)
}
