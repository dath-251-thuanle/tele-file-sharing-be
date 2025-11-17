package storage

import (
	"context"
	"errors"
	"log"

	"file-sharing/internal/model"
	"github.com/jmoiron/sqlx"
)

// Context tự động thực thi truy vấn và ánh xạ vào Struct phù hợp 
type ShareRepository interface {
	// RevokeShare cập nhật trạng thái revoked = true
	// Check ownerUserID để đảm bảo chính chủ mới được revoke
	RevokeShare(ctx context.Context, shareID int64, ownerUserID int64) error
	
    // GetShareByID lấy thông tin share (nếu cần dùng sau này)
    GetShareByID(ctx context.Context, shareID int64) (*model.Share, error)

	ListSharesByOwnerUserID(ctx context.Context, ownerUserID int64, limit, offset int) ([]model.Share, error)

	GetPasswordHash(ctx context.Context, shareID int64) (string, error)
}

type postgresShareRepository struct {
	db *sqlx.DB
}

func NewShareRepository(db *sqlx.DB) ShareRepository {
	return &postgresShareRepository{db: db}
}

func (r *postgresShareRepository) RevokeShare(ctx context.Context, shareID int64, ownerUserID int64) error {
	// Original state set to False: not revoked yet 
	const query = `
		UPDATE shares
		SET revoked = true, updated_at = NOW()
		WHERE id = $1 AND owner_user_id = $2 
	`

	result, err := r.db.ExecContext(ctx, query, shareID, ownerUserID)
	if err != nil {
		log.Printf("Failed to revoke share ID %d: %v", shareID, err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("share not found or access denied")
	}

	return nil
}

func (r *postgresShareRepository) GetShareByID(ctx context.Context, shareID int64) (*model.Share, error) {
    const query = `SELECT * FROM shares WHERE id = $1`
    var share model.Share
    err := r.db.GetContext(ctx, &share, query, shareID)
    return &share, err
}

// Các hàm truy xuất DB liên quan đến authorize password 
func (r *postgresShareRepository) GetPasswordHash(ctx context.Context, shareID int64) (string, error) {
	const query = `SELECT hash_password FROM shares WHERE id = $1`
	var passwordHash string
	err := r.db.GetContext(ctx, &passwordHash, query, shareID)
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}


// ListSharesByOwnerUserID retrieves all shares for a user with pagination
func (r *postgresShareRepository) ListSharesByOwnerUserID(ctx context.Context, ownerUserID int64, limit, offset int) ([]model.Share, error) {
	const query = `
		SELECT
			id,
			file_id,
			owner_user_id,
			hash,
			require_password,
			revoked,
			expires_at,
			created_at,
			updated_at
		FROM shares
		WHERE owner_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	var shares []model.Share
	err := r.db.SelectContext(ctx, &shares, query, ownerUserID, limit, offset)
	if err != nil {
		log.Printf("Failed to list shares for user ID %d: %v", ownerUserID, err)
		return nil, err
	}

	return shares, nil
}