package handlers

import (
	"elect-tn-backend/internal/api"
	"time"

	"github.com/labstack/echo/v4"
)

// Health check endpoint
func (s *ServerHandler) GetHealth(c echo.Context) error {
	now := time.Now()
	status := "healthy"
	return api.GetHealth200JSONResponse{
		Status:    &status,
		Timestamp: &now,
	}.VisitGetHealthResponse(c.Response())
}
