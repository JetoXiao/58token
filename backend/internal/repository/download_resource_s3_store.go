package repository

import (
	"context"
	"fmt"
	"mime"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// S3DownloadResourceStore keeps public download binaries in a private S3-compatible bucket.
// The application only returns short-lived presigned URLs to public callers.
type S3DownloadResourceStore struct {
	client *s3.Client
	bucket string
}

func NewS3DownloadResourceStoreFactory() service.DownloadResourceObjectStoreFactory {
	return func(ctx context.Context, cfg service.DownloadResourceS3Config) (service.DownloadResourceObjectStore, error) {
		awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(cfg.Region),
			awsconfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("load download storage AWS config: %w", err)
		}

		client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			if cfg.Endpoint != "" {
				o.BaseEndpoint = &cfg.Endpoint
			}
			o.UsePathStyle = cfg.ForcePathStyle
			o.APIOptions = append(o.APIOptions, v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware)
			o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		})

		return &S3DownloadResourceStore{client: client, bucket: cfg.Bucket}, nil
	}
}

func (s *S3DownloadResourceStore) HeadBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: &s.bucket})
	if err != nil {
		return fmt.Errorf("R2 HeadBucket: %w", err)
	}
	return nil
}

func (s *S3DownloadResourceStore) HeadObject(ctx context.Context, key string) (service.DownloadResourceObjectMetadata, error) {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: &s.bucket, Key: &key})
	if err != nil {
		return service.DownloadResourceObjectMetadata{}, fmt.Errorf("R2 HeadObject: %w", err)
	}
	metadata := service.DownloadResourceObjectMetadata{}
	if result.ContentLength != nil {
		metadata.SizeBytes = *result.ContentLength
	}
	if result.ContentType != nil {
		metadata.ContentType = *result.ContentType
	}
	if result.LastModified != nil {
		metadata.UploadedAt = *result.LastModified
	}
	return metadata, nil
}

func (s *S3DownloadResourceStore) PresignDownload(ctx context.Context, key, fileName, contentType string, expiry time.Duration) (string, error) {
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": fileName})
	presign := s3.NewPresignClient(s.client)
	result, err := presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     &s.bucket,
		Key:                        &key,
		ResponseContentDisposition: &disposition,
		ResponseContentType:        &contentType,
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("presign download URL: %w", err)
	}
	return result.URL, nil
}

func (s *S3DownloadResourceStore) PresignUpload(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	presign := s3.NewPresignClient(s.client)
	result, err := presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		ContentType: &contentType,
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("presign upload URL: %w", err)
	}
	return result.URL, nil
}
