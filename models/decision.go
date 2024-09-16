package models

import (
	"github.com/google/uuid"
	"time"
)

type DecisionType string

const (
	Approved DecisionType = "Approved"
	Rejected DecisionType = "Rejected"
)

type Decision struct {
	ID           uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	BidID        uuid.UUID    `gorm:"type:uuid;not null" json:"bidId"`
	AuthorID     uuid.UUID    `gorm:"type:uuid;not null" json:"authorId"`
	DecisionType DecisionType `gorm:"type:decision_type;not null" json:"decisionType"`
	CreatedAt    time.Time    `gorm:"autoCreateTime" json:"createdAt"`
}
