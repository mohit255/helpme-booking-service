package models

import "time"

// KYC verification source (kyc_verifications.verification_source)
const (
	KYCVerificationSourceManual        = "manual"
	KYCVerificationSourceAutomationBot = "automation_bot"
)

// KycVerification is the overall verification case for an employee; the
// supporting files live in KycDocument (1:N) so a case is no longer limited
// to a single document_type/document_number pair.
type KycVerification struct {
	ID                 uint64    `gorm:"primaryKey"`
	EmployeeID         uint64    `gorm:"not null;index:idx_kyc_employee"`
	Employee           Employee  `gorm:"foreignKey:EmployeeID"`
	// Denormalized from employees.employer_id.
	VendorID           uint64    `gorm:"not null;index:idx_kyc_vendor_status,priority:1"`
	Vendor             Employer  `gorm:"foreignKey:VendorID"`
	DocumentType       string    `gorm:"type:varchar(50)"`
	DocumentNumber     string    `gorm:"type:varchar(100)"`
	VerificationSource string    `gorm:"type:varchar(20);check:kyc_verification_source_chk,verification_source IN ('manual','automation_bot')"`
	Status             string    `gorm:"type:varchar(20);index:idx_kyc_status;index:idx_kyc_vendor_status,priority:2;check:kyc_verifications_status_chk,status IN ('pending','verified','rejected')"`
	AutomationJobID    *uint64
	AutomationJob      *AutomationJob `gorm:"foreignKey:AutomationJobID"`
	VerifiedAt         *time.Time
}

// KycDocument is one supporting file (front/back scan, secondary proof, etc.)
// submitted for a KycVerification case.
type KycDocument struct {
	ID                uint64          `gorm:"primaryKey"`
	KycVerificationID uint64          `gorm:"not null;index:idx_kyc_documents_verification"`
	KycVerification   KycVerification `gorm:"foreignKey:KycVerificationID"`
	// e.g. aadhaar_front, aadhaar_back, pan, police_verification
	DocumentType   string  `gorm:"type:varchar(50);not null"`
	DocumentNumber *string `gorm:"type:varchar(100)"`
	DocumentURL    string  `gorm:"type:text;not null"`
	UploadedAt     time.Time
}
