package models

import "time"

// BookingAssignmentAttempt is the audit trail of auto-assignment offers made
// to employees for a ServiceRequest. Deliberately has NO response/status
// column — whether an attempt "won" is derived by comparing EmployeeID to
// the parent ServiceRequest.AssignedEmployeeID (see HLD §4.2/§5), never
// stored redundantly here.
type BookingAssignmentAttempt struct {
	ID               uint64         `gorm:"primaryKey"`
	ServiceRequestID uint64         `gorm:"not null;uniqueIndex:uq_assignment_attempt,priority:1"`
	ServiceRequest   ServiceRequest `gorm:"foreignKey:ServiceRequestID"`
	EmployeeID       uint64         `gorm:"not null;uniqueIndex:uq_assignment_attempt,priority:2;index:idx_assignment_attempts_employee,priority:1"`
	Employee         Employee       `gorm:"foreignKey:EmployeeID"`
	// Denormalized: employee.employer_id at the moment the offer was made.
	VendorID  uint64   `gorm:"not null;index:idx_assignment_attempts_vendor"`
	Vendor    Employer `gorm:"foreignKey:VendorID"`
	OfferedAt time.Time `gorm:"not null;index:idx_assignment_attempts_employee,priority:2"`
	RespondedAt *time.Time
}
