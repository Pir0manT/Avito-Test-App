package models

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID           uuid.UUID                 `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	Username     string                    `gorm:"type:varchar(50);unique;not null" json:"username"`
	FirstName    string                    `gorm:"type:varchar(50)" json:"firstName"`
	LastName     string                    `gorm:"type:varchar(50)" json:"lastName"`
	CreatedAt    time.Time                 `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time                 `gorm:"autoUpdateTime" json:"updatedAt"`
	Organization []OrganizationResponsible `gorm:"foreignKey:UserID;references:ID;"`
}
