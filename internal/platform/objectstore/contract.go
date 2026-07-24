package objectstore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
)

const (
	maxTotalAttempts      = 2
	defaultRetryBackoff   = 100 * time.Millisecond
	uploadTargetTimeout   = 10 * time.Second
	objectMetadataTimeout = 10 * time.Second
	objectMutationTimeout = 30 * time.Second
	objectReadTimeout     = 30 * time.Second
)

var retryBackoff = defaultRetryBackoff

type Purpose string

const (
	PurposeProductUpload       Purpose = "product_upload"
	PurposeProductRead         Purpose = "product_read"
	PurposeProbeStartup        Purpose = "probe_startup"
	PurposeTestCleanup         Purpose = "test_cleanup"
	PurposeBackupManifest      Purpose = "backup_manifest"
	PurposeRestoreVerification Purpose = "restore_verification"
	PurposeMigrationCopy       Purpose = "migration_copy"
	PurposeMigrationValidation Purpose = "migration_validation"
	PurposeDiagnostic          Purpose = "diagnostic"
	PurposeStagedCleanup       Purpose = "staged_cleanup"
)

type Operation string

const (
	OperationCreateUploadTarget     Operation = "CreateUploadTarget"
	OperationCompleteUploadTarget   Operation = "CompleteUploadTarget"
	OperationPutObject              Operation = "PutObject"
	OperationHeadObject             Operation = "HeadObject"
	OperationGetObject              Operation = "GetObject"
	OperationGetObjectRange         Operation = "GetObjectRange"
	OperationListPrefix             Operation = "ListPrefix"
	OperationDeleteObject           Operation = "DeleteObject"
	OperationEnsureBucket           Operation = "EnsureBucket"
	OperationEnsureBucketForDevTest Operation = "EnsureBucketForDevTest"
	OperationStartupValidation      Operation = "StartupValidation"
)

type ErrorCode string

const (
	ErrorCodeUnavailable         ErrorCode = "object_store_unavailable"
	ErrorCodeAccessRejected      ErrorCode = "object_store_access_rejected"
	ErrorCodeObjectNotFound      ErrorCode = "object_not_found"
	ErrorCodeRangeNotSatisfiable ErrorCode = "object_range_not_satisfiable"
	ErrorCodeIntegrityMismatch   ErrorCode = "object_store_integrity_mismatch"
	ErrorCodeInvalidRequest      ErrorCode = "object_store_invalid_request"
	ErrorCodeCleanupFailed       ErrorCode = "object_store_cleanup_failed"
	ErrorCodeDeadlineExceeded    ErrorCode = "object_store_deadline_exceeded"
	ErrorCodeRetryExhausted      ErrorCode = "object_store_retry_exhausted"
)

type ReasonCode string

const (
	ReasonEndpointUnreachable ReasonCode = "endpoint_unreachable"
	ReasonBucketMissing       ReasonCode = "bucket_missing"
	ReasonCredentialDenied    ReasonCode = "credential_denied"
	ReasonCapabilityMissing   ReasonCode = "capability_missing"
	ReasonCORSRejected        ReasonCode = "cors_rejected"
	ReasonDeadlineExceeded    ReasonCode = "deadline_exceeded"
	ReasonRetryExhausted      ReasonCode = "retry_exhausted"
	ReasonInvalidRequest      ReasonCode = "invalid_request"
	ReasonObjectMissing       ReasonCode = "object_missing"
	ReasonRangeInvalid        ReasonCode = "range_not_satisfiable"
	ReasonIntegrityMismatch   ReasonCode = "integrity_mismatch"
	ReasonCleanupFailed       ReasonCode = "cleanup_failed"
)

type Metadata map[string]string

type UploadTargetRequest struct {
	Key         string
	ByteSize    int64
	ContentType string
	SHA256Hex   string
	ExpiresAt   time.Time
	Purpose     Purpose
}

type PutObjectRequest struct {
	Key         string
	Body        io.Reader
	Size        int64
	ContentType string
	Metadata    Metadata
	Purpose     Purpose
}

type HeadObjectRequest struct {
	Key     string
	Purpose Purpose
}

type GetObjectRequest struct {
	Key        string
	RangeStart *int64
	RangeEnd   *int64
	Purpose    Purpose
}

type ListPrefixRequest struct {
	Prefix            string
	ContinuationToken string
	Purpose           Purpose
}

type DeleteObjectRequest struct {
	Key     string
	Purpose Purpose
}

type EnsureBucketRequest struct {
	Profile string
	Purpose Purpose
}

type PutObjectResult struct {
	ETag        string
	SizeBytes   int64
	ContentType string
	Metadata    Metadata
}

type ListPrefixResult struct {
	Objects           []ObjectInfo
	ContinuationToken string
}

type EnsureBucketResult struct {
	Created       bool
	AlreadyExists bool
}

type TypedStore interface {
	CreateUploadTarget(ctx context.Context, request UploadTargetRequest) (UploadTarget, error)
	Put(ctx context.Context, request PutObjectRequest) (PutObjectResult, error)
	Head(ctx context.Context, request HeadObjectRequest) (ObjectInfo, error)
	Get(ctx context.Context, request GetObjectRequest) (io.ReadCloser, ObjectInfo, error)
	ListPrefix(ctx context.Context, request ListPrefixRequest) (ListPrefixResult, error)
	Delete(ctx context.Context, request DeleteObjectRequest) error
	EnsureBucketForDevTest(ctx context.Context, request EnsureBucketRequest) (EnsureBucketResult, error)
}

type AdapterError struct {
	Code      ErrorCode
	Reason    ReasonCode
	Operation Operation
	Retryable bool
	Message   string
	cause     error
}

func (e *AdapterError) Error() string {
	if e == nil {
		return ""
	}
	message := e.Message
	if message == "" {
		message = string(e.Code)
	}
	if e.Operation == "" {
		return message
	}
	return fmt.Sprintf("object store %s: %s", e.Operation, message)
}

func (e *AdapterError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func AsAdapterError(err error) (*AdapterError, bool) {
	var adapterErr *AdapterError
	if errors.As(err, &adapterErr) {
		return adapterErr, true
	}
	return nil, false
}

func IsObjectNotFound(err error) bool {
	adapterErr, ok := AsAdapterError(err)
	return ok && adapterErr.Code == ErrorCodeObjectNotFound
}

func IsDependencyError(err error) bool {
	adapterErr, ok := AsAdapterError(err)
	if !ok {
		return false
	}
	switch adapterErr.Code {
	case ErrorCodeUnavailable, ErrorCodeAccessRejected, ErrorCodeDeadlineExceeded, ErrorCodeRetryExhausted:
		return true
	default:
		return false
	}
}

func adapterError(operation Operation, code ErrorCode, reason ReasonCode, retryable bool, message string, cause error) error {
	return &AdapterError{
		Code:      code,
		Reason:    reason,
		Operation: operation,
		Retryable: retryable,
		Message:   message,
		cause:     cause,
	}
}

func invalidRequest(operation Operation, message string) error {
	return adapterError(operation, ErrorCodeInvalidRequest, ReasonInvalidRequest, false, message, nil)
}

func mapBackendError(operation Operation, err error) error {
	if err == nil {
		return nil
	}
	if adapterErr, ok := AsAdapterError(err); ok {
		if adapterErr.Operation == "" {
			adapterErr.Operation = operation
		}
		return adapterErr
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return adapterError(operation, ErrorCodeDeadlineExceeded, ReasonDeadlineExceeded, true, "operation deadline exceeded", err)
	}
	if errors.Is(err, context.Canceled) {
		return adapterError(operation, ErrorCodeUnavailable, ReasonEndpointUnreachable, true, "operation context canceled", err)
	}
	if errors.Is(err, os.ErrNotExist) || os.IsNotExist(err) {
		return adapterError(operation, ErrorCodeObjectNotFound, ReasonObjectMissing, false, "object not found", err)
	}
	if minioErr := minio.ToErrorResponse(err); minioErr.Code != "" {
		return mapMinIOErrorResponse(operation, minioErr, err)
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return adapterError(operation, ErrorCodeDeadlineExceeded, ReasonDeadlineExceeded, true, "object-store request timed out", err)
		}
		return adapterError(operation, ErrorCodeUnavailable, ReasonEndpointUnreachable, true, "object-store endpoint unreachable", err)
	}
	return adapterError(operation, ErrorCodeUnavailable, ReasonEndpointUnreachable, true, "object-store operation failed", err)
}

func mapMinIOErrorResponse(operation Operation, response minio.ErrorResponse, cause error) error {
	switch response.Code {
	case "NoSuchKey", "NoSuchObject", "NotFound":
		return adapterError(operation, ErrorCodeObjectNotFound, ReasonObjectMissing, false, "object not found", cause)
	case "NoSuchBucket":
		return adapterError(operation, ErrorCodeUnavailable, ReasonBucketMissing, true, "bucket missing", cause)
	case "AccessDenied", "InvalidAccessKeyId", "InvalidSecurity", "SignatureDoesNotMatch", "AllAccessDisabled":
		return adapterError(operation, ErrorCodeAccessRejected, ReasonCredentialDenied, false, "object-store credential denied", cause)
	case "InvalidRange", "RequestedRangeNotSatisfiable":
		return adapterError(operation, ErrorCodeRangeNotSatisfiable, ReasonRangeInvalid, false, "object range not satisfiable", cause)
	case "NotImplemented", "MethodNotAllowed", "InvalidRequest", "InvalidArgument", "XNotImplemented":
		return adapterError(operation, ErrorCodeAccessRejected, ReasonCapabilityMissing, false, "object-store capability missing", cause)
	case "BadDigest", "InvalidDigest", "ContentSHA256Mismatch":
		return adapterError(operation, ErrorCodeIntegrityMismatch, ReasonIntegrityMismatch, false, "object-store integrity mismatch", cause)
	case "RequestTimeout":
		return adapterError(operation, ErrorCodeDeadlineExceeded, ReasonDeadlineExceeded, true, "object-store request timed out", cause)
	case "InternalError", "ServiceUnavailable", "SlowDown", "TemporaryRedirect", "OperationAborted":
		return adapterError(operation, ErrorCodeUnavailable, ReasonEndpointUnreachable, true, "object-store operation unavailable", cause)
	default:
		return adapterError(operation, ErrorCodeUnavailable, ReasonEndpointUnreachable, true, "object-store operation failed", cause)
	}
}

func runWithRetry[T any](ctx context.Context, operation Operation, perAttemptTimeout time.Duration, retryAllowed bool, fn func(context.Context) (T, error)) (T, error) {
	var zero T
	attemptLimit := 1
	if retryAllowed {
		attemptLimit = maxTotalAttempts
	}
	var lastErr error
	for attempt := 1; attempt <= attemptLimit; attempt++ {
		if err := ctx.Err(); err != nil {
			return zero, mapBackendError(operation, err)
		}
		attemptCtx, cancel := context.WithTimeout(ctx, perAttemptTimeout)
		result, err := fn(attemptCtx)
		cancel()
		if err == nil {
			return result, nil
		}
		lastErr = mapBackendError(operation, err)
		adapterErr, _ := AsAdapterError(lastErr)
		if adapterErr == nil || !adapterErr.Retryable || attempt >= attemptLimit {
			break
		}
		if retryBackoff > 0 {
			timer := time.NewTimer(retryBackoff)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return zero, mapBackendError(operation, ctx.Err())
			}
		}
	}
	if retryAllowed && attemptLimit == maxTotalAttempts {
		if adapterErr, ok := AsAdapterError(lastErr); ok && adapterErr.Retryable {
			return zero, adapterError(operation, ErrorCodeRetryExhausted, ReasonRetryExhausted, false, "retryable object-store operation exhausted attempts", lastErr)
		}
	}
	return zero, lastErr
}

func SetRetryBackoffForTest(backoff time.Duration) func() {
	previous := retryBackoff
	retryBackoff = backoff
	return func() {
		retryBackoff = previous
	}
}

var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func validateUploadTargetRequest(request UploadTargetRequest) error {
	if err := validateKey(OperationCreateUploadTarget, request.Key); err != nil {
		return err
	}
	if request.Purpose != PurposeProductUpload && request.Purpose != PurposeProbeStartup {
		return invalidRequest(OperationCreateUploadTarget, "purpose is not allowed for upload target creation")
	}
	if request.ByteSize < -1 {
		return invalidRequest(OperationCreateUploadTarget, "byte size must be non-negative")
	}
	if err := validateContentType(OperationCreateUploadTarget, request.ContentType); err != nil {
		return err
	}
	if request.SHA256Hex != "" && !sha256Pattern.MatchString(request.SHA256Hex) {
		return invalidRequest(OperationCreateUploadTarget, "sha256_hex must be lowercase 64-character hexadecimal")
	}
	if !request.ExpiresAt.After(time.Now()) {
		return invalidRequest(OperationCreateUploadTarget, "expires_at must be in the future")
	}
	return nil
}

func validatePutObjectRequest(request PutObjectRequest) error {
	if err := validateKey(OperationPutObject, request.Key); err != nil {
		return err
	}
	if request.Body == nil {
		return invalidRequest(OperationPutObject, "body is required")
	}
	if request.Size < -1 {
		return invalidRequest(OperationPutObject, "size must be non-negative")
	}
	if err := validateContentType(OperationPutObject, request.ContentType); err != nil {
		return err
	}
	if err := validateMetadata(OperationPutObject, request.Metadata); err != nil {
		return err
	}
	switch request.Purpose {
	case PurposeProductUpload, PurposeProbeStartup, PurposeMigrationCopy:
		return nil
	default:
		return invalidRequest(OperationPutObject, "purpose is not allowed for put")
	}
}

func validateHeadObjectRequest(request HeadObjectRequest) error {
	if err := validateKey(OperationHeadObject, request.Key); err != nil {
		return err
	}
	switch request.Purpose {
	case PurposeProductUpload, PurposeProductRead, PurposeProbeStartup, PurposeBackupManifest, PurposeRestoreVerification, PurposeMigrationCopy, PurposeMigrationValidation, PurposeDiagnostic:
		return nil
	default:
		return invalidRequest(OperationHeadObject, "purpose is not allowed for head")
	}
}

func validateGetObjectRequest(request GetObjectRequest) error {
	operation := OperationGetObject
	if request.RangeStart != nil {
		operation = OperationGetObjectRange
	}
	if err := validateKey(operation, request.Key); err != nil {
		return err
	}
	if request.RangeStart != nil {
		start := *request.RangeStart
		if start < 0 {
			return invalidRequest(OperationGetObjectRange, "range start must be non-negative")
		}
		if request.RangeEnd != nil && start > *request.RangeEnd {
			return invalidRequest(OperationGetObjectRange, "range start must be less than or equal to range end")
		}
	}
	switch request.Purpose {
	case PurposeProductRead, PurposeProbeStartup, PurposeBackupManifest, PurposeRestoreVerification, PurposeMigrationCopy, PurposeMigrationValidation:
		return nil
	default:
		return invalidRequest(operation, "purpose is not allowed for get")
	}
}

func validateListPrefixRequest(request ListPrefixRequest) error {
	if request.Prefix != "" {
		if err := validatePrefix(OperationListPrefix, request.Prefix); err != nil {
			return err
		}
	}
	if request.ContinuationToken != "" {
		return invalidRequest(OperationListPrefix, "continuation tokens are not supported by this adapter page")
	}
	switch request.Purpose {
	case PurposeTestCleanup, PurposeBackupManifest, PurposeRestoreVerification, PurposeMigrationCopy, PurposeMigrationValidation, PurposeDiagnostic:
		return nil
	default:
		return invalidRequest(OperationListPrefix, "purpose is not allowed for list")
	}
}

func validateDeleteObjectRequest(request DeleteObjectRequest) error {
	if err := validateKey(OperationDeleteObject, request.Key); err != nil {
		return err
	}
	switch request.Purpose {
	case PurposeProbeStartup, PurposeTestCleanup, PurposeMigrationValidation, PurposeStagedCleanup:
		return nil
	default:
		return invalidRequest(OperationDeleteObject, "purpose is not allowed for delete")
	}
}

func validateEnsureBucketRequest(request EnsureBucketRequest) error {
	switch request.Profile {
	case "local_dev", "ci_service_backed", "developer_debug", "test":
		return nil
	default:
		return invalidRequest(OperationEnsureBucketForDevTest, "bucket creation is allowed only in dev/test profiles")
	}
}

func validateKey(operation Operation, key string) error {
	if key == "" || strings.ContainsRune(key, '\x00') {
		return invalidRequest(operation, "object key is required")
	}
	if strings.ContainsAny(key, "\r\n") {
		return invalidRequest(operation, "object key must not contain CR or LF")
	}
	if filepath.IsAbs(key) || strings.HasPrefix(key, "/") {
		return invalidRequest(operation, "absolute object keys are not allowed")
	}
	cleaned := filepath.Clean(filepath.FromSlash(key))
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return invalidRequest(operation, "object key escapes the object-store namespace")
	}
	return nil
}

func validatePrefix(operation Operation, prefix string) error {
	if strings.ContainsRune(prefix, '\x00') || strings.ContainsAny(prefix, "\r\n") {
		return invalidRequest(operation, "object prefix contains invalid control characters")
	}
	if strings.HasPrefix(prefix, "/") || filepath.IsAbs(prefix) {
		return invalidRequest(operation, "absolute object prefixes are not allowed")
	}
	cleaned := filepath.Clean(filepath.FromSlash(prefix))
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return invalidRequest(operation, "object prefix escapes the object-store namespace")
	}
	return nil
}

func validateBucketName(bucket string) error {
	if bucket == "" {
		return invalidRequest(OperationStartupValidation, "bucket is required")
	}
	if len(bucket) > 63 {
		return invalidRequest(OperationStartupValidation, "bucket name exceeds 63 characters")
	}
	if strings.ContainsAny(bucket, "/\\\x00\r\n") {
		return invalidRequest(OperationStartupValidation, "bucket name contains invalid characters")
	}
	return nil
}

func validateContentType(operation Operation, contentType string) error {
	if contentType == "" {
		return nil
	}
	if strings.ContainsAny(contentType, "\r\n") {
		return invalidRequest(operation, "content_type must not contain CR or LF")
	}
	if len([]byte(contentType)) > 255 {
		return invalidRequest(operation, "content_type exceeds 255 bytes")
	}
	return nil
}

var allowedMetadataKeys = map[string]struct{}{
	"cartulary-object-blob-id":         {},
	"cartulary-upload-contract-sha256": {},
	"cartulary-migration-run-id":       {},
	"cartulary-probe-id":               {},
}

func validateMetadata(operation Operation, metadata Metadata) error {
	total := 0
	for key, value := range metadata {
		if _, ok := allowedMetadataKeys[key]; !ok {
			return invalidRequest(operation, "metadata key is outside the adapter vocabulary")
		}
		if strings.ContainsAny(key, "\r\n") || strings.ContainsAny(value, "\r\n") {
			return invalidRequest(operation, "metadata must not contain CR or LF")
		}
		if len([]byte(value)) > 1024 {
			return invalidRequest(operation, "metadata value exceeds 1024 bytes")
		}
		total += len([]byte(key)) + len([]byte(value))
		if total > 8192 {
			return invalidRequest(operation, "metadata exceeds 8192 bytes")
		}
	}
	return nil
}

type StreamEvent struct {
	Event     string
	Operation Operation
	Key       string
}

var streamObserver struct {
	mu       sync.Mutex
	observer func(StreamEvent)
}

func SetStreamObserverForTest(observer func(StreamEvent)) func() {
	streamObserver.mu.Lock()
	previous := streamObserver.observer
	streamObserver.observer = observer
	streamObserver.mu.Unlock()
	return func() {
		streamObserver.mu.Lock()
		streamObserver.observer = previous
		streamObserver.mu.Unlock()
	}
}

func observeStream(event StreamEvent) {
	streamObserver.mu.Lock()
	observer := streamObserver.observer
	streamObserver.mu.Unlock()
	if observer != nil {
		observer(event)
	}
}

type observedReadCloser struct {
	io.ReadCloser
	operation Operation
	key       string
	closed    bool
}

func closeObservedStream(operation Operation, key string, reader io.ReadCloser) io.ReadCloser {
	observeStream(StreamEvent{Event: "open", Operation: operation, Key: key})
	return &observedReadCloser{ReadCloser: reader, operation: operation, key: key}
}

func (r *observedReadCloser) Close() error {
	if !r.closed {
		r.closed = true
		observeStream(StreamEvent{Event: "close", Operation: r.operation, Key: r.key})
	}
	return r.ReadCloser.Close()
}
