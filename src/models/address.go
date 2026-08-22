package models

import "time"

// Address owner types (addresses.owner_type) — polymorphic pair with OwnerID.
const (
	AddressOwnerClient   = "client"
	AddressOwnerEmployer = "employer"
)

// Address is polymorphic (owner_type + owner_id) rather than split into
// client_addresses/employer_addresses, per the HLD's chosen tradeoff. It is
// deliberately NOT a GORM foreign key relation — the target table depends on
// OwnerType and can't be expressed as a single FK constraint.
type Address struct {
	ID        uint64  `gorm:"primaryKey"`
	OwnerType string  `gorm:"type:varchar(20);not null;index:idx_addresses_owner,priority:1;check:addresses_owner_type_chk,owner_type IN ('client','employer')"`
	OwnerID   uint64  `gorm:"not null;index:idx_addresses_owner,priority:2"`
	Line1     string  `gorm:"type:varchar(255)"`
	Line2     *string `gorm:"type:varchar(255)"`
	City      string  `gorm:"type:varchar(100);index:idx_addresses_city_pincode,priority:1"`
	State     string  `gorm:"type:varchar(100)"`
	Pincode   string  `gorm:"type:varchar(10);index:idx_addresses_city_pincode,priority:2"`
	Lat       *float64 `gorm:"type:double precision"`
	Lng       *float64 `gorm:"type:double precision"`
	CreatedAt time.Time
}
