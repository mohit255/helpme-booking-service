package models

import "time"

// Client is a customer who posts service requests. 1:1 with User.
type Client struct {
	ID                uint64   `gorm:"primaryKey"`
	UserID            uint64   `gorm:"not null;uniqueIndex:uq_clients_user_id"`
	User              User     `gorm:"foreignKey:UserID"`
	FullName          string   `gorm:"type:varchar(255);not null"`
	DefaultAddressID  *uint64
	DefaultAddress    *Address `gorm:"foreignKey:DefaultAddressID"`
	WalletBalance     float64  `gorm:"type:numeric(10,2);not null;default:0"`
	RatingAvg         float64  `gorm:"type:numeric(3,2);not null;default:0"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
