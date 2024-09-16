package models

import (
	"github.com/google/uuid"
	"time"
)

type TenderHistory struct {
	ID          uuid.UUID   `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	TenderID    uuid.UUID   `gorm:"type:uuid;not null"`
	Name        string      `gorm:"type:varchar(100);not null"`
	Description string      `gorm:"type:varchar(500);not null"`
	ServiceType ServiceType `gorm:"type:service_type;not null"`
	Status      Status      `gorm:"type:status;not null"`
	Version     int         `gorm:"type:int"`
	CreatedAt   time.Time   `gorm:"autoCreateTime"`
	Tender      Tender      `gorm:"foreignKey:TenderID;references:ID;constraint:OnDelete:CASCADE"`
}
