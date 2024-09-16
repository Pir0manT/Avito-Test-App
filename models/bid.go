package models

import (
	"github.com/google/uuid"
	"time"
)

type AuthorType string

const (
	AuthorOrganization AuthorType = "Organization"
	AuthorUser         AuthorType = "User"
)

type BidStatus string

const (
	BidCreated   BidStatus = "Created"
	BidPublished BidStatus = "Published"
	BidCanceled  BidStatus = "Canceled"
)

type Bid struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Description string     `gorm:"type:varchar(500);not null" json:"description"`
	Status      BidStatus  `gorm:"type:bid_status;not null" json:"status"`
	TenderID    uuid.UUID  `gorm:"type:uuid;not null" json:"tenderId"`
	AuthorType  AuthorType `gorm:"type:author_type;not null" json:"authorType"`
	AuthorID    uuid.UUID  `gorm:"type:uuid;not null" json:"authorId"`
	Version     int        `gorm:"type:int;default:1" json:"version"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
}
