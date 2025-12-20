package handlers

import (
	"elect-tn-backend/internal/api"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime/types"
)

// Election endpoints
func (s *ServerHandler) GetElections(c echo.Context) error {
	if s.electionService == nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"elections": []interface{}{},
		})
	}

	elections, _ := s.electionService.ListElections() // Assume service works

	return c.JSON(http.StatusOK, map[string]interface{}{
		"elections": elections,
	})
}

func (s *ServerHandler) GetElection(c echo.Context, electionUuid types.UUID) error {
	if s.electionService == nil {
		return api.GetElection404JSONResponse(api.Error{
			Error:   stringPtr("not_found"),
			Message: stringPtr("Service not available"),
		}).VisitGetElectionResponse(c.Response())
	}

	election, err := s.electionService.GetElection(uuid.UUID(electionUuid))
	if err != nil {
		return api.GetElection404JSONResponse(api.Error{
			Error:   stringPtr("not_found"),
			Message: stringPtr("Election not found"),
		}).VisitGetElectionResponse(c.Response())
	}

	// Convert to API response
	uuidVal := types.UUID(election.UUID)
	title := election.Title
	description := stringPtr(election.Description)
	isOpen := election.IsOpen
	isLiveResults := election.IsLiveResults
	createdByUUID := types.UUID(election.CreatedByUUID)
	createdAt := election.CreatedAt
	updatedAt := election.UpdatedAt

	return api.GetElection200JSONResponse(api.Election{
		Uuid:          &uuidVal,
		Title:         &title,
		Description:   description,
		IsOpen:        &isOpen,
		IsLiveResults: &isLiveResults,
		CreatedByUuid: &createdByUUID,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt,
	}).VisitGetElectionResponse(c.Response())
}

func (s *ServerHandler) GetElectionCandidacies(c echo.Context, electionUuid types.UUID) error {
	if s.candidacyService == nil {
		emptyCandidacies := []api.Candidacy{}
		return api.GetElectionCandidacies200JSONResponse{
			Candidacies: &emptyCandidacies,
		}.VisitGetElectionCandidaciesResponse(c.Response())
	}

	// Get candidacies with user information
	candidacies, err := s.candidacyService.GetCandidaciesWithUserByElection(uuid.UUID(electionUuid))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.Error{
			Error:   stringPtr("internal_error"),
			Message: stringPtr("Failed to get candidacies"),
		})
	}

	// Convert to API response
	apiCandidacies := make([]api.Candidacy, len(candidacies))
	for i, candidacy := range candidacies {
		uuidVal := types.UUID(candidacy.UUID)
		electionUuidVal := types.UUID(candidacy.ElectionUUID)
		userUuidVal := types.UUID(candidacy.UserUUID)
		userNamePtr := stringPtr(candidacy.UserName)

		apiCandidacies[i] = api.Candidacy{
			Uuid:         &uuidVal,
			ElectionUuid: &electionUuidVal,
			UserUuid:     &userUuidVal,
			UserName:     userNamePtr,
			CreatedAt:    &candidacy.CreatedAt,
		}
	}

	return api.GetElectionCandidacies200JSONResponse{
		Candidacies: &apiCandidacies,
	}.VisitGetElectionCandidaciesResponse(c.Response())
}

// User but only yourself
func (s *ServerHandler) PostElectionCandidacy(c echo.Context, electionUuid types.UUID) error {

	// Get current user information from session (authentication already validated by middleware)
	currentUserUUID, _ := s.GetUserFromSession(c)

	// Submit candidacy
	candidacy, err := s.candidacyService.SubmitCandidacy(uuid.UUID(electionUuid), currentUserUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.Error{
			Error:   stringPtr("internal_error"),
			Message: stringPtr("Failed to submit candidacy : " + err.Error()),
		})
	}
	fmt.Println("here")

	// Convert to API response
	uuidVal := types.UUID(candidacy.UUID)
	electionUuidVal := types.UUID(candidacy.ElectionUUID)
	userUuidVal := types.UUID(candidacy.UserUUID)
	namePtr := stringPtr("")

	// Use generated response type
	return api.PostElectionCandidacy201JSONResponse(api.Candidacy{
		Uuid:         &uuidVal,
		ElectionUuid: &electionUuidVal,
		UserUuid:     &userUuidVal,
		UserName:     namePtr,
		CreatedAt:    &candidacy.CreatedAt,
	}).VisitPostElectionCandidacyResponse(c.Response())
}

func (s *ServerHandler) PostElectionVote(c echo.Context, electionUuid types.UUID) error {
	if s.voteService == nil {
		return c.NoContent(http.StatusCreated)
	}

	// Parse request body for candidacy UUID
	var req struct {
		CandidacyUUID string `json:"candidacy_uuid"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, api.Error{
			Error:   stringPtr("invalid_request"),
			Message: stringPtr("Failed to parse request body"),
		})
	}

	candidacyUUID, err := uuid.Parse(req.CandidacyUUID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, api.Error{
			Error:   stringPtr("invalid_request"),
			Message: stringPtr("Invalid candidacy UUID format"),
		})
	}

	// Get current user information from session (authentication already validated by middleware)
	currentUserUUID, _ := s.GetUserFromSession(c)

	// Submit vote
	err = s.voteService.Vote(uuid.UUID(electionUuid), currentUserUUID, candidacyUUID)
	if err != nil {
		if err.Error() == "user has already voted in this election" {
			return c.JSON(http.StatusConflict, api.Error{
				Error:   stringPtr("conflict"),
				Message: stringPtr("User has already voted in this election"),
			})
		}
		if err.Error() == "candidacy not found" {
			return c.JSON(http.StatusNotFound, api.Error{
				Error:   stringPtr("not_found"),
				Message: stringPtr("Candidacy not found"),
			})
		}
		return c.JSON(http.StatusInternalServerError, api.Error{
			Error:   stringPtr("internal_error"),
			Message: stringPtr("Failed to submit vote : " + err.Error()),
		})
	}

	return c.NoContent(http.StatusCreated)
}

func (s *ServerHandler) GetElectionLive(c echo.Context, electionUuid types.UUID) error {

	// Check if the election is live
	election, err := s.electionService.GetElection(uuid.UUID(electionUuid))
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

	if !election.IsLiveResults {
		return c.JSON(http.StatusForbidden, api.Error{
			Error:   stringPtr("forbidden"),
			Message: stringPtr("Live results are not enabled for this election"),
		})
	}

	results, err := s.electionService.GetElectionResults(uuid.UUID(electionUuid))
	if err != nil {
		// Check for specific error cases
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
