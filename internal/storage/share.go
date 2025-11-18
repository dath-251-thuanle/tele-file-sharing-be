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

  // Lấy hash mật khẩu của share
  GetPasswordHash(ctx context.Context, shareID int64) (string, error)

  GetShareMetadata(ctx context.Context, id int64) (*model.ShareMetadata, error)

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
	if err != nil {
		log.Printf("Failed to get share with ID %d: %v", shareID, err)
	}

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


func (r *postgresShareRepository) GetShareMetadata(ctx context.Context, id int64) (*model.ShareMetadata, error) {
    const query = `
    SELECT
        sh.id, sh.file_id, sh.owner_user_id, sh.hash, sh.require_password, sh.revoked, sh.expires_at, sh.created_at,
        f.id, f.owner_user_id, f.object_key, f.filename, f.size, f.mime, f.status, f.created_at, f.updated_at,
        u.id, u.username, u.telegram_user_id, u.created_at
    FROM shares sh
    JOIN files f ON sh.file_id = f.id
    JOIN users u ON sh.owner_user_id = u.id
    WHERE sh.id = $1
    LIMIT 1;
    `
    row := r.db.QueryRowxContext(ctx, query, id)

    var sh model.Share
    var f model.FileWithOwner
    var u model.User

    err := row.Scan(
        &sh.ID, &sh.FileID, &sh.OwnerUserID, &sh.Hash, &sh.RequirePassword, &sh.Revoked, &sh.ExpiresAt, &sh.CreatedAt,
        &f.ID, &f.OwnerUserID, &f.ObjectKey, &f.Filename, &f.Size, &f.Mime, &f.Status, &f.CreatedAt, &f.UpdatedAt,
        &u.ID, &u.Username, &u.TelegramID, &u.CreatedAt,
    )
    if err != nil {
        return nil, err
    }

    return &model.ShareMetadata{
        Share: &sh,
        File:  &f,
        User:  &u,
    }, nil
}
