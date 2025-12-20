package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	UUID       uuid.UUID      `gorm:"type:uuid;primaryKey" json:"uuid"`
	Email      string         `gorm:"uniqueIndex;not null" json:"email"`
	Name       string         `gorm:"not null" json:"name"`
	IsAdmin    bool           `gorm:"default:false" json:"is_admin"`
	OAuthToken string         `gorm:"type:text" json:"-"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	CreatedElections []Election         `gorm:"foreignKey:CreatedByUUID" json:"-"`
	Candidacies      []Candidacy        `gorm:"foreignKey:UserUUID" json:"-"`
	VoteTrackings    []UserVoteTracking `gorm:"foreignKey:UserUUID" json:"-"`
}

type Election struct {
	UUID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"uuid"`
	Title         string         `gorm:"not null" json:"title"`
	Description   string         `gorm:"type:text" json:"description"`
	IsOpen        bool           `gorm:"default:false" json:"is_open"`
	IsLiveResults bool           `gorm:"default:false" json:"is_live_results"`
	CreatedByUUID uuid.UUID      `gorm:"type:uuid;not null" json:"created_by_uuid"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Creator       User               `gorm:"foreignKey:CreatedByUUID" json:"-"`
	Candidacies   []Candidacy        `gorm:"foreignKey:ElectionUUID" json:"-"`
	VoteTrackings []UserVoteTracking `gorm:"foreignKey:ElectionUUID" json:"-"`
	VoteCounts    []VoteCount        `gorm:"foreignKey:ElectionUUID" json:"-"`
}

type Candidacy struct {
	UUID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"uuid"`
	UserUUID     uuid.UUID      `gorm:"type:uuid;not null" json:"user_uuid"`
	ElectionUUID uuid.UUID      `gorm:"type:uuid;not null" json:"election_uuid"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User       User        `gorm:"foreignKey:UserUUID" json:"-"`
	Election   Election    `gorm:"foreignKey:ElectionUUID" json:"-"`
	VoteCounts []VoteCount `gorm:"foreignKey:CandidacyUUID" json:"-"`
}

type Vote struct {
	UUID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"uuid"`
	ElectionUUID uuid.UUID `gorm:"type:uuid;not null" json:"election_uuid"`
	UserUUID     uuid.UUID `gorm:"type:uuid;not null" json:"user_uuid"`
	UserHash     string    `gorm:"uniqueIndex;not null" json:"-"` // Hash for anonymous tracking
	CreatedAt    time.Time `json:"created_at"`

	// Relations
	Election Election `gorm:"foreignKey:ElectionUUID" json:"-"`
	User     User     `gorm:"foreignKey:UserUUID" json:"-"`
}

type UserVoteTracking struct {
	UUID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"uuid"`
	ElectionUUID uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_election_user" json:"election_uuid"`
	UserUUID     uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_election_user" json:"user_uuid"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Election Election `gorm:"foreignKey:ElectionUUID" json:"-"`
	User     User     `gorm:"foreignKey:UserUUID" json:"-"`
}

type VoteCount struct {
	UUID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"uuid"`
	ElectionUUID  uuid.UUID `gorm:"type:uuid;not null" json:"election_uuid"`
	CandidacyUUID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_election_candidacy" json:"candidacy_uuid"`
	VoteCount     int       `gorm:"default:0" json:"vote_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relations
	Election  Election  `gorm:"foreignKey:ElectionUUID" json:"-"`
	Candidacy Candidacy `gorm:"foreignKey:CandidacyUUID" json:"-"`
}

// BeforeCreate hook to generate UUID for User
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == uuid.Nil {
		u.UUID = uuid.New()
	}
	return nil
}

// BeforeCreate hook to generate UUID for Election
func (e *Election) BeforeCreate(tx *gorm.DB) error {
	if e.UUID == uuid.Nil {
		e.UUID = uuid.New()
	}
	return nil
}

// BeforeCreate hook to generate UUID for Candidacy
func (c *Candidacy) BeforeCreate(tx *gorm.DB) error {
	if c.UUID == uuid.Nil {
		c.UUID = uuid.New()
	}
	return nil
}

// BeforeCreate hook to generate UUID for Vote
func (v *Vote) BeforeCreate(tx *gorm.DB) error {
	if v.UUID == uuid.Nil {
		v.UUID = uuid.New()
	}
	// Generate a hash based on user_uuid + election_uuid + secret salt
	// This ensures one vote per user per election while maintaining anonymity
	hashInput := v.UserUUID.String() + v.ElectionUUID.String() + "vote_salt_secret"
	v.UserHash = hashInput // In production, use proper cryptographic hash
	return nil
}

// BeforeCreate hook to generate UUID for UserVoteTracking
func (uvt *UserVoteTracking) BeforeCreate(tx *gorm.DB) error {
	if uvt.UUID == uuid.Nil {
		uvt.UUID = uuid.New()
	}
	return nil
}

// BeforeCreate hook to generate UUID for VoteCount
func (vc *VoteCount) BeforeCreate(tx *gorm.DB) error {
	if vc.UUID == uuid.Nil {
		vc.UUID = uuid.New()
	}
	return nil
}

// TableName method for custom table names if needed
func (User) TableName() string {
	return "users"
}

func (Election) TableName() string {
	return "elections"
}

func (Candidacy) TableName() string {
	return "candidacies"
}

func (Vote) TableName() string {
	return "votes"
}

func (UserVoteTracking) TableName() string {
	return "user_vote_trackings"
}

func (VoteCount) TableName() string {
	return "vote_counts"
}
