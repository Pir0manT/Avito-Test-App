package models

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationType string

const (
	IE  OrganizationType = "IE"
	LLC OrganizationType = "LLC"
	JSC OrganizationType = "JSC"
)

type Organization struct {
	ID          uuid.UUID                 `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	Name        string                    `gorm:"type:varchar(100);not null" json:"name"`
	Description string                    `gorm:"type:text" json:"description"`
	Type        OrganizationType          `gorm:"type:organization_type" json:"type"`
	CreatedAt   time.Time                 `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time                 `gorm:"autoUpdateTime" json:"updatedAt"`
	Employees   []OrganizationResponsible `gorm:"foreignKey:OrganizationID;references:ID;"`
}
