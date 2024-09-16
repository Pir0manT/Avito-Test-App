package models

import (
	"github.com/google/uuid"
)

type OrganizationResponsible struct {
	ID             uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	OrganizationID uuid.UUID    `gorm:"type:uuid;not null"`
	UserID         uuid.UUID    `gorm:"type:uuid;not null"`
	Organization   Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:OnDelete:CASCADE"`
	User           Employee     `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}
