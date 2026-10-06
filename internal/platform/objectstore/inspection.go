package objectstore

import (
	"context"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore/s3transport"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// InspectEmpty performs only reads. In particular, an absent bucket or directory
// stays absent. Callers own the admission decision and initialization exclusion.
func InspectEmpty(ctx context.Context, settings Settings) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	switch settings.BindingKind {
	case "filesystem_root":
		return rootedfs.InspectEmpty(settings.RootPath)
	case "managed_service":
		if err := validateBucketName(settings.Bucket); err != nil {
			return false, err
		}
		connection, err := newS3Client(settings)
		if err != nil {
			return false, fmt.Errorf("inspect object storage: %w", err)
		}
		defer connection.Close()
		client := connection.Client
		ctx, cancel := context.WithTimeout(ctx, objectMetadataTimeout)
		defer cancel()
		exists, err := s3transport.BucketExists(ctx, client, settings.Bucket)
		if err != nil || !exists {
			return !exists && err == nil, err
		}
		page, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(settings.Bucket), MaxKeys: aws.Int32(1)})
		if err != nil {
			return false, err
		}
		return len(page.Contents) == 0 && !aws.ToBool(page.IsTruncated), nil
	default:
		return false, fmt.Errorf("inspect object storage: unsupported binding kind")
	}
}
