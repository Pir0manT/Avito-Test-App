package models

import (
	"github.com/google/uuid"
	"time"
)

type BidHistory struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	BidID       uuid.UUID `gorm:"type:uuid;not null"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Description string    `gorm:"type:varchar(500);not null"`
	Status      BidStatus `gorm:"type:bid_status;not null"`
	Version     int       `gorm:"type:int"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	Bid         Bid       `gorm:"foreignKey:BidID;references:ID;constraint:OnDelete:CASCADE"`
}
