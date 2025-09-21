package storage

import (
	"context"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Wrapper for minio client used by service layer
type MinioClient struct {
	Client *minio.Client
	Bucket string
}

// Creation of client and ensuring bucket exists
// endpoint: localhost:9000

func NewMinioClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioClient, error) {
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	// Ensure bucket exists
	ctx := context.Background()
	err = mc.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	if err != nil {
		exists, e2 := mc.BucketExists(ctx, bucket)

		if e2 != nil {
			return nil, e2
		}
		if !exists {
			return nil, err
		}
	}
	return &MinioClient{Client: mc, Bucket: bucket}, nil
}

// Uploading the data in the minIo in the form of reader
// Return the upload info or error
func (m *MinioClient) Upload(ctx context.Context, objectKey, contentType string, reader io.Reader, size int64) (minio.UploadInfo, error) {
	info, err := m.Client.PutObject(ctx, m.Bucket, objectKey, reader, size, minio.PutObjectOptions{ContentType: contentType})

	if err != nil {
		return minio.UploadInfo{}, err
	}
	return info, nil
}

// Get object Reader -- convenience to fetch object (used by download service)
func (m *MinioClient) GetObject(ctx context.Context, objectKey string) (*minio.Object, error) {
	return m.Client.GetObject(ctx, m.Bucket, objectKey, minio.GetObjectOptions{})
}

// for debugging only for test puroses not for production
func (m *MinioClient) ListObjects(ctx context.Context) {
	log.Printf("Listing objects in bucket %s:\n", m.Bucket)
}
