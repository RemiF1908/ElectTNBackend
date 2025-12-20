package handlers

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime/types"
)

// Helper function to convert string to pointer string
func stringPtr(s string) *string {
	return &s
}

// Helper function to convert bool to pointer bool
func boolPtr(b bool) *bool {
	return &b
}

// Helper function to convert int to pointer int32
func int32Ptr(i int32) *int32 {
	return &i
}

// Helper function to convert float32 to pointer float32
func float32Ptr(f float32) *float32 {
	return &f
}

// Helper function to convert UUID to types.UUID
func typesUUID(uuid uuid.UUID) types.UUID {
	return types.UUID(uuid)
}

// GetUserFromSession safely retrieves current user information from session
// Authentication and authorization are already validated by middleware at this point
func (s *ServerHandler) GetUserFromSession(c echo.Context) (userID uuid.UUID, isAdmin bool) {
	userIDStr, _, _, isAdmin, _ := s.sessionStore.GetUser(c)
	return uuid.MustParse(userIDStr), isAdmin
}
