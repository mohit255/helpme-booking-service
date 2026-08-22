package models

import "time"

// Payment status (payments.status)
const (
	PaymentStatusPending  = "pending"
	PaymentStatusSuccess  = "success"
	PaymentStatusFailed   = "failed"
	PaymentStatusRefunded = "refunded"
)

// Payment wraps the payment gateway transaction for one ServiceRequest.
type Payment struct {
	ID               uint64         `gorm:"primaryKey"`
	ServiceRequestID uint64         `gorm:"not null;uniqueIndex:uq_payments_service_request"`
	ServiceRequest   ServiceRequest `gorm:"foreignKey:ServiceRequestID"`
	ClientID         uint64         `gorm:"not null;index:idx_payments_client"`
	Client           Client         `gorm:"foreignKey:ClientID"`
	Amount           float64        `gorm:"type:numeric(10,2);not null"`
	PaymentMethod    string         `gorm:"type:varchar(50)"`
	Status           string         `gorm:"type:varchar(20);index:idx_payments_status;check:payments_status_chk,status IN ('pending','success','failed','refunded')"`
	// Partial unique index — the idempotency guard against a webhook firing
	// twice for the same transaction, without blocking rows where no
	// transaction ref exists yet.
	GatewayTxnRef *string    `gorm:"type:varchar(255);uniqueIndex:uq_payments_gateway_txn_ref,where:gateway_txn_ref IS NOT NULL"`
	PaidAt        *time.Time
}
