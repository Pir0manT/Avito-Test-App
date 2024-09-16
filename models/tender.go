package models

import (
	"github.com/google/uuid"
	"time"
)

type ServiceType string

const (
	Construction ServiceType = "Construction"
	Delivery     ServiceType = "Delivery"
	Manufacture  ServiceType = "Manufacture"
)

type Status string

const (
	Created   Status = "Created"
	Published Status = "Published"
	Closed    Status = "Closed"
)

type Tender struct {
	ID             uuid.UUID   `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	Name           string      `gorm:"type:varchar(100);not null" json:"name"`
	Description    string      `gorm:"type:varchar(500);not null" json:"description"`
	ServiceType    ServiceType `gorm:"type:service_type;not null" json:"serviceType"`
	Status         Status      `gorm:"type:status;not null" json:"status"`
	OrganizationID uuid.UUID   `gorm:"type:uuid;not null" json:"organizationId"`
	Version        int         `gorm:"type:int;default:1" json:"version"`
	CreatedAt      time.Time   `gorm:"autoCreateTime" json:"createdAt"`
}
