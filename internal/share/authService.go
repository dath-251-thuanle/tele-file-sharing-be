package share

import (
	"context"
	"errors"
	"file-sharing/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
	"time"
	
)

type AuthServices interface{
// Hàm AuthorizeSharePassword trả về token truy cập tạm thời nếu mật khẩu đúng
	AuthorizeSharePasswordAndIssueToken(ctx context.Context, shareID int64, password string) (token string, err error)
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
}

type AuthServices interface {
	// Hàm AuthorizeSharePassword trả về token truy cập tạm thời nếu mật khẩu đúng
	AuthorizeSharePasswordAndIssueToken(ctx context.Context, shareID int64, password string) (string, error)
}

type AuthService struct {
	repo storage.ShareRepository
}

func NewAuthService(repo storage.ShareRepository) AuthServices {
	return &AuthService{
		repo: repo,
	}
}

// Triển khai các hàm liên quan đến authorize password
func CheckPasswordHash(password, hash string) bool {
	// Giả sử sử dụng bcrypt để so sánh
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateAccessToken(shareID int64) (string, error) {
	// Giả sử sử dụng JWT để tạo token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"share_id": shareID,
		"exp":      time.Now().Add(15 * time.Minute).Unix(), // Token có hạn trong 15 phút
	})
	secretKey := []byte("your-secret-key") // Thay bằng khóa bí mật thực tế
	return token.SignedString(secretKey)
}

func (a *AuthService) AuthorizeSharePasswordAndIssueToken(ctx context.Context, shareID int64, password string) (string, error) {
	// Lấy hash mật khẩu từ repository
	passwordHash, err := a.repo.GetPasswordHash(ctx, shareID)
	if err != nil {
		return "", err
	}

	// So sánh mật khẩu đã mã hóa với mật khẩu cung cấp
	if !CheckPasswordHash(password, passwordHash) {
		return "", errors.New("invalid password")
	}
	// Nếu đúng, trả về một token để truy cập tạm thời
	token, err = GenerateAccessToken(shareID)
	if err != nil {
		return "", err
	}
	return token, nil
}
