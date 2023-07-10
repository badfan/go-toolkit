package blob

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
)

type Blob struct {
	client  *azblob.Client
	timeout time.Duration
}

type BlobConfig struct {
	AccountName string
	AccountKey  string
}

func NewBlob(c BlobConfig, timeout time.Duration) (*Blob, error) {
	cred, err := azblob.NewSharedKeyCredential(c.AccountName, c.AccountKey)
	if err != nil {
		return nil, err
	}

	// The service URL for blob endpoints is usually in the form: http(s)://<account>.blob.core.windows.net/
	client, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", c.AccountName), cred, nil)
	if err != nil {
		return nil, err
	}

	return &Blob{
		client:  client,
		timeout: timeout,
	}, nil
}

// CreateContainer creates a container with the specified name.
func (b *Blob) CreateContainer(ctx context.Context, containerName string) (*azblob.CreateContainerResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	res, err := b.client.CreateContainer(ctx, containerName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create container : %v", err)
	}

	return &res, nil
}

// UploadBlob puts body's data into a blob in a container folder.
// BlobFolderPath must be of the following format: "exampleFolder1/exampleFolder2/". BlobFolderPath can be an empty string
func (b *Blob) UploadBlob(ctx context.Context,
	containerName string,
	blobFolderPath string,
	blobName string,
	body []byte) (*azblob.UploadBufferResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	res, err := b.client.UploadBuffer(ctx, containerName, blobFolderPath+blobName, body, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to upload blob : %v", err)
	}

	return &res, nil
}

// DownloadBlob gets a blob from a container folder and stores it in a body.
// BlobFolderPath must be of the following format: "exampleFolder1/exampleFolder2/". BlobFolderPath can be an empty string
func (b *Blob) DownloadBlob(ctx context.Context, containerName string, blobFolderPath string, blobName string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	result, err := b.client.DownloadStream(ctx, containerName, blobFolderPath+blobName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download blob : %v", err)
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to download blob : %v", err)
	}

	return body, nil
}

// DeleteBlob deletes a blob from a container folder.
// BlobFolderPath must be of the following format: "exampleFolder1/exampleFolder2/". BlobFolderPath can be an empty string
func (b *Blob) DeleteBlob(ctx context.Context, containerName string, blobFolderPath string, blobName string) (*azblob.DeleteBlobResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	res, err := b.client.DeleteBlob(ctx, containerName, blobFolderPath+blobName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to delete blob : %v", err)
	}

	return &res, nil
}

// ListContainers lists the containers in the current account.
func (b *Blob) ListContainers(ctx context.Context) ([]*service.ContainerItem, error) {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	pager := b.client.NewListContainersPager(nil)
	var res []*service.ContainerItem

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list containers : %v", err)
		}

		res = append(res, page.ContainerItems...)
	}

	return res, nil
}

// ListBlobs lists the blobs in a container folder.
// BlobFolderPath must be of the following format: "exampleFolder1/exampleFolder2/". BlobFolderPath can be an empty string
func (b *Blob) ListBlobs(ctx context.Context, containerName string, blobFolderPath string) ([]*container.BlobItem, error) {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	pager := b.client.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{Prefix: &blobFolderPath})
	var res []*container.BlobItem

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list blobs from container : %v", err)
		}

		res = append(res, page.Segment.BlobItems...)
	}

	return res, nil
}

// UploadFolder puts local folder's files in a container folder.
// BlobFolderPath must be of the following format: "exampleFolder1/exampleFolder2/". BlobFolderPath can be an empty string
func (b *Blob) UploadFolder(ctx context.Context, containerName string, blobFolderPath string, localFolderPath string) ([]*azblob.UploadFileResponse, error) {
	files, err := os.ReadDir(localFolderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read files from folder : %v", err)
	}

	var output []*azblob.UploadFileResponse
	for _, file := range files {
		f, err := os.Open(localFolderPath + "/" + file.Name())
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("failed to open file : %v", err)
		}

		res, err := b.client.UploadFile(ctx, containerName, blobFolderPath+file.Name(), f, nil)
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("failed to upload file : %v", err)
		}

		output = append(output, &res)
		f.Close()
	}

	return output, nil
}
