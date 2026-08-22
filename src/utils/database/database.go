package database

import (
	"go-helpme-booking/src/config"
	"go-helpme-booking/src/models"
	"go-helpme-booking/src/utils/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	cfg := config.App.Database
	logLevel := gormlogger.Info
	if config.App.App.IsProd {
		logLevel = gormlogger.Error
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("failed to get sql.DB from gorm", zap.Error(err))
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	DB = db
	logger.Info("database connected")
}

func Migrate() {
	// Order matters: GORM creates tables (and their FK constraints) in the
	// order given, so every referenced table must appear before the table
	// that references it.
	if err := DB.AutoMigrate(
		&models.Booking{}, // existing single-sided booking table — unrelated to the marketplace schema below

		// Marketplace schema (services-marketplace-hld.docx) — identity & roster
		&models.User{},
		&models.Address{},
		&models.Category{},
		&models.Employer{},
		&models.Employee{},
		&models.Client{},

		// Category taxonomy / pricing / skills
		&models.EmployeeCategory{},
		&models.CategoryPricing{},

		// Automation (referenced by KycVerification)
		&models.AutomationJob{},

		// Booking lifecycle
		&models.ServiceRequest{},
		&models.BookingAssignmentAttempt{},
		&models.Payment{},
		&models.Review{},
		&models.EmployeeAvailability{},

		// KYC
		&models.KycVerification{},
		&models.KycDocument{},
	); err != nil {
		logger.Fatal("auto-migration failed", zap.Error(err))
	}
	logger.Info("database migration complete")
}

// Ping checks that the database connection is still alive.
func Ping() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
