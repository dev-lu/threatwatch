package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IPv4Addresses struct {
	IPAddress string    `json:"ip_address" gorm:"primaryKey;unique; not null"`
	AddedAt   time.Time `json:"added_at" gorm:"index"`
	UpdatedAt time.Time `json:"updated_at"`
	ISP       string    `json:"isp"`
	Country   string    `json:"country"`
	Region    string    `json:"region"`
	City      string    `json:"city"`
}

type Reports struct {
	ID              uuid.UUID `json:"report_id" gorm:"primaryKey;unique; not null"`
	Malicious       bool      `json:"malicious"`
	Comment         string    `json:"comment"`
	ReportingUser   string    `json:"reporting_user"`
	ReportAddedAt   time.Time `json:"added_at" gorm:"index"`
	ReportingClient string    `json:"reporting_client"`
	FKIPAddress     string    `json:"ip" gorm:"foreignKey:IPAddress;references:IPv4Addresses"`
}

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&IPv4Addresses{}, &Reports{})
	return err
}
