package storage

import (
	"context"
	"errors"
	"log"

	"github.com/jmoiron/sqlx"
	"file-sharing/internal/model"
)

// FileRepository interface for file operations
type FileRepository interface {
	// GetFileByID retrieves a file by its ID
	GetFileByID(ctx context.Context, fileID int64) (*model.FileWithOwner, error)

	// GetFileByIDAndUserID retrieves a file by its ID and owner user ID
	GetFileByIDAndUserID(ctx context.Context, fileID, userID int64) (*model.FileWithOwner, error)

	// UpdateFileStatus updates the status of a file
	UpdateFileStatus(ctx context.Context, fileID int64, status string) error

	// CreateUploadReport creates a new upload report
	CreateUploadReport(ctx context.Context, report *model.UploadReport) (*model.UploadReport, error)

	// GetUploadReport retrieves an upload report by file ID
	GetUploadReport(ctx context.Context, fileID int64) (*model.UploadReport, error)

	// GetUploadReportsByUserID retrieves all upload reports for a user
	GetUploadReportsByUserID(ctx context.Context, userID int64, limit, offset int) ([]model.UploadReport, error)

	// UpdateUploadReportStatus updates the status of an upload report
	UpdateUploadReportStatus(ctx context.Context, reportID int64, status string) error
}

// postgresFileRepository is the PostgreSQL implementation of FileRepository
type postgresFileRepository struct {
	db *sqlx.DB
}

// NewFileRepository creates a new file repository
func NewFileRepository(db *sqlx.DB) FileRepository {
	return &postgresFileRepository{db: db}
}


// Implementations 
// GetFileByID retrieves a file by its ID
func (r *postgresFileRepository) GetFileByID(ctx context.Context, fileID int64) (*model.FileWithOwner, error) {
	const query = `
		SELECT
			f.id,
			f.owner_user_id,
			f.object_key,
			f.filename,
			f.size,
			f.mime,
			f.status,
			f.created_at,
			f.updated_at,
			u.telegram_user_id,
			u.username
		FROM files f
		JOIN users u ON u.id = f.owner_user_id
		WHERE f.id = $1
	`

	var file model.FileWithOwner
	err := r.db.GetContext(ctx, &file, query, fileID)
	if err != nil {
		log.Printf("Failed to get file by ID %d: %v", fileID, err)
		return nil, err
	}

	return &file, nil
}

// GetFileByIDAndUserID retrieves a file by its ID and owner user ID
func (r *postgresFileRepository) GetFileByIDAndUserID(ctx context.Context, fileID, userID int64) (*model.FileWithOwner, error) {
	const query = `
		SELECT
			f.id,
			f.owner_user_id,
			f.object_key,
			f.filename,
			f.size,
			f.mime,
			f.status,
			f.created_at,
			f.updated_at,
			u.telegram_user_id,
			u.username
		FROM files f
		JOIN users u ON u.id = f.owner_user_id
		WHERE f.id = $1 AND f.owner_user_id = $2
	`

	var file model.FileWithOwner
	err := r.db.GetContext(ctx, &file, query, fileID, userID)
	if err != nil {
		log.Printf("Failed to get file by ID %d and user ID %d: %v", fileID, userID, err)
		return nil, err
	}

	return &file, nil
}

// UpdateFileStatus updates the status of a file
func (r *postgresFileRepository) UpdateFileStatus(ctx context.Context, fileID int64, status string) error {
	const query = `
		UPDATE files
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, fileID)
	if err != nil {
		log.Printf("Failed to update file status for file ID %d: %v", fileID, err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("file not found")
	}

	return nil
}

// CreateUploadReport creates a new upload report
func (r *postgresFileRepository) CreateUploadReport(ctx context.Context, report *model.UploadReport) (*model.UploadReport, error) {
	const query = `
		INSERT INTO upload_reports (
			file_id,
			owner_user_id,
			status,
			report_type,
			message,
			error_code,
			error_message,
			file_checksum,
			file_size_actual,
			upload_duration_ms,
			bandwidth_kbps,
			reported_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id, file_id, owner_user_id, status, report_type, message, error_code, error_message,
				  file_checksum, file_size_actual, upload_duration_ms, bandwidth_kbps, reported_at, updated_at
	`

	var createdReport model.UploadReport
	err := r.db.GetContext(
		ctx,
		&createdReport,
		query,
		report.FileID,
		report.OwnerUserID,
		report.Status,
		report.ReportType,
		report.Message,
		report.ErrorCode,
		report.ErrorMessage,
		report.FileChecksum,
		report.FileSizeActual,
		report.UploadDurationMs,
		report.BandwidthKbps,
	)
	if err != nil {
		log.Printf("Failed to create upload report for file ID %d: %v", report.FileID, err)
		return nil, err
	}

	return &createdReport, nil
}

// GetUploadReport retrieves an upload report by file ID
func (r *postgresFileRepository) GetUploadReport(ctx context.Context, fileID int64) (*model.UploadReport, error) {
	const query = `
		SELECT
			id,
			file_id,
			owner_user_id,
			status,
			report_type,
			message,
			error_code,
			error_message,
			file_checksum,
			file_size_actual,
			upload_duration_ms,
			bandwidth_kbps,
			reported_at,
			updated_at
		FROM upload_reports
		WHERE file_id = $1
		ORDER BY reported_at DESC
		LIMIT 1
	`

	var report model.UploadReport
	err := r.db.GetContext(ctx, &report, query, fileID)
	if err != nil {
		log.Printf("Failed to get upload report for file ID %d: %v", fileID, err)
		return nil, err
	}

	return &report, nil
}

// GetUploadReportsByUserID retrieves all upload reports for a user
func (r *postgresFileRepository) GetUploadReportsByUserID(ctx context.Context, userID int64, limit, offset int) ([]model.UploadReport, error) {
	const query = `
		SELECT
			id,
			file_id,
			owner_user_id,
			status,
			report_type,
			message,
			error_code,
			error_message,
			file_checksum,
			file_size_actual,
			upload_duration_ms,
			bandwidth_kbps,
			reported_at,
			updated_at
		FROM upload_reports
		WHERE owner_user_id = $1
		ORDER BY reported_at DESC
		LIMIT $2 OFFSET $3
	`

	var reports []model.UploadReport
	err := r.db.SelectContext(ctx, &reports, query, userID, limit, offset)
	if err != nil {
		log.Printf("Failed to get upload reports for user ID %d: %v", userID, err)
		return nil, err
	}

	return reports, nil
}

// UpdateUploadReportStatus updates the status of an upload report
func (r *postgresFileRepository) UpdateUploadReportStatus(ctx context.Context, reportID int64, status string) error {
	const query = `
		UPDATE upload_reports
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, reportID)
	if err != nil {
		log.Printf("Failed to update upload report status for report ID %d: %v", reportID, err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("upload report not found")
	}

	return nil
}
