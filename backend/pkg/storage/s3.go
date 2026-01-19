package storage

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/registryx/registryx/backend/pkg/config"
)

// Driver interface abstracts the underlying storage backend.
// In the future, we could add 'Filesystem' or 'GCS' drivers.
type Driver interface {
	// Writer returns a writer to upload a blob.
	Writer(ctx context.Context, path string) (io.WriteCloser, error)
	// Reader returns a reader to download a blob.
	Reader(ctx context.Context, path string) (io.ReadCloser, error)
	// Stat returns file info.
	Stat(ctx context.Context, path string) (int64, error)
	// URLFor returns a presigned URL for direct client uploads/downloads (if supported).
	URLFor(ctx context.Context, path string, method string, expiry time.Duration) (string, error)
	// Delete removes a blob from storage.
	Delete(ctx context.Context, path string) error
}

type S3Driver struct {
	client     *minio.Client
	bucketName string
}

func NewS3Driver(cfg *config.Config) (*S3Driver, error) {
	// Initialize MinIO client object.
	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioUser, cfg.MinioPass, ""),
		Secure: cfg.MinioSecure,
	})
	if err != nil {
		return nil, err
	}

	// Ensure bucket exists
	ctx := context.Background()
	bucketName := cfg.MinioBucket
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			// already exists, proceed
		} else {
			return nil, err
		}
	}

	return &S3Driver{
		client:     minioClient,
		bucketName: bucketName,
	}, nil
}

func (d *S3Driver) Writer(ctx context.Context, path string) (io.WriteCloser, error) {
	// Create a pipe for streaming to MinIO
	r, w := io.Pipe()
	
	// Create a channel to signal when upload is complete
	done := make(chan error, 1)
	
	// Launch goroutine to upload to MinIO
	go func() {
		_, err := d.client.PutObject(ctx, d.bucketName, path, r, -1, minio.PutObjectOptions{})
		if err != nil {
			r.CloseWithError(err)
			done <- err
		} else {
			r.Close()
			done <- nil
		}
	}()
	
	// Return a wrapper that waits for upload to complete on Close()
	return &syncWriter{
		writer: w,
		done:   done,
	}, nil
}

// syncWriter wraps a pipe writer and waits for upload completion on Close()
type syncWriter struct {
	writer *io.PipeWriter
	done   chan error
}

func (sw *syncWriter) Write(p []byte) (n int, err error) {
	return sw.writer.Write(p)
}

func (sw *syncWriter) Close() error {
	// Close the writer side of the pipe
	if err := sw.writer.Close(); err != nil {
		return err
	}
	
	// Wait for the upload goroutine to complete
	return <-sw.done
}

func (d *S3Driver) Reader(ctx context.Context, path string) (io.ReadCloser, error) {
	// Check existence first so we can return an error if missing
	_, err := d.client.StatObject(ctx, d.bucketName, path, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}

	obj, err := d.client.GetObject(ctx, d.bucketName, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (d *S3Driver) Stat(ctx context.Context, path string) (int64, error) {
	info, err := d.client.StatObject(ctx, d.bucketName, path, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

func (d *S3Driver) URLFor(ctx context.Context, path string, method string, expiry time.Duration) (string, error) {
	// Generate presigned URL
	// method: "PUT" or "GET"
	
	// Implementation note: 
	// PresignedPutObject (for upload)
	// PresignedGetObject (for download)
	
	if method == "PUT" {
		u, err := d.client.PresignedPutObject(ctx, d.bucketName, path, expiry)
		if err != nil {
			return "", err
		}
		return u.String(), nil
	}
	
	// Default to GET
	u, err := d.client.PresignedGetObject(ctx, d.bucketName, path, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (d *S3Driver) Delete(ctx context.Context, path string) error {
	return d.client.RemoveObject(ctx, d.bucketName, path, minio.RemoveObjectOptions{})
}
