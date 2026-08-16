package repositories

import (
	"github.com/google/uuid"
	"go-helpme-booking/src/helpers"
	"go-helpme-booking/src/models"
	"go-helpme-booking/src/utils/database"
	"go-helpme-booking/src/utils/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BookingRepository interface {
	Create(booking *models.Booking) error
	FindByID(id uuid.UUID) (*models.Booking, error)
	FindByUserID(userID uuid.UUID, offset, limit int) ([]models.Booking, int64, error)
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository() BookingRepository {
	return &bookingRepository{db: database.DB}
}

func (r *bookingRepository) Create(booking *models.Booking) (err error) {
	err = helpers.Try(func() error {
		return r.db.Create(booking).Error
	})
	if err != nil {
		logger.Error("BookingRepository.Create failed", zap.Error(err), zap.String("user_id", booking.UserID.String()))
	}
	return
}

func (r *bookingRepository) FindByID(id uuid.UUID) (booking *models.Booking, err error) {
	err = helpers.Try(func() error {
		booking = &models.Booking{}
		return r.db.First(booking, "id = ?", id).Error
	})
	if err != nil {
		logger.Debug("BookingRepository.FindByID", zap.Error(err), zap.String("id", id.String()))
		return nil, err
	}
	return
}

func (r *bookingRepository) FindByUserID(userID uuid.UUID, offset, limit int) (bookings []models.Booking, count int64, err error) {
	err = helpers.Try(func() error {
		if e := r.db.Model(&models.Booking{}).Where("user_id = ?", userID).Count(&count).Error; e != nil {
			return e
		}
		return r.db.Where("user_id = ?", userID).Offset(offset).Limit(limit).Order("scheduled_at desc").Find(&bookings).Error
	})
	if err != nil {
		logger.Error("BookingRepository.FindByUserID failed", zap.Error(err), zap.String("user_id", userID.String()))
	}
	return
}
