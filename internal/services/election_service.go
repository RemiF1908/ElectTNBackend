package services

import (
	"errors"
	"fmt"

	"elect-tn-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ElectionService struct {
	db *gorm.DB
}

func NewElectionService(db *gorm.DB) *ElectionService {
	return &ElectionService{db: db}
}

func (s *ElectionService) CreateElection(title, description string, createdByUUID uuid.UUID, isLiveResults bool) (*models.Election, error) {
	election := &models.Election{
		UUID:          uuid.New(),
		Title:         title,
		Description:   description,
		IsOpen:        false,
		IsLiveResults: isLiveResults,
		CreatedByUUID: createdByUUID,
	}

	if err := s.db.Create(election).Error; err != nil {
		return nil, fmt.Errorf("failed to create election: %w", err)
	}

	return election, nil
}

func (s *ElectionService) GetElection(electionUUID uuid.UUID) (*models.Election, error) {
	var election models.Election
	if err := s.db.Where("uuid = ?", electionUUID).First(&election).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("election not found")
		}
		return nil, fmt.Errorf("failed to get election: %w", err)
	}
	return &election, nil
}

func (s *ElectionService) ListElections() ([]*models.Election, error) {
	var elections []*models.Election
	if err := s.db.Find(&elections).Error; err != nil {
		return nil, fmt.Errorf("failed to list elections: %w", err)
	}
	return elections, nil
}

func (s *ElectionService) UpdateElection(electionUUID uuid.UUID, title, description *string, isLiveResults *bool, userUUID uuid.UUID, isAdmin bool) (*models.Election, error) {
	var election models.Election
	if err := s.db.Where("uuid = ?", electionUUID).First(&election).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("election not found")
		}
		return nil, fmt.Errorf("failed to get election: %w", err)
	}

	// Check if user is admin or creator
	if !isAdmin && election.CreatedByUUID != userUUID {
		return nil, fmt.Errorf("unauthorized: admin or creator access required")
	}

	updates := make(map[string]interface{})
	if title != nil {
		updates["title"] = *title
	}
	if description != nil {
		updates["description"] = *description
	}
	if isLiveResults != nil {
		updates["is_live_results"] = *isLiveResults
	}

	if err := s.db.Model(&election).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update election: %w", err)
	}

	// Refresh the object
	if err := s.db.Where("uuid = ?", electionUUID).First(&election).Error; err != nil {
		return nil, fmt.Errorf("failed to refresh election: %w", err)
	}

	return &election, nil
}

func (s *ElectionService) DeleteElection(electionUUID uuid.UUID, userUUID uuid.UUID, isAdmin bool) error {
	var election models.Election
	if err := s.db.Where("uuid = ?", electionUUID).First(&election).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("election not found")
		}
		return fmt.Errorf("failed to get election: %w", err)
	}

	// Check if user is admin or creator
	if !isAdmin && election.CreatedByUUID != userUUID {
		return fmt.Errorf("unauthorized: admin or creator access required")
	}

	if err := s.db.Delete(&election).Error; err != nil {
		return fmt.Errorf("failed to delete election: %w", err)
	}

	return nil
}

func (s *ElectionService) ToggleElectionOpen(electionUUID uuid.UUID, isOpen bool, userUUID uuid.UUID, isAdmin bool) (*models.Election, error) {
	var election models.Election
	if err := s.db.Where("uuid = ?", electionUUID).First(&election).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("election not found")
		}
		return nil, fmt.Errorf("failed to get election: %w", err)
	}

	// Check if user is admin or creator
	if !isAdmin && election.CreatedByUUID != userUUID {
		return nil, fmt.Errorf("unauthorized: admin or creator access required")
	}

	if err := s.db.Model(&election).Update("is_open", isOpen).Error; err != nil {
		return nil, fmt.Errorf("failed to update election status: %w", err)
	}
	election.IsOpen = isOpen

	return &election, nil
}

func (s *ElectionService) GetElectionResults(electionUUID uuid.UUID) (*models.ElectionResults, error) {
	var election models.Election
	if err := s.db.Where("uuid = ?", electionUUID).First(&election).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("election not found")
		}
		return nil, fmt.Errorf("failed to get election: %w", err)
	}
	
	// Get vote counts with candidate info
	type VoteResult struct {
		CandidacyUUID uuid.UUID `json:"candidacy_uuid"`
		CandidateName string    `json:"candidate_name"`
		VoteCount     int       `json:"vote_count"`
	}

	var results []VoteResult
	query := `
		SELECT 
			c.uuid as candidacy_uuid,
			u.name as candidate_name,
			COALESCE(vc.vote_count, 0) as vote_count
		FROM candidacies c
		JOIN users u ON c.user_uuid = u.uuid
		LEFT JOIN vote_counts vc ON c.uuid = vc.candidacy_uuid
		WHERE c.election_uuid = ?
		ORDER BY vote_count DESC, candidate_name ASC
	`

	if err := s.db.Raw(query, electionUUID).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get election results: %w", err)
	}

	// Calculate total votes
	totalVotes := 0
	for _, result := range results {
		totalVotes += result.VoteCount
	}

	// Convert to API format
	apiResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		percentage := 0.0
		if totalVotes > 0 {
			percentage = float64(result.VoteCount) / float64(totalVotes) * 100
		}

		apiResults[i] = map[string]interface{}{
			"candidacy_uuid": result.CandidacyUUID,
			"candidate_name": result.CandidateName,
			"vote_count":     result.VoteCount,
			"percentage":     percentage,
		}
	}

	electionResults := &models.ElectionResults{
		ElectionUUID:  electionUUID,
		ElectionTitle: election.Title,
		TotalVotes:    totalVotes,
		Results:       apiResults,
	}

	return electionResults, nil
}
