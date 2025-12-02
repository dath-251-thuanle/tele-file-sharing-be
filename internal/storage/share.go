package storage

import (
	"context"
	"errors"
	"log"
	"file-sharing/internal/model"
	"github.com/jmoiron/sqlx"
)

type ShareRepository interface {
	// RevokeShare cập nhật trạng thái revoked = true
	// Check ownerUserID để đảm bảo chính chủ mới được revoke
	RevokeShare(ctx context.Context, shareID int64, ownerUserID int64) error
	
    // GetShareByID lấy thông tin share (nếu cần dùng sau này)
    GetShareByID(ctx context.Context, shareID int64) (*model.Share, error)
	

	ListSharesByOwnerUserID(ctx context.Context, ownerUserID int64, limit, offset int) ([]model.Share, error)

	// Lấy hash mật khẩu của share
    GetPasswordHash(ctx context.Context, shareID int64) (string, error)

	// Lấy đường dẫn file liên kết với share nếu share đang active
	GetActiveFilePathByShareID(ctx context.Context, shareID int64) (string, string, error)

	// Tăng số lượt tải xuống của share, trả về lỗi nếu đã đạt max_downloads
	CheckAndIncrementDownload(ctx context.Context, shareID int64) error

	// PresignObject tạo presigned URL cho objectKey, expirySeconds
	PresignObject(ctx context.Context, objectKey string, expirySeconds int) (string, error)

  GetShareMetadata(ctx context.Context, id int64) (*model.ShareMetadata, error)

}

type postgresShareRepository struct {
	db *sqlx.DB
	minio *MinioRepo
}

func NewShareRepository(db *sqlx.DB, minio *MinioRepo) ShareRepository {
	return &postgresShareRepository{db: db, minio: minio}
}

func (r *postgresShareRepository) RevokeShare(ctx context.Context, shareID int64, ownerUserID int64) error {
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

//Các hàm truy xuất DB liên quan đến authorize password 
func (r *postgresShareRepository) GetPasswordHash(ctx context.Context, shareID int64) (string, error) {
	const query = `SELECT hash_password FROM shares WHERE id = $1`
	// var passwordHash string
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

// GetActiveFilePathByShareID trả về path và filename
func (r *postgresShareRepository) GetActiveFilePathByShareID(ctx context.Context, shareID int64) (string, string, error) {
	query := `
    SELECT f.object_key, f.filename
    FROM shares s
    JOIN files f ON s.file_id = f.id
    WHERE s.id = $1
		AND NOT s.revoked
		AND (s.expires_at IS NULL OR s.expires_at > now())
    LIMIT 1
	`

	var res struct {
        Path     string `db:"object_key"`
        Filename string `db:"filename"`
    }

	err := r.db.GetContext(ctx, &res, query, shareID)
	if err != nil {
		return "", "", err
	}

    return res.Path, res.Filename, nil
}

// CheckAndIncrementDownload tăng số lượt tải xuống của share, trả về lỗi nếu đã đạt max_downloads
// HIỆN LÀ NO-OP
func (r *postgresShareRepository) CheckAndIncrementDownload(ctx context.Context, shareID int64) error {

	// CheckAndIncrementDownload tăng số lượt tải xuống của share, trả về lỗi nếu đã vượt max_downloads.
	// Đang mặc định không làm gì. 
	// Nếu muốn bật giới hạn, cần thêm trường max_downloads và download_count vào DB.

    /*
	const q = `
        UPDATE shares
        SET download_count = download_count + 1
        WHERE id = $1
        AND (max_downloads IS NULL OR download_count < max_downloads)
        RETURNING download_count
    `
	var newCount int
    err := r.db.GetContext(ctx, &newCount, q, shareID)
    if err != nil {
        if err == sql.ErrNoRows {
            return ErrMaxDownloadsExceeded
        }
        return err
    }
    _ = newCount
	*/
    return nil
}

var ErrMaxDownloadsExceeded = errors.New("max downloads exceeded")

// PresignObject tạo presigned URL cho objectKey, expirySeconds
func (r *postgresShareRepository) PresignObject(ctx context.Context, objectKey string, expirySeconds int) (string, error) {
	if r.minio == nil {
		return "", errors.New("object storage not configured")
	}
	return r.minio.PresignObject(ctx, objectKey, expirySeconds)
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
