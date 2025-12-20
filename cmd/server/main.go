package main

import (
	"log"

	"elect-tn-backend/internal/app"
	"elect-tn-backend/internal/config"
	"elect-tn-backend/internal/database"
	"elect-tn-backend/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Test database connection
	if err := database.Ping(db); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connected successfully")

	// Initialize services with database connection
	electionService := services.NewElectionService(db)
	userService := services.NewUserService(db)
	voteService := services.NewVoteService(db)
	candidacyService := services.NewCandidacyService(db)

	// Create server instance
	srv := app.NewServer(
		electionService,
		userService,
		voteService,
		candidacyService,
		cfg,
	)

	// Start server
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
