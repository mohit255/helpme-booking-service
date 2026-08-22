package models

import "time"

// User types (users.user_type)
const (
	UserTypeClient   = "client"
	UserTypeEmployer = "employer"
	UserTypeAdmin    = "admin"
)

// User account status (users.status)
const (
	UserStatusActive    = "active"
	UserStatusSuspended = "suspended"
	UserStatusDeleted   = "deleted"
)

// User is the single auth identity for every actor in the marketplace
// (client, employer, or admin). Login is phone-OTP based; password_hash is
// only populated for accounts that also set a password.
//
// Table is named marketplace_users, not users: this same Postgres database
// already has a live "users" table (UUID id, name/email/password/role) owned
// by the existing User Service (see src/clients/user_http_client.go) — that
// table has real rows and must not be altered or collided with. The HLD's
// phone-OTP identity model is a different shape for the same concept; until
// the two are reconciled, this stays a separate table.
type User struct {
	ID              uint64 `gorm:"primaryKey"`
	PhoneNumber     string `gorm:"type:varchar(15);not null;uniqueIndex:uq_users_phone_number"`
	Email           *string `gorm:"type:varchar(255);uniqueIndex:uq_users_email,where:email IS NOT NULL"`
	PasswordHash    *string `gorm:"type:varchar(255)"`
	UserType        string `gorm:"type:varchar(20);not null;index:idx_users_user_type;check:users_user_type_chk,user_type IN ('client','employer','admin')"`
	IsPhoneVerified bool   `gorm:"not null;default:false"`
	Status          string `gorm:"type:varchar(20);not null;default:'active';check:users_status_chk,status IN ('active','suspended','deleted')"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (User) TableName() string { return "marketplace_users" }
