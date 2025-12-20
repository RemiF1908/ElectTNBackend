package auth

import (
	"fmt"

	"elect-tn-backend/internal/api"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	middleware "github.com/oapi-codegen/echo-middleware"
)

// CreateAuthMiddleware creates the oapi-codegen validation middleware with authentication
// This replaces the global auth middleware and integrates authentication with request validation
func CreateAuthMiddleware(sessionStore *SessionStore) ([]echo.MiddlewareFunc, error) {
	// Get the OpenAPI spec
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("loading OpenAPI spec: %w", err)
	}

	// Create validator middleware with authentication
	// This middleware will:
	// 1. Validate the request against OpenAPI spec
	// 2. Call our authentication function during validation
	// 3. Only allow requests to proceed if both validation and authentication succeed
	validator := middleware.OapiRequestValidatorWithOptions(swagger,
		&middleware.Options{
			Options: openapi3filter.Options{
				AuthenticationFunc: NewAuthenticator(sessionStore),
			},
		})

	return []echo.MiddlewareFunc{validator}, nil
}
