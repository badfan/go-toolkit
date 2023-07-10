package blob

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
)

type IBlobClient interface {
	CreateContainer(ctx context.Context, containerName string) (*azblob.CreateContainerResponse, error)
	UploadBlob(ctx context.Context, containerName string, blobFolderPath string, blobName string, body []byte) (*azblob.UploadBufferResponse, error)
	DownloadBlob(ctx context.Context, containerName string, blobFolderPath string, blobName string) ([]byte, error)
	DeleteBlob(ctx context.Context, containerName string, blobFolderPath string, blobName string) (*azblob.DeleteBlobResponse, error)
	ListContainers(ctx context.Context) ([]*service.ContainerItem, error)
	ListBlobs(ctx context.Context, containerName string, blobFolderPath string) ([]*container.BlobItem, error)
	UploadFolder(ctx context.Context, containerName string, blobFolderPath string, localFolderPath string) ([]*azblob.UploadFileResponse, error)
}
