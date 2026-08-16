package services

import (
	"time"

	"github.com/google/uuid"
	"go-helpme-booking/src/helpers"
	"go-helpme-booking/src/models"
	"go-helpme-booking/src/repositories"
)

type CreateBookingInput struct {
	ServiceName string    `json:"service_name" binding:"required" example:"Plumbing repair"`
	ScheduledAt time.Time `json:"scheduled_at" binding:"required" example:"2026-08-20T10:00:00Z"`
	Notes       string    `json:"notes"                            example:"Leaky faucet"`
}

type BookingService interface {
	Create(userID uuid.UUID, input CreateBookingInput) (*models.Booking, error)
	ListByUserID(userID uuid.UUID, page, pageSize int) ([]models.Booking, int64, error)
}

type bookingService struct {
	repo repositories.BookingRepository
}

func NewBookingService(repo repositories.BookingRepository) BookingService {
	return &bookingService{repo: repo}
}

func (s *bookingService) Create(userID uuid.UUID, input CreateBookingInput) (*models.Booking, error) {
	booking := &models.Booking{
		UserID:      userID,
		ServiceName: input.ServiceName,
		ScheduledAt: input.ScheduledAt,
		Status:      "pending",
		Notes:       input.Notes,
	}
	if err := s.repo.Create(booking); err != nil {
		return nil, err
	}
	return booking, nil
}

func (s *bookingService) ListByUserID(userID uuid.UUID, page, pageSize int) ([]models.Booking, int64, error) {
	return s.repo.FindByUserID(userID, helpers.Offset(page, pageSize), pageSize)
}
