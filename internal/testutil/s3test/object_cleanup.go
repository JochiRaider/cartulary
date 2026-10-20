package s3test

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// The S3 DeleteObjects contract admits at most 1,000 keys per request.
const objectCleanupBatchLimit = 1000

type objectCleanupClient interface {
	ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	DeleteObjects(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
}

func cleanupObjectPrefix(ctx context.Context, client objectCleanupClient, bucket, prefix string) error {
	if strings.TrimSpace(bucket) == "" {
		return errors.New("object cleanup requires an exact bucket")
	}
	var token *string
	seen := make(map[string]struct{})
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		page, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(bucket), Prefix: aws.String(prefix),
			MaxKeys: aws.Int32(objectCleanupBatchLimit), ContinuationToken: token,
		})
		if err := ctx.Err(); err != nil {
			return err
		}
		if isNoSuchBucketError(err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("list object cleanup namespace: %w", err)
		}
		if err := validateCleanupPage(page, prefix); err != nil {
			return err
		}
		if aws.ToBool(page.IsTruncated) {
			next := aws.ToString(page.NextContinuationToken)
			if next == "" {
				return errors.New("incomplete object cleanup traversal has no continuation token")
			}
			if _, exists := seen[next]; exists {
				return errors.New("object cleanup traversal repeated a continuation token")
			}
			seen[next] = struct{}{}
			token = aws.String(next)
		}
		for start := 0; start < len(page.Contents); start += objectCleanupBatchLimit {
			end := min(start+objectCleanupBatchLimit, len(page.Contents))
			objects := make([]types.ObjectIdentifier, 0, end-start)
			for _, object := range page.Contents[start:end] {
				objects = append(objects, types.ObjectIdentifier{Key: object.Key})
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			result, err := client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(bucket), Delete: &types.Delete{Objects: objects, Quiet: aws.Bool(true)},
			})
			if err := ctx.Err(); err != nil {
				return err
			}
			if err != nil {
				return fmt.Errorf("delete object cleanup batch: %w", err)
			}
			if result == nil {
				return errors.New("object cleanup batch returned no result")
			}
			if len(result.Errors) != 0 {
				return fmt.Errorf("object cleanup batch failed for %d objects", len(result.Errors))
			}
		}
		if !aws.ToBool(page.IsTruncated) {
			break
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// A new traversal proves the postcondition independently of pagination and
	// successful delete responses. Never silently accept residue or skipped keys.
	page, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket), Prefix: aws.String(prefix), MaxKeys: aws.Int32(1),
	})
	if err := ctx.Err(); err != nil {
		return err
	}
	if isNoSuchBucketError(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("verify empty object cleanup namespace: %w", err)
	}
	if err := validateCleanupPage(page, prefix); err != nil {
		return err
	}
	if len(page.Contents) != 0 || aws.ToBool(page.IsTruncated) {
		return errors.New("object cleanup namespace is not empty")
	}
	return nil
}

func validateCleanupPage(page *s3.ListObjectsV2Output, prefix string) error {
	if page == nil || page.IsTruncated == nil || len(page.CommonPrefixes) != 0 {
		return errors.New("object cleanup received an incomplete listing")
	}
	for _, object := range page.Contents {
		key := aws.ToString(object.Key)
		if key == "" || !strings.HasPrefix(key, prefix) {
			return errors.New("object cleanup listing contains a key outside the owned namespace")
		}
	}
	return nil
}
