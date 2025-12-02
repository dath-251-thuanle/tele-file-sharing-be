package share

import (
	"context"
	"errors"
	"database/sql"
	"file-sharing/internal/storage"
	"log"
	"file-sharing/internal/model"
	"time"
)

type Service interface {
	RevokeShare(ctx context.Context, shareID int64, userID int64) error

	ListShares(ctx context.Context, userID int64, limit, offset int) ([]model.Share, error)

	// Tạo presigned URL cho download
  CreatePresignedURL(ctx context.Context, shareID int64, requesterUserID int64, expirySeconds int) (string, error)
  
	GetMetadata(ctx context.Context, id int64) (*model.ShareMetadataResponseDTO, error)
}

type shareService struct {
	repo storage.ShareRepository
}

func NewShareService(repo storage.ShareRepository) Service {
	return &shareService{
		repo: repo,
	}
}

func (s *shareService) RevokeShare(ctx context.Context, shareID int64, userID int64) error {
	return s.repo.RevokeShare(ctx, shareID, userID)
}

// ListShares lists all shares for a user
func (s *shareService) ListShares(ctx context.Context, userID int64, limit, offset int) ([]model.Share, error) {
	reports, err := s.repo.ListSharesByOwnerUserID(ctx, userID, limit, offset)
	if err != nil {
		log.Printf("Failed to list shares in service for user %d: %v", userID, err)
		return nil, err
	}

	if reports == nil {
		reports = []model.Share{} // Trả về mảng rỗng thay vì null
	}

	return reports, nil
}

// prepareDownload 
// kiểm tra điều kiện: revoked/expired, check và cập nhật số lần download
// trả về: objectKey và filename
func (s *shareService) prepareDownload(ctx context.Context, shareID int64, requesterUserID int64) (string, string, error) {
    // load share meta
    shareRec, err := s.repo.GetShareByID(ctx, shareID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return "", "", ErrShareNotFoundOrAccessDenied
        }
        return "", "", err
    }

    // check expired/revoked
    if shareRec != nil {
        if shareRec.Revoked {
            return "", "", ErrShareRevokedOrExpired
        }
        if shareRec.ExpiresAt != nil && shareRec.ExpiresAt.Before(time.Now()) {
            return "", "", ErrShareRevokedOrExpired
        }
    }

    // lấy object key + filename
    objectKey, filename, err := s.repo.GetActiveFilePathByShareID(ctx, shareID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return "", "", ErrShareNotFoundOrAccessDenied
        }
        return "", "", err
    }

	// check + increment download counter (Hiện là no-op)
	err = s.repo.CheckAndIncrementDownload(ctx, shareID); 
	if err != nil {
		return "", "", ErrMaxDownloadsExceeded
    }

    return objectKey, filename, nil
}

func (s *shareService) CreatePresignedURL(ctx context.Context, shareID int64, requesterUserID int64, expirySeconds int) (string, error) {
    // prepareDownload kiểm tra các điều kiện và trả về object key + filename + lỗi
    objectKey, _, err := s.prepareDownload(ctx, shareID, requesterUserID)
    if err != nil {
        return "", err
    }

    // presign via repo (MinIO)
    url, err := s.repo.PresignObject(ctx, objectKey, expirySeconds)
    if err != nil {
        return "", err
    }
    return url, nil
}

var (
    // trả về khi share không tồn tại hoặc không được phép truy cập
    ErrShareNotFoundOrAccessDenied = errors.New("share not found or access denied")
    // trả về khi share đã bị revoke hoặc đã hết hạn
    ErrShareRevokedOrExpired = errors.New("share revoked or expired")
	// trả khi đã đạt max_downloads
    ErrMaxDownloadsExceeded = errors.New("max downloads exceeded")
)

func (s *shareService) GetMetadata(ctx context.Context, id int64) (*model.ShareMetadataResponseDTO, error) {
	md, err := s.repo.GetShareMetadata(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.ShareMetadataResponseDTO{
		ID:        md.Share.ID,
		Hash:      md.Share.Hash,
		Revoked:   md.Share.Revoked,
		ExpiresAt: md.Share.ExpiresAt,
		CreatedAt: md.Share.CreatedAt,
		File: model.FileMetadataDTO{
			ID:        md.File.ID,
			Filename:  md.File.Filename,
			ObjectKey: md.File.ObjectKey,
			Size:      md.File.Size,
			Mime:      md.File.Mime,
			Status:    md.File.Status,
		},
		Owner: model.UserMinimalDTO{
			ID:         md.User.ID,
			Username:   md.User.Username,
			TelegramID: md.User.TelegramID,
		},
	}, nil
}
