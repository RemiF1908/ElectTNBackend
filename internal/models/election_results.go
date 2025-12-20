package models

import (
	"github.com/google/uuid"
)

type ElectionResults struct {
	ElectionUUID  uuid.UUID                `json:"election_uuid"`
	ElectionTitle string                   `json:"election_title"`
	TotalVotes    int                      `json:"total_votes"`
	Results       []map[string]interface{} `json:"results"`
}
