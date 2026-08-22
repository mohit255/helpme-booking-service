package models

import "time"

// Billing types shared by category_pricing and service_requests.
const (
	BillingTypeOneTime = "one_time"
	BillingTypeDaily   = "daily"
	BillingTypeWeekly  = "weekly"
	BillingTypeMonthly = "monthly"
)

// Category is a self-referencing tree (parent_category_id) so new
// sub-categories can be added without a schema change — see the HLD's
// taxonomy in section 1.2 (Cleaning, Cooking, Shopping, Driver, Pickup & Delivery).
type Category struct {
	ID               uint64    `gorm:"primaryKey"`
	ParentCategoryID *uint64   `gorm:"index:idx_categories_parent"`
	Parent           *Category `gorm:"foreignKey:ParentCategoryID"`
	Name             string    `gorm:"type:varchar(100);not null"`
	Slug             string    `gorm:"type:varchar(100);not null;uniqueIndex:uq_categories_slug"`
	Description      *string   `gorm:"type:text"`
	IconURL          *string   `gorm:"type:text"`
	IsActive         bool      `gorm:"not null;default:true;index:idx_categories_active,where:is_active = true"`
}

// EmployeeCategory is the M:N junction between Employee and Category — this
// is what lets one worker cover multiple categories (e.g. cleaning + driving).
// It carries extra columns (experience_years, custom_rate) so it needs its
// own entity rather than a plain many2many join table.
type EmployeeCategory struct {
	EmployeeID uint64   `gorm:"primaryKey"`
	Employee   Employee `gorm:"foreignKey:EmployeeID"`
	CategoryID uint64   `gorm:"primaryKey;index:idx_employee_categories_category;index:idx_employee_categories_category_verified,priority:1"`
	Category   Category `gorm:"foreignKey:CategoryID"`

	// Denormalized from employee.employer_id — set in the same service-layer
	// call that inserts the row, not via a DB trigger.
	EmployerID uint64   `gorm:"not null;index:idx_employee_categories_employer,priority:1"`
	Employer   Employer `gorm:"foreignKey:EmployerID"`

	ExperienceYears *float64 `gorm:"type:numeric(3,1)"`
	CustomRate      *float64 `gorm:"type:numeric(10,2)"`
	IsVerifiedSkill bool     `gorm:"not null;default:false;index:idx_employee_categories_category_verified,priority:2"`
	CreatedAt       time.Time
}

// CategoryPricing resolves price by category × city × billing type, with
// effective_from/effective_to enabling price history without collisions.
type CategoryPricing struct {
	ID            uint64    `gorm:"primaryKey"`
	CategoryID    uint64    `gorm:"not null;uniqueIndex:uq_category_pricing_version,priority:1;index:idx_category_pricing_lookup,priority:1"`
	Category      Category  `gorm:"foreignKey:CategoryID"`
	City          *string   `gorm:"type:varchar(100);uniqueIndex:uq_category_pricing_version,priority:2;index:idx_category_pricing_lookup,priority:3"`
	BillingType   string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_category_pricing_version,priority:3;index:idx_category_pricing_lookup,priority:2;check:category_pricing_billing_type_chk,billing_type IN ('one_time','daily','weekly','monthly')"`
	Price         float64   `gorm:"type:numeric(10,2);not null"`
	Currency      string    `gorm:"type:char(3);not null;default:'INR'"`
	EffectiveFrom time.Time `gorm:"not null;uniqueIndex:uq_category_pricing_version,priority:4"`
	EffectiveTo   *time.Time `gorm:"index:idx_category_pricing_lookup,priority:4"`
}

// TableName pins the name to match the HLD exactly — GORM's default
// pluralization would otherwise produce "category_pricings".
func (CategoryPricing) TableName() string { return "category_pricing" }
