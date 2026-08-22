package models

import (
	"time"

	"gorm.io/datatypes"
)

// Automation job types (automation_jobs.job_type)
const (
	AutomationJobTypeKYCCheck   = "kyc_check"
	AutomationJobTypePriceScrape = "price_scrape"
	AutomationJobTypeReport     = "report"
)

// Automation job status (automation_jobs.status)
const (
	AutomationJobStatusQueued  = "queued"
	AutomationJobStatusRunning = "running"
	AutomationJobStatusSuccess = "success"
	AutomationJobStatusFailed  = "failed"
)

// AutomationJob is a unit of work handed to the Python/Chrome automation
// worker (KYC/ID verification, price scraping, report generation). The
// worker never writes to Postgres directly — it writes results back via an
// internal API call, so this table (and every table in this package) stays
// the single schema owned by this service.
type AutomationJob struct {
	ID               uint64  `gorm:"primaryKey"`
	JobType          string  `gorm:"type:varchar(20);not null;index:idx_automation_jobs_active_status_type,priority:3;check:automation_jobs_type_chk,job_type IN ('kyc_check','price_scrape','report')"`
	TargetEmployeeID *uint64 `gorm:"index:idx_automation_jobs_target_employee"`
	TargetEmployee   *Employee `gorm:"foreignKey:TargetEmployeeID"`
	Status           string  `gorm:"type:varchar(20);index:idx_automation_jobs_active_status_type,priority:2;check:automation_jobs_status_chk,status IN ('queued','running','success','failed')"`
	// Lets ops disable a job record without deleting its history.
	IsActive    bool           `gorm:"not null;default:true;index:idx_automation_jobs_active_status_type,priority:1"`
	Payload     datatypes.JSON `gorm:""`
	Result      datatypes.JSON `gorm:""`
	StartedAt   *time.Time
	CompletedAt *time.Time
}
