package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Booking struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"                        json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index"                    json:"user_id"`
	ServiceName string         `gorm:"type:varchar(150);not null"                  json:"service_name"`
	ScheduledAt time.Time      `gorm:"not null"                                    json:"scheduled_at"`
	Status      string         `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Notes       string         `gorm:"type:text"                                   json:"notes,omitempty"`
	CreatedAt   time.Time      `                                                   json:"created_at"`
	UpdatedAt   time.Time      `                                                   json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                       json:"-"`
}

func (b *Booking) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
