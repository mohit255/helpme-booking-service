package models

import "time"

// Review is a client's rating of the employee for one completed ServiceRequest.
type Review struct {
	ID               uint64         `gorm:"primaryKey"`
	ServiceRequestID uint64         `gorm:"not null;uniqueIndex:uq_reviews_service_request"`
	ServiceRequest   ServiceRequest `gorm:"foreignKey:ServiceRequestID"`
	ClientID         uint64         `gorm:"not null;index:idx_reviews_client"`
	Client           Client         `gorm:"foreignKey:ClientID"`
	EmployeeID       uint64         `gorm:"not null;index:idx_reviews_employee"`
	Employee         Employee       `gorm:"foreignKey:EmployeeID"`
	Rating           int16          `gorm:"not null;check:reviews_rating_chk,rating BETWEEN 1 AND 5"`
	Comment          *string        `gorm:"type:text"`
	CreatedAt        time.Time
}
