package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

type S3 struct {
	client     *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
	timeout    time.Duration
}

type S3Config struct {
	Address string
	Region  string
}

func NewS3(ctx context.Context, c S3Config, timeout time.Duration) (*S3, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if c.Address != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           c.Address,
				SigningRegion: c.Region,
			}, nil
		}

		// returning EndpointNotFoundError will allow the service to fall back to its default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(c.Region),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, err
	}

	otelaws.AppendMiddlewares(&cfg.APIOptions)

	client := s3.NewFromConfig(cfg, func(opt *s3.Options) {
		opt.UsePathStyle = true
	})

	return &S3{
		client:     client,
		uploader:   manager.NewUploader(client),
		downloader: manager.NewDownloader(client),
		timeout:    timeout,
	}, nil
}

// CreateBucket creates a bucket with the specified name in the specified region.
func (s *S3) CreateBucket(ctx context.Context, bucketName string, bucketRegion string) (*s3.CreateBucketOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	res, err := s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(bucketRegion),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bucket : %v", err)
	}

	return res, nil
}

// DeleteBucket deletes a bucket. The bucket must be empty or an error is returned.
func (s *S3) DeleteBucket(ctx context.Context, bucketName string) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	_, err := s.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket : %v", err)
	}

	return nil
}

// UploadObject puts body's data into an object in a bucket.
func (s *S3) UploadObject(ctx context.Context, bucketName string, objectKey string, body io.Reader) (*manager.UploadOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	res, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   body,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload object : %v", err)
	}

	return res, nil
}

// DownloadObject gets an object from a bucket and stores it in a body.
func (s *S3) DownloadObject(ctx context.Context, bucketName string, objectKey string, body io.WriterAt) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	_, err := s.downloader.Download(ctx, body, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("failed to download object : %v", err)
	}

	return nil
}

// DeleteObjects deletes a list of objects from a bucket.
func (s *S3) DeleteObjects(ctx context.Context, bucketName string, objectKeys []string) (*s3.DeleteObjectsOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	var objectIds []types.ObjectIdentifier
	for _, key := range objectKeys {
		objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(key)})
	}

	res, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucketName),
		Delete: &types.Delete{Objects: objectIds},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete objects : %v", err)
	}

	return res, nil
}

// ListBucketObjects lists the objects in a bucket.
func (s *S3) ListBucketObjects(ctx context.Context, bucketName string) (*s3.ListObjectsV2Output, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	res, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list bucket's objects : %v", err)
	}

	return res, nil
}

// ListBuckets lists the buckets in the current account.
func (s *S3) ListBuckets(ctx context.Context) (*s3.ListBucketsOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	res, err := s.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets : %v", err)
	}

	return res, nil
}

// UploadFolder puts folder's files in a bucket.
func (s *S3) UploadFolder(ctx context.Context, bucketName string, folderPath string) ([]*manager.UploadOutput, error) {
	walker := make(fileWalk)
	if err := filepath.Walk(folderPath, walker.walk); err != nil {
		return nil, fmt.Errorf("walk failed : %v", err)
	}

	var output []*manager.UploadOutput
	for path := range walker {
		rel, err := filepath.Rel(folderPath, path)
		if err != nil {
			return nil, fmt.Errorf("failed to get relative : %v", err)
		}

		file, err := os.Open(path)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to open file : %v", err)
		}

		res, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(rel),
			Body:   file,
		})
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to upload : %v", err)
		}
		output = append(output, res)
		file.Close()
	}

	return output, nil
}

type fileWalk chan string

func (f fileWalk) walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		f <- path
	}
	return nil
}
