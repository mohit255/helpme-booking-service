package models

import (
	"time"

	"gorm.io/datatypes"
)

// ServiceRequestStatus values — see the HLD's booking state machine (§3.3):
// pending -> searching -> assigned/unassigned/cancelled -> in_progress -> completed.
const (
	ServiceRequestStatusPending     = "pending"
	ServiceRequestStatusSearching   = "searching"
	ServiceRequestStatusAssigned    = "assigned"
	ServiceRequestStatusInProgress  = "in_progress"
	ServiceRequestStatusCompleted   = "completed"
	ServiceRequestStatusCancelled   = "cancelled"
	ServiceRequestStatusUnassigned  = "unassigned"
)

// ServiceRequest is the marketplace booking entity (the HLD names the table
// "service_requests" — this is intentionally a separate model from the
// existing Booking; see docs/HLD-service-hiring-platform.md and the
// migration note in database.go for why both currently coexist).
//
// For recurring (weekly/monthly) bookings, one ServiceRequest row acts as
// the "subscription" (RecurrenceRule holds the RRULE-like schedule) — this
// HLD does not yet model the per-occurrence child rows; add a
// booking_occurrences table when that's implemented.
type ServiceRequest struct {
	ID              uint64          `gorm:"primaryKey"`
	ClientID        uint64          `gorm:"not null;index:idx_service_requests_client"`
	Client          Client          `gorm:"foreignKey:ClientID"`
	CategoryID      uint64          `gorm:"not null;index:idx_service_requests_status_category,priority:2"`
	Category        Category        `gorm:"foreignKey:CategoryID"`
	AddressID       uint64          `gorm:"not null"`
	Address         Address         `gorm:"foreignKey:AddressID"`
	BillingType     string          `gorm:"type:varchar(20);not null;check:service_requests_billing_type_chk,billing_type IN ('one_time','daily','weekly','monthly')"`
	RecurrenceRule  datatypes.JSON  `gorm:""`
	ScheduledAt     time.Time       `gorm:"not null;index:idx_service_requests_scheduled_at"`
	Status          string          `gorm:"type:varchar(20);not null;default:'pending';index:idx_service_requests_status_category,priority:1;index:idx_service_requests_employee_status,priority:2;index:idx_service_requests_vendor_status,priority:2;check:service_requests_status_chk,status IN ('pending','searching','assigned','in_progress','completed','cancelled','unassigned')"`
	AssignedEmployeeID *uint64      `gorm:"index:idx_service_requests_employee_status,priority:1"`
	AssignedEmployee   *Employee    `gorm:"foreignKey:AssignedEmployeeID"`
	// VendorID is the assigned employee's employer, set together with
	// AssignedEmployeeID; null until assigned.
	VendorID        *uint64         `gorm:"index:idx_service_requests_vendor_status,priority:1"`
	Vendor          *Employer       `gorm:"foreignKey:VendorID"`
	PriceQuoted     *float64        `gorm:"type:numeric(10,2)"`
	PriceFinal      *float64        `gorm:"type:numeric(10,2)"`
	SpecialInstructions *string     `gorm:"type:text"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
