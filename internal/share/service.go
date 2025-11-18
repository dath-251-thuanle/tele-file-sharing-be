package share

import (
	"context"
	"file-sharing/internal/storage"
	"log"
	"file-sharing/internal/model"
)

type Service interface {
	RevokeShare(ctx context.Context, shareID int64, userID int64) error

	ListShares(ctx context.Context, userID int64, limit, offset int) ([]model.Share, error)

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