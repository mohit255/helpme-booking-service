package models

import (
	"time"

	"gorm.io/gorm"
)

// Employee availability (employees.availability_status)
const (
	AvailabilityStatusAvailable = "available"
	AvailabilityStatusBusy      = "busy"
	AvailabilityStatusOffline   = "offline"
)

// Genders (employees.gender)
const (
	GenderMale   = "male"
	GenderFemale = "female"
	GenderOther  = "other"
)

// Employee is the person who actually performs the work. Always belongs to
// exactly one Employer (self-employed or agency-posted) and can be skilled
// in multiple categories via EmployeeCategory.
type Employee struct {
	ID         uint64   `gorm:"primaryKey"`
	EmployerID uint64   `gorm:"not null;index:idx_employees_employer_id;index:idx_employees_employer_active,priority:1"`
	Employer   Employer `gorm:"foreignKey:EmployerID"`

	// Partial unique index keeps the "no duplicate worker phone number" rule
	// alive while still allowing a soft-deleted number to be re-onboarded.
	PhoneNumber string `gorm:"type:varchar(15);not null;uniqueIndex:uq_employees_phone_number,where:deleted_at IS NULL"`

	FullName           string  `gorm:"type:varchar(255);not null"`
	Email              *string `gorm:"type:varchar(255)"`
	DOB                *time.Time `gorm:"type:date"`
	Gender             *string `gorm:"type:varchar(10);check:employees_gender_chk,gender IN ('male','female','other')"`
	ProfilePhotoURL    *string `gorm:"type:text"`
	IsSelf             bool    `gorm:"not null;default:false"`
	AvailabilityStatus string  `gorm:"type:varchar(20);not null;default:'offline';index:idx_employees_matching,priority:1;check:employees_availability_chk,availability_status IN ('available','busy','offline')"`
	CurrentLat         *float64 `gorm:"type:double precision"`
	CurrentLng         *float64 `gorm:"type:double precision"`
	KYCStatus          string  `gorm:"type:varchar(20);not null;default:'pending';index:idx_employees_matching,priority:3;check:employees_kyc_chk,kyc_status IN ('pending','verified','rejected')"`
	RatingAvg          float64 `gorm:"type:numeric(3,2);not null;default:0"`
	TotalJobsCompleted int     `gorm:"not null;default:0"`
	IsActive           bool    `gorm:"not null;default:true;index:idx_employees_employer_active,priority:2;index:idx_employees_matching,priority:2"`
	DeletedAt          gorm.DeletedAt `gorm:"index"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
