package services

import (
	"errors"
	"fmt"

	"elect-tn-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VoteService struct {
	db *gorm.DB
}

func NewVoteService(db *gorm.DB) *VoteService {
	return &VoteService{db: db}
}

func generateUserHash(userUUID, electionUUID uuid.UUID) string {
	// Hash function no longer needed - using direct UUID tracking
	return ""
}

func (s *VoteService) CanVote(electionUUID, userUUID uuid.UUID) error {
	var election models.Election
	if err := s.db.Where("uuid = ?", electionUUID).First(&election).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("election not found")
		}
		return fmt.Errorf("failed to get election: %w", err)
	}

	if !election.IsOpen {
		return fmt.Errorf("election is not open")
	}

	// Check if user has already voted using UserVoteTracking
	var existingTracking models.UserVoteTracking
	if err := s.db.Where("election_uuid = ? AND user_uuid = ?", electionUUID, userUUID).First(&existingTracking).Error; err == nil {
		return fmt.Errorf("user has already voted in this election")
	}

	return nil
}

func (s *VoteService) HasVoted(electionUUID, userUUID uuid.UUID) (bool, error) {
	var existingTracking models.UserVoteTracking
	if err := s.db.Where("election_uuid = ? AND user_uuid = ?", electionUUID, userUUID).First(&existingTracking).Error; err == nil {
		return true, nil
	}
	return false, nil
}

func (s *VoteService) Vote(electionUUID, userUUID, candidacyUUID uuid.UUID) error {
	// Check if candidacy exists and belongs to the election
	var candidacy models.Candidacy
	if err := s.db.Where("uuid = ? AND election_uuid = ?", candidacyUUID, electionUUID).First(&candidacy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("candidacy not found")
		}
		return fmt.Errorf("failed to get candidacy: %w", err)
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if user has already voted within transaction
	var existingTracking models.UserVoteTracking
	if err := tx.Where("election_uuid = ? AND user_uuid = ?", electionUUID, userUUID).First(&existingTracking).Error; err == nil {
		tx.Rollback()
		return fmt.Errorf("user has already voted in this election")
	}

	// Create user vote tracking record
	userVoteTracking := &models.UserVoteTracking{
		ElectionUUID: electionUUID,
		UserUUID:     userUUID,
	}

	if err := tx.Create(userVoteTracking).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create vote tracking: %w", err)
	}

	// Update vote count
	var voteCount models.VoteCount
	if err := tx.Where("election_uuid = ? AND candidacy_uuid = ?", electionUUID, candidacyUUID).First(&voteCount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new vote count record
			voteCount = models.VoteCount{
				ElectionUUID:  electionUUID,
				CandidacyUUID: candidacyUUID,
				VoteCount:     1,
			}
			if err := tx.Create(&voteCount).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create vote count: %w", err)
			}
		} else {
			tx.Rollback()
			return fmt.Errorf("failed to get vote count: %w", err)
		}
	} else {
		// Increment existing vote count
		if err := tx.Model(&voteCount).Update("vote_count", voteCount.VoteCount+1).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update vote count: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit vote transaction: %w", err)
	}

	return nil
}

func (s *VoteService) GetVoteCount(electionUUID, candidacyUUID uuid.UUID) (int, error) {
	var voteCount models.VoteCount
	if err := s.db.Where("election_uuid = ? AND candidacy_uuid = ?", electionUUID, candidacyUUID).First(&voteCount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get vote count: %w", err)
	}
	return voteCount.VoteCount, nil
}

func (s *VoteService) GetAllVoteCounts(electionUUID uuid.UUID) ([]*models.VoteCount, error) {
	var voteCounts []*models.VoteCount
	if err := s.db.Where("election_uuid = ?", electionUUID).Find(&voteCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get vote counts: %w", err)
	}
	return voteCounts, nil
}
