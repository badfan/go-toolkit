package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type IS3Client interface {
	CreateBucket(ctx context.Context, bucketName string, bucketRegion string) (*s3.CreateBucketOutput, error)
	DeleteBucket(ctx context.Context, bucketName string) error
	UploadObject(ctx context.Context, bucketName string, objectKey string, body io.Reader) (*manager.UploadOutput, error)
	UploadFolder(ctx context.Context, bucketName string, folderPath string) ([]*manager.UploadOutput, error)
	DownloadObject(ctx context.Context, bucketName string, objectKey string, body io.WriterAt) error
	DeleteObjects(ctx context.Context, bucketName string, objectKeys []string) (*s3.DeleteObjectsOutput, error)
	ListBucketObjects(ctx context.Context, bucketName string) (*s3.ListObjectsV2Output, error)
	ListBuckets(ctx context.Context) (*s3.ListBucketsOutput, error)
}
