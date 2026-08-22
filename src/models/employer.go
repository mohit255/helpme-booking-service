package models

import "time"

// Employer types (employers.employer_type)
const (
	EmployerTypeSelfEmployed = "self_employed"
	EmployerTypeAgency       = "agency"
)

// KYC status shared by employers, employees and kyc_verifications
const (
	KYCStatusPending  = "pending"
	KYCStatusVerified = "verified"
	KYCStatusRejected = "rejected"
)

// Employer is either a self-employed worker (employer IS the employee) or an
// agency that manages a roster of employees. 1:1 with User.
type Employer struct {
	ID           uint64  `gorm:"primaryKey"`
	UserID       uint64  `gorm:"not null;uniqueIndex:uq_employers_user_id"`
	User         User    `gorm:"foreignKey:UserID"`
	EmployerType string  `gorm:"type:varchar(20);not null;index:idx_employers_type_kyc,priority:1;check:employers_type_chk,employer_type IN ('self_employed','agency')"`
	BusinessName *string `gorm:"type:varchar(255)"`
	GSTNumber    *string `gorm:"type:varchar(20)"`
	AddressID    *uint64
	Address      *Address `gorm:"foreignKey:AddressID"`
	KYCStatus    string   `gorm:"type:varchar(20);not null;default:'pending';index:idx_employers_type_kyc,priority:2;check:employers_kyc_chk,kyc_status IN ('pending','verified','rejected')"`
	RatingAvg    float64  `gorm:"type:numeric(3,2);not null;default:0"`
	TotalJobs    int      `gorm:"not null;default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
