package model

import "time"

type Share struct {
	ID          int64      `json:"id" db:"id"`
	FileID      int64      `json:"file_id" db:"file_id"`
	OwnerUserID int64      `json:"owner_user_id" db:"owner_user_id"`
	Hash        string     `json:"hash" db:"hash"`
	RequirePassword bool   `json:"require_password" db:"require_password"`
    HashPassword    string `json:"hash_password,omitempty" db:"hash_password"`
	Revoked     bool       `json:"revoked" db:"revoked"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type ShareMetadata struct {
    Share *Share
    File  *FileWithOwner
    User  *User
}

type FileMetadataDTO struct {
	ID        int64  `json:"id"`
	Filename  string `json:"filename"`
	ObjectKey string `json:"object_key"`
	Size      int64  `json:"size"`
	Mime      string `json:"mime"`
	Status    string `json:"status"`
}

type UserMinimalDTO struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	TelegramID int64  `json:"telegram_user_id"`
}

type ShareMetadataResponseDTO struct {
	ID        int64           `json:"id"`
	Hash      string          `json:"hash"`
	Revoked   bool            `json:"revoked"`
	ExpiresAt *time.Time      `json:"expires_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	File      FileMetadataDTO `json:"file"`
	Owner     UserMinimalDTO  `json:"owner"`
}