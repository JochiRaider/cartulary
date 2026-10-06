package objectstore

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore/s3transport"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	transfertypes "github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	s3UploadPartSize     = 8 << 20
	s3MaximumUploadParts = 10000
	s3MaximumUploadBytes = int64(s3UploadPartSize) * s3MaximumUploadParts
)

type S3Store struct {
	connection *s3transport.Connection
	client     *s3.Client
	bucket     string
}

func newS3Client(settings Settings) (*s3transport.Connection, error) {
	if !settings.Secure {
		return nil, s3transport.ErrConfiguration
	}
	return s3transport.New(s3transport.Options{Endpoint: settings.Endpoint, Region: settings.Region, AccessKey: settings.AccessKey, SecretKey: settings.SecretKey, RootCertificatePath: settings.RootCertificatePath})
}

func newS3Store(ctx context.Context, settings Settings) (Store, error) {
	if err := validateBucketName(settings.Bucket); err != nil {
		return nil, err
	}
	connection, err := newS3Client(settings)
	if err != nil {
		return nil, fmt.Errorf("create S3 client: %w", err)
	}
	store := &S3Store{connection: connection, client: connection.Client, bucket: settings.Bucket}
	exists, err := runWithRetry(ctx, OperationStartupValidation, objectMetadataTimeout, true, func(ctx context.Context) (bool, error) {
		return s3transport.BucketExists(ctx, store.client, store.bucket)
	})
	if err == nil && !exists {
		err = adapterError(OperationStartupValidation, ErrorCodeUnavailable, ReasonBucketMissing, true, "configured bucket is missing", nil)
	}
	if err == nil {
		err = store.validateStartupCapabilities(ctx)
	}
	if err != nil {
		_ = store.Close()
		return nil, err
	}
	return store, nil
}

func EnsureBucket(ctx context.Context, settings Settings) (EnsureBucketResult, error) {
	switch settings.BindingKind {
	case "filesystem_root":
		store, err := NewFilesystemStore(settings.RootPath)
		if err != nil {
			return EnsureBucketResult{}, err
		}
		_ = store.Close()
		return EnsureBucketResult{AlreadyExists: true}, nil
	case "managed_service":
		return ensureManagedServiceBucket(ctx, settings)
	default:
		return EnsureBucketResult{}, fmt.Errorf("ensure object-store bucket: unsupported binding kind")
	}
}

func ensureManagedServiceBucket(ctx context.Context, settings Settings) (EnsureBucketResult, error) {
	if err := validateBucketName(settings.Bucket); err != nil {
		return EnsureBucketResult{}, err
	}
	connection, err := newS3Client(settings)
	if err != nil {
		return EnsureBucketResult{}, err
	}
	defer connection.Close()
	store := &S3Store{connection: connection, client: connection.Client, bucket: settings.Bucket}
	result, err := store.ensureBucket(ctx, OperationEnsureBucket)
	if err == nil {
		err = store.validateStartupCapabilities(ctx)
	}
	return result, err
}

func (s *S3Store) ensureBucket(ctx context.Context, operation Operation) (EnsureBucketResult, error) {
	exists, err := runWithRetry(ctx, operation, objectMetadataTimeout, true, func(ctx context.Context) (bool, error) { return s3transport.BucketExists(ctx, s.client, s.bucket) })
	if err != nil {
		return EnsureBucketResult{}, err
	}
	if exists {
		return EnsureBucketResult{AlreadyExists: true}, nil
	}
	input := &s3.CreateBucketInput{Bucket: aws.String(s.bucket)}
	if region := s.client.Options().Region; region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{LocationConstraint: types.BucketLocationConstraint(region)}
	}
	_, err = runWithRetry(ctx, operation, objectMutationTimeout, true, func(ctx context.Context) (*s3.CreateBucketOutput, error) { return s.client.CreateBucket(ctx, input) })
	if err != nil {
		return EnsureBucketResult{}, err
	}
	return EnsureBucketResult{Created: true}, nil
}

func (s *S3Store) validateStartupCapabilities(ctx context.Context) error {
	_, err := s.Head(ctx, HeadObjectRequest{Key: ".cartulary/startup/capability-check", Purpose: PurposeProductRead})
	if err != nil && !IsObjectNotFound(err) {
		return mapBackendError(OperationStartupValidation, err)
	}
	digest := sha256.Sum256(nil)
	_, err = s.CreateUploadTarget(ctx, UploadTargetRequest{Key: ".cartulary/startup/direct-put-check", ByteSize: 0, SHA256Hex: hex.EncodeToString(digest[:]), ExpiresAt: time.Now().Add(time.Minute), Purpose: PurposeProductUpload})
	return mapBackendError(OperationStartupValidation, err)
}

func (s *S3Store) UploadTarget(context.Context, string, time.Time) (UploadTarget, error) {
	return UploadTarget{}, adapterError(OperationCreateUploadTarget, ErrorCodeInvalidRequest, ReasonInvalidRequest, false, "direct upload requires a SHA-256 checksum", nil)
}

func (s *S3Store) CreateUploadTarget(ctx context.Context, request UploadTargetRequest) (UploadTarget, error) {
	if err := validateUploadTargetRequest(request); err != nil {
		return UploadTarget{}, err
	}
	digest, err := hex.DecodeString(request.SHA256Hex)
	if err != nil || len(digest) != sha256.Size || request.ByteSize < 0 {
		return UploadTarget{}, adapterError(OperationCreateUploadTarget, ErrorCodeInvalidRequest, ReasonInvalidRequest, false, "direct upload requires a SHA-256 checksum and size", nil)
	}
	input := &s3.PutObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(request.Key), ChecksumSHA256: aws.String(base64.StdEncoding.EncodeToString(digest)), ContentLength: aws.Int64(request.ByteSize)}
	if request.ContentType != "" {
		input.ContentType = aws.String(request.ContentType)
	}
	signed, err := s3.NewPresignClient(s.client).PresignPutObject(ctx, input, func(options *s3.PresignOptions) { options.Expires = time.Until(request.ExpiresAt) })
	if err != nil {
		return UploadTarget{}, mapBackendError(OperationCreateUploadTarget, err)
	}
	headers := make(map[string]string, len(signed.SignedHeader))
	for name, values := range signed.SignedHeader {
		if strings.EqualFold(name, "Host") || strings.EqualFold(name, "Content-Length") {
			continue
		}
		headers[name] = strings.Join(values, ",")
	}
	return UploadTarget{Href: signed.URL, Method: signed.Method, Headers: headers}, nil
}

func (s *S3Store) CompleteUploadTarget(context.Context, string, io.Reader, string) error {
	return adapterError(OperationCompleteUploadTarget, ErrorCodeInvalidRequest, ReasonInvalidRequest, false, "direct upload targets must use the presigned object-store URL", nil)
}
func (s *S3Store) PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	_, err := s.Put(ctx, PutObjectRequest{Key: key, Body: body, Size: size, ContentType: contentType, Purpose: PurposeMigrationCopy})
	return err
}
func (s *S3Store) Put(ctx context.Context, request PutObjectRequest) (PutObjectResult, error) {
	if err := validatePutObjectRequest(request); err != nil {
		return PutObjectResult{}, err
	}
	if request.Size > s3MaximumUploadBytes {
		return PutObjectResult{}, invalidRequest(OperationPutObject, "object exceeds the S3 upload limit")
	}
	// Hide seeker/length hints from the SDK so it cannot enlarge part buffers.
	// The reader verifies the declared length before the SDK can commit.
	body := &s3UploadReader{reader: request.Body, remaining: s3MaximumUploadBytes, exact: request.Size >= 0}
	if body.exact {
		body.remaining = request.Size
	}
	input := &transfermanager.UploadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(request.Key), Body: body, ContentType: aws.String(request.ContentType), Metadata: map[string]string(request.Metadata), ChecksumAlgorithm: transfertypes.ChecksumAlgorithmSha256}
	manager := transfermanager.New(s.client, func(options *transfermanager.Options) {
		options.PartSizeBytes = s3UploadPartSize
		options.MultipartUploadThreshold = s3UploadPartSize
		options.Concurrency = 2
		options.MaxUploadParts = s3MaximumUploadParts
		options.FailTimeout = 15 * time.Second
		options.ChecksumAlgorithm = transfertypes.ChecksumAlgorithmSha256
		options.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	})
	output, err := runWithRetry(ctx, OperationPutObject, objectMutationTimeout, false, func(ctx context.Context) (*transfermanager.UploadObjectOutput, error) {
		return manager.UploadObject(ctx, input)
	})
	if err != nil {
		return PutObjectResult{}, err
	}
	return PutObjectResult{ETag: aws.ToString(output.ETag), SizeBytes: aws.ToInt64(output.ContentLength), ContentType: request.ContentType, Metadata: request.Metadata}, nil
}

func (s *S3Store) ReadObject(ctx context.Context, key string, options ReadOptions) (io.ReadCloser, ObjectInfo, error) {
	return s.Get(ctx, GetObjectRequest{Key: key, RangeStart: options.RangeStart, RangeEnd: options.RangeEnd, Purpose: PurposeProductRead})
}
func (s *S3Store) Get(ctx context.Context, request GetObjectRequest) (io.ReadCloser, ObjectInfo, error) {
	if err := validateGetObjectRequest(request); err != nil {
		return nil, ObjectInfo{}, err
	}
	operation := OperationGetObject
	if request.RangeStart != nil {
		operation = OperationGetObjectRange
	}
	info, err := s.Head(ctx, HeadObjectRequest{Key: request.Key, Purpose: request.Purpose})
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	if request.RangeStart != nil && *request.RangeStart >= info.Size {
		return nil, ObjectInfo{}, adapterError(operation, ErrorCodeRangeNotSatisfiable, ReasonRangeInvalid, false, "object range not satisfiable", nil)
	}
	input := &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(request.Key)}
	if request.RangeStart != nil {
		value := fmt.Sprintf("bytes=%d-", *request.RangeStart)
		if request.RangeEnd != nil {
			value += fmt.Sprint(*request.RangeEnd)
		}
		input.Range = aws.String(value)
	}
	for attempt := 1; attempt <= maxTotalAttempts; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, objectReadTimeout)
		output, err := s.client.GetObject(attemptCtx, input)
		if err == nil {
			return closeObservedStream(operation, request.Key, cancelOnCloseReadCloser{ReadCloser: output.Body, cancel: cancel}), info, nil
		}
		cancel()
		mapped := mapBackendError(operation, err)
		adapterErr, _ := AsAdapterError(mapped)
		if ctx.Err() != nil || adapterErr == nil || !adapterErr.Retryable {
			return nil, ObjectInfo{}, mapped
		}
		if attempt == maxTotalAttempts {
			return nil, ObjectInfo{}, adapterError(operation, ErrorCodeRetryExhausted, ReasonRetryExhausted, false, "retryable object-store operation exhausted attempts", mapped)
		}
		timer := time.NewTimer(retryBackoff)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return nil, ObjectInfo{}, mapBackendError(operation, ctx.Err())
		}
	}
	return nil, ObjectInfo{}, mapBackendError(operation, context.DeadlineExceeded)
}
func (s *S3Store) StatObject(ctx context.Context, key string) (ObjectInfo, error) {
	return s.Head(ctx, HeadObjectRequest{Key: key, Purpose: PurposeProductRead})
}
func (s *S3Store) Head(ctx context.Context, request HeadObjectRequest) (ObjectInfo, error) {
	if err := validateHeadObjectRequest(request); err != nil {
		return ObjectInfo{}, err
	}
	output, err := runWithRetry(ctx, OperationHeadObject, objectMetadataTimeout, true, func(ctx context.Context) (*s3.HeadObjectOutput, error) {
		return s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(request.Key)})
	})
	if err != nil {
		return ObjectInfo{}, err
	}
	return ObjectInfo{Key: request.Key, Size: aws.ToInt64(output.ContentLength), ContentType: aws.ToString(output.ContentType)}, nil
}
func (s *S3Store) ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	result, err := s.ListPrefix(ctx, ListPrefixRequest{Prefix: prefix, Purpose: PurposeTestCleanup})
	return result.Objects, err
}
func (s *S3Store) ListPrefix(ctx context.Context, request ListPrefixRequest) (ListPrefixResult, error) {
	if err := validateListPrefixRequest(request); err != nil {
		return ListPrefixResult{}, err
	}
	objects, err := runWithRetry(ctx, OperationListPrefix, objectMutationTimeout, true, func(ctx context.Context) ([]ObjectInfo, error) {
		objects := make([]ObjectInfo, 0)
		pages := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{Bucket: aws.String(s.bucket), Prefix: aws.String(request.Prefix)}, func(o *s3.ListObjectsV2PaginatorOptions) { o.StopOnDuplicateToken = true })
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			for _, item := range page.Contents {
				objects = append(objects, ObjectInfo{Key: aws.ToString(item.Key), Size: aws.ToInt64(item.Size)})
			}
		}
		return objects, nil
	})
	if err != nil {
		return ListPrefixResult{}, err
	}
	sort.Slice(objects, func(i, j int) bool { return objects[i].Key < objects[j].Key })
	return ListPrefixResult{Objects: objects}, nil
}
func (s *S3Store) DeleteObject(ctx context.Context, key string) error {
	return s.Delete(ctx, DeleteObjectRequest{Key: key, Purpose: PurposeTestCleanup})
}
func (s *S3Store) Delete(ctx context.Context, request DeleteObjectRequest) error {
	if err := validateDeleteObjectRequest(request); err != nil {
		return err
	}
	_, err := runWithRetry(ctx, OperationDeleteObject, objectMutationTimeout, true, func(ctx context.Context) (*s3.DeleteObjectOutput, error) {
		return s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(request.Key)})
	})
	return err
}
func (s *S3Store) EnsureBucketForDevTest(ctx context.Context, request EnsureBucketRequest) (EnsureBucketResult, error) {
	if err := validateEnsureBucketRequest(request); err != nil {
		return EnsureBucketResult{}, err
	}
	return s.ensureBucket(ctx, OperationEnsureBucketForDevTest)
}
func (s *S3Store) Close() error { return s.connection.Close() }
