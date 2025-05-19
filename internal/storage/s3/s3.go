package s3Storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/mahdi-vajdi/go-image-processor/internal/storage"
)

type S3Store struct {
	client *s3.Client
	bucket string
	prefix string
}

var _ storage.Storage = (*S3Store)(nil)

func NewS3Store(ctx context.Context, endpointURL string, accessKey string, secretKey string, bucket string, prefix string, region string) (storage.Storage, error) {
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("credentials for S3 are empty")
	}
	if bucket == "" {
		return nil, fmt.Errorf("s3 bucket name cannot be empty")
	}
	if region == "" {
		fmt.Println("Warning: S3 region not provided. Using 'us-east-1' as default for signing.")
		region = "us-east-1"
	}

	// Configure credentials provider and region
	credsProvider := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credsProvider),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpointURL != "" {
			o.BaseEndpoint = aws.String(endpointURL)
			o.UsePathStyle = true
			// o.HTTPClient = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
		}
	})

	return &S3Store{
		client: client,
		bucket: bucket,
		prefix: prefix,
	}, nil
}

func (s *S3Store) generateS3Key(originalFilename string) string {
	extension := filepath.Ext(originalFilename)
	base := originalFilename[:len(originalFilename)-len(extension)]
	key := fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), extension)

	if s.prefix != "" {
		// Clean the prefix path to handle potential leading/trailing slashes issues
		cleanPrefix := filepath.Clean(s.prefix)
		if cleanPrefix == "." { // If cleanPrefix is just ".", use empty string
			cleanPrefix = ""
		}
		if cleanPrefix != "" {
			// Ensure there's a slash between prefix and key
			key = cleanPrefix + "" + key
		}

		// filepath.Join uses OS separator, so we replace it
		key = filepath.ToSlash(key)

	}

	return key
}

func (s *S3Store) Save(ctx context.Context, originalFilename string, data io.Reader) (string, error) {
	key := s.generateS3Key(originalFilename)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   data,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3 bucket %s with key %s", s.bucket, key)
	}

	return key, nil
}

func (s *S3Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if the error is an S3 "Not Found" error
		var noSushKeyErr *s3Types.NoSuchKey
		if ok := errors.As(err, &noSushKeyErr); ok {
			return nil, fmt.Errorf("file not found in the S3 bucket %s with key %s: %w", s.bucket, key, os.ErrNotExist)
		}
		return nil, fmt.Errorf("failed to get file from S3 bucket %s with key %s: %w", s.bucket, key, err)
	}

	return resp.Body, nil
}

func (s *S3Store) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	// If the object doesn't exist, S3 doesn't return an error. So we don't need to check for NotFound
	if err != nil {
		return fmt.Errorf("failed to delete file from S3 bucket %s with key %s: %w", s.bucket, key, err)
	}

	return nil
}
