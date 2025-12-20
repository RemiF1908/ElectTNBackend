package handlers

import (
	"elect-tn-backend/internal/api"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime/types"
)

func (s *ServerHandler) PostElections(c echo.Context) error {
	// Authorization handled by middleware

	var req api.CreateElectionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, api.Error{
			Error:   stringPtr("invalid_request"),
			Message: stringPtr("Failed to parse request body"),
		})
	}

	if req.Title == "" {
		return api.PostElections400JSONResponse(api.Error{
			Error:   stringPtr("invalid_request"),
			Message: stringPtr("Title is required"),
		}).VisitPostElectionsResponse(c.Response())
	}

	// Get current user information from session (authentication already validated by middleware)
	currentUserUUID, _ := s.GetUserFromSession(c)

	// Debug logging
	fmt.Printf("DEBUG: Current user UUID from session: %s\n", currentUserUUID.String())

	var description string
	if req.Description != nil {
		description = *req.Description
	}

	var isLiveResults bool
	if req.IsLiveResults != nil {
		isLiveResults = *req.IsLiveResults
	}

	election, err := s.electionService.CreateElection(
		req.Title,
		description,
		currentUserUUID,
		isLiveResults,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.Error{
			Error:   stringPtr("internal_error"),
			Message: stringPtr("Failed to create election"),
		})
	}

	// Convert to API response
	uuidVal := types.UUID(election.UUID)
	descriptionPtr := stringPtr(election.Description)
	isOpen := election.IsOpen
	createdByUUID := types.UUID(election.CreatedByUUID)
	createdAt := election.CreatedAt
	updatedAt := election.UpdatedAt

	// Use generated response type
	return api.PostElections201JSONResponse(api.Election{
		Uuid:          &uuidVal,
		Title:         &req.Title,
		Description:   descriptionPtr,
		IsOpen:        &isOpen,
		IsLiveResults: &isLiveResults,
		CreatedByUuid: &createdByUUID,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt,
	}).VisitPostElectionsResponse(c.Response())
}

func (s *ServerHandler) PutElection(c echo.Context, electionUuid types.UUID) error {
	// Authorization handled by middleware

	// Parse request body
	var req api.UpdateElectionRequest
	if err := c.Bind(&req); err != nil {
		return api.PutElection400JSONResponse(api.Error{
			Error:   stringPtr("invalid_request"),
			Message: stringPtr("Failed to parse request body"),
		}).VisitPutElectionResponse(c.Response())
	}

	// Get current user information from session
	currentUserUUID, isAdmin := s.GetUserFromSession(c)

	election, err := s.electionService.UpdateElection(
		uuid.UUID(electionUuid),
		req.Title,
		req.Description,
		req.IsLiveResults,
		currentUserUUID,
		isAdmin,
	)
	if err != nil {
		if err.Error() == "election not found" {
			return api.PutElection404JSONResponse(api.Error{
				Error:   stringPtr("not_found"),
				Message: stringPtr("Election not found"),
			}).VisitPutElectionResponse(c.Response())
		}
		if err.Error() == "unauthorized: admin or creator access required" {
			return api.PutElection403JSONResponse(api.Error{
				Error:   stringPtr("forbidden"),
				Message: stringPtr("Admin or creator access required"),
			}).VisitPutElectionResponse(c.Response())
		}
		return c.JSON(http.StatusInternalServerError, api.Error{
			Error:   stringPtr("internal_error"),
			Message: stringPtr("Failed to update election"),
		})
	}

	// Convert to API response
	uuidVal := types.UUID(election.UUID)
	descriptionPtr := stringPtr(election.Description)
	isOpen := election.IsOpen
	isLiveResults := election.IsLiveResults
	createdByUUID := types.UUID(election.CreatedByUUID)
	createdAt := election.CreatedAt
	updatedAt := election.UpdatedAt

	return api.PutElection200JSONResponse(api.Election{
		Uuid:          &uuidVal,
		Title:         stringPtr(election.Title),
		Description:   descriptionPtr,
		IsOpen:        &isOpen,
		IsLiveResults: &isLiveResults,
		CreatedByUuid: &createdByUUID,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt,
	}).VisitPutElectionResponse(c.Response())
}

func (s *ServerHandler) DeleteElection(c echo.Context, electionUuid types.UUID) error {
	// Authorization handled by middleware

	// Get current user information from session
	currentUserUUID, isAdmin := s.GetUserFromSession(c)

	err := s.electionService.DeleteElection(uuid.UUID(electionUuid), currentUserUUID, isAdmin)
	if err != nil {
		if err.Error() == "election not found" {
			return c.JSON(http.StatusNotFound, api.Error{
				Error:   stringPtr("not_found"),
				Message: stringPtr("Election not found"),
			})
		}
		if err.Error() == "unauthorized: admin or creator access required" {
			return c.JSON(http.StatusForbidden, api.Error{
				Error:   stringPtr("forbidden"),
				Message: stringPtr("Admin or creator access required"),
			})
		}
		return c.JSON(http.StatusInternalServerError, api.Error{
			Error:   stringPtr("internal_error"),
			Message: stringPtr("Failed to delete election"),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *ServerHandler) PostElectionOpen(c echo.Context, electionUuid types.UUID) error {
	// Authorization handled by middleware

	// Parse request body for isOpen flag
	var req struct {
		IsOpen bool `json:"is_open"`
	}

	_ = c.Bind(&req) // Ignore binding errors for simplicity

	// Get current user information from session (authentication already validated by middleware)
	currentUserUUID, isAdmin := s.GetUserFromSession(c)

	if s.electionService != nil {
		election, err := s.electionService.ToggleElectionOpen(uuid.UUID(electionUuid), req.IsOpen, currentUserUUID, isAdmin)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, api.Error{
				Error:   stringPtr("internal_error"),
				Message: stringPtr("Failed to update election status"),
			})
		}

		// Convert to API response
		uuidVal := types.UUID(election.UUID)
		descriptionPtr := stringPtr(election.Description)
		isLiveResults := election.IsLiveResults
		createdByUUID := types.UUID(election.CreatedByUUID)
		createdAt := election.CreatedAt
		updatedAt := election.UpdatedAt

		return c.JSON(http.StatusOK, api.Election{
			Uuid:          &uuidVal,
			Title:         stringPtr(election.Title),
			Description:   descriptionPtr,
			IsOpen:        &election.IsOpen,
			IsLiveResults: &isLiveResults,
			CreatedByUuid: &createdByUUID,
			CreatedAt:     &createdAt,
			UpdatedAt:     &updatedAt,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"uuid":            electionUuid,
		"is_open":         req.IsOpen,
		"is_live_results": false,
	})
}

func (s *ServerHandler) GetElectionResults(c echo.Context, electionUuid types.UUID) error {
	// Authorization handled by middleware

	results, err := s.electionService.GetElectionResults(uuid.UUID(electionUuid))
	if err != nil {
		if err.Error() == "election not found" {
			return c.JSON(http.StatusNotFound, api.Error{
				Error:   stringPtr("not_found"),
				Message: stringPtr("Election not found"),
			})
		}
		return c.JSON(http.StatusInternalServerError, api.Error{
			Error:   stringPtr("internal_error"),
			Message: stringPtr(err.Error()),
		})
	}

	return c.JSON(http.StatusOK, results)
}
