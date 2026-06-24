// MinIO 客户端封装：S3 兼容对象存储，用于上传/下载附件。
// 配置走环境变量：MINIO_ENDPOINT / MINIO_ACCESS_KEY / MINIO_SECRET_KEY / MINIO_BUCKET / MINIO_USE_SSL。
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

type ObjectStore struct {
	client *minio.Client
	bucket string
}

// New 从环境变量初始化；若 MINIO_ENDPOINT 为空则返回 nil（关闭对象存储功能）。
func New() (*ObjectStore, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		log.Warn().Msg("MINIO_ENDPOINT 未配置，对象存储功能禁用")
		return nil, nil
	}
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "ptis"
	}
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio 客户端初始化失败: %w", err)
	}

	// 自动创建 bucket
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("bucket 检测失败: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("创建 bucket 失败: %w", err)
		}
		log.Info().Str("bucket", bucket).Msg("MinIO bucket 已自动创建")
	}

	return &ObjectStore{client: client, bucket: bucket}, nil
}

func (s *ObjectStore) Enabled() bool { return s != nil }

// Put 上传对象。
func (s *ObjectStore) Put(ctx context.Context, objectKey string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, objectKey, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// Remove 删除对象。
func (s *ObjectStore) Remove(ctx context.Context, objectKey string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
}

// Get 下载对象内容（调用方负责 Close）。重新解析时取回原件用。
func (s *ObjectStore) Get(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
}

// PresignedGetURL 生成临时下载 URL（默认 10 分钟）。
func (s *ObjectStore) PresignedGetURL(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	u, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, ttl, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
