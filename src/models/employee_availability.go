package models

import "gorm.io/datatypes"

// EmployeeAvailability holds an employee's recurring weekly slots.
type EmployeeAvailability struct {
	ID         uint64   `gorm:"primaryKey"`
	EmployeeID uint64   `gorm:"not null;uniqueIndex:uq_employee_availability_slot,priority:1;index:idx_employee_availability_employee"`
	Employee   Employee `gorm:"foreignKey:EmployeeID"`
	// Denormalized from employees.employer_id.
	VendorID  uint64            `gorm:"not null;index:idx_employee_availability_vendor"`
	Vendor    Employer          `gorm:"foreignKey:VendorID"`
	DayOfWeek int16             `gorm:"not null;uniqueIndex:uq_employee_availability_slot,priority:2;check:employee_availability_dow_chk,day_of_week BETWEEN 0 AND 6"`
	StartTime datatypes.Time    `gorm:"not null;uniqueIndex:uq_employee_availability_slot,priority:3"`
	EndTime   datatypes.Time    `gorm:"not null"`
	IsActive  bool              `gorm:"not null;default:true"`
}

// TableName pins the name to match the HLD exactly — GORM's default
// pluralization would otherwise produce "employee_availabilities".
func (EmployeeAvailability) TableName() string { return "employee_availability" }
