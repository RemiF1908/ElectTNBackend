package services

import (
	"errors"
	"fmt"
	"time"

	"elect-tn-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CandidacyService struct {
	db *gorm.DB
}

func NewCandidacyService(db *gorm.DB) *CandidacyService {
	return &CandidacyService{db: db}
}

func (s *CandidacyService) SubmitCandidacy(electionUUID, userUUID uuid.UUID) (*models.Candidacy, error) {
	// Check if election exists and is not closed
	var election models.Election
	if err := s.db.Where("uuid = ?", electionUUID).First(&election).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("election not found")
		}
		return nil, fmt.Errorf("failed to get election: %w", err)
	}

	if election.IsOpen {
		return nil, fmt.Errorf("cannot submit candidacy to an open election")
	}

	// Check if user already has a candidacy for this election
	var existingCandidacy models.Candidacy
	if err := s.db.Where("election_uuid = ? AND user_uuid = ?", electionUUID, userUUID).First(&existingCandidacy).Error; err == nil {
		return nil, fmt.Errorf("user already has a candidacy for this election")
	}

	// Create candidacy
	candidacy := &models.Candidacy{
		ElectionUUID: electionUUID,
		UserUUID:     userUUID,
	}

	if err := s.db.Create(candidacy).Error; err != nil {
		return nil, fmt.Errorf("failed to create candidacy: %w", err)
	}

	// Initialize vote count for this candidacy
	voteCount := &models.VoteCount{
		ElectionUUID:  electionUUID,
		CandidacyUUID: candidacy.UUID,
		VoteCount:     0,
	}

	if err := s.db.Create(voteCount).Error; err != nil {
		return nil, fmt.Errorf("failed to initialize vote count: %w", err)
	}

	return candidacy, nil
}

func (s *CandidacyService) GetCandidaciesByElection(electionUUID uuid.UUID) ([]*models.Candidacy, error) {
	var candidacies []*models.Candidacy
	if err := s.db.Where("election_uuid = ?", electionUUID).Find(&candidacies).Error; err != nil {
		return nil, fmt.Errorf("failed to get candidacies: %w", err)
	}
	return candidacies, nil
}

// CandidacyWithUser includes user information for API responses
type CandidacyWithUser struct {
	UUID         uuid.UUID `json:"uuid"`
	UserUUID     uuid.UUID `json:"user_uuid"`
	ElectionUUID uuid.UUID `json:"election_uuid"`
	UserName     string    `json:"user_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s *CandidacyService) GetCandidaciesWithUserByElection(electionUUID uuid.UUID) ([]*CandidacyWithUser, error) {
	var results []*CandidacyWithUser
	query := `
		SELECT 
			c.uuid,
			c.user_uuid,
			c.election_uuid,
			u.name as user_name,
			c.created_at,
			c.updated_at
		FROM candidacies c
		JOIN users u ON c.user_uuid = u.uuid
		WHERE c.election_uuid = ?
		ORDER BY u.name ASC
	`

	if err := s.db.Raw(query, electionUUID).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get candidacies with user info: %w", err)
	}
	return results, nil
}

func (s *CandidacyService) GetCandidacyByUUID(candidacyUUID uuid.UUID) (*models.Candidacy, error) {
	var candidacy models.Candidacy
	if err := s.db.Where("uuid = ?", candidacyUUID).First(&candidacy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("candidacy not found")
		}
		return nil, fmt.Errorf("failed to get candidacy: %w", err)
	}
	return &candidacy, nil
}

func (s *CandidacyService) WithdrawCandidacy(candidacyUUID, userUUID uuid.UUID) error {
	var candidacy models.Candidacy
	if err := s.db.Where("uuid = ?", candidacyUUID).First(&candidacy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("candidacy not found")
		}
		return fmt.Errorf("failed to get candidacy: %w", err)
	}

	// Check if user owns this candidacy
	if candidacy.UserUUID != userUUID {
		return fmt.Errorf("unauthorized: can only withdraw own candidacy")
	}

	// Check if election is still closed
	var election models.Election
	if err := s.db.Where("uuid = ?", candidacy.ElectionUUID).First(&election).Error; err != nil {
		return fmt.Errorf("failed to get election: %w", err)
	}

	if election.IsOpen {
		return fmt.Errorf("cannot withdraw candidacy from an open election")
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete vote counts for this candidacy
	if err := tx.Where("candidacy_uuid = ?", candidacyUUID).Delete(&models.VoteCount{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete vote counts: %w", err)
	}

	// Delete the candidacy
	if err := tx.Delete(&candidacy).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete candidacy: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit withdrawal transaction: %w", err)
	}

	return nil
}

func (s *CandidacyService) GetUserCandidacyInElection(electionUUID, userUUID uuid.UUID) (*models.Candidacy, error) {
	var candidacy models.Candidacy
	if err := s.db.Where("election_uuid = ? AND user_uuid = ?", electionUUID, userUUID).First(&candidacy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("candidacy not found")
		}
		return nil, fmt.Errorf("failed to get candidacy: %w", err)
	}
	return &candidacy, nil
}
