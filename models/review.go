package models

import (
	"github.com/google/uuid"
	"time"
)

type Review struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	BidID       uuid.UUID `gorm:"type:uuid;not null" json:"bidId"`
	BidAuthorID uuid.UUID `gorm:"type:uuid;not null" json:"bidAuthorId"`
	Description string    `gorm:"type:varchar(1000);not null" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
}
