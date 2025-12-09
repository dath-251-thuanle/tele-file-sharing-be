package storage

import (
	"context"
	"net/url"
	"time"
	"errors"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioRepo struct {
	client     *minio.Client
	bucketName string
}

func NewMinioRepo(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioRepo, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	// *** FIX: Thêm kiểm tra nếu MinIO client bị nil mà không có lỗi (Vấn đề logic) ***
	if minioClient == nil {
		return nil, errors.New("minio client initialization failed without explicit error")
	}
	// *********************************************************************************

	// Kiểm tra kết nối (Tuyệt vời nếu bạn có thể ping MinIO tại đây)
	// Ví dụ: _, err = minioClient.BucketExists(context.Background(), bucket)
    // Nếu có lỗi, bạn cũng nên trả về nil, err

	return &MinioRepo{client: minioClient, bucketName: bucket}, nil
}

// PresignObject creates a presigned GET URL for the object
func (m *MinioRepo) PresignObject(ctx context.Context, objectKey string, expirySeconds int) (string, error) {
	reqParams := make(url.Values)
	expiry := time.Duration(expirySeconds) * time.Second
	u, err := m.client.PresignedGetObject(ctx, m.bucketName, objectKey, expiry, reqParams)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// CreatePresignedPutURL tạo URL pre-signed cho việc tải lên (PUT) đối tượng.
// Object Key là tên file duy nhất trên MinIO.
// ContentType (dùng cho validation phía client)
func (m *MinioRepo) CreatePresignedPutURL(ctx context.Context, objectKey string, contentType string) (string, error) {
	// Thời hạn của URL (ví dụ: 15 phút)
	expiry := time.Minute * 15

	u, err := m.client.PresignedPutObject(ctx, m.bucketName, objectKey, expiry)
	if err != nil {
		return "", err
	}

	// Trả về URL đã tạo
	return u.String(), nil
}
