package objectstore

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/JochiRaider/cartulary/internal/platform/config"
)

const (
	EndpointEnv  = "CARTULARY_S3_ENDPOINT"
	AccessKeyEnv = "CARTULARY_S3_ACCESS_KEY_ID"
	SecretKeyEnv = "CARTULARY_S3_SECRET_ACCESS_KEY"
	SecureEnv    = "CARTULARY_S3_SECURE"
	BucketEnv    = "CARTULARY_S3_BUCKET"
)

type Settings struct {
	BindingKind string
	RootPath    string
	Endpoint    string
	AccessKey   string
	SecretKey   string
	Secure      bool
	Bucket      string
}

type ServiceRefEnvKeys struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    string
	Bucket    string
}

type Store interface {
	UploadTarget(ctx context.Context, key string, expiresAt time.Time) (UploadTarget, error)
	CompleteUploadTarget(ctx context.Context, token string, body io.Reader, contentType string) error
	PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	ReadObject(ctx context.Context, key string, options ReadOptions) (io.ReadCloser, ObjectInfo, error)
	StatObject(ctx context.Context, key string) (ObjectInfo, error)
	ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error)
	DeleteObject(ctx context.Context, key string) error
	Close() error
}

type UploadTarget struct {
	Href    string
	Method  string
	Headers map[string]string
}

type ReadOptions struct {
	RangeStart *int64
	RangeEnd   *int64
}

type ObjectInfo struct {
	Key         string
	Size        int64
	ContentType string
}

func Setup(ctx context.Context, cfg config.Config) (Store, error) {
	return SetupWithEnv(ctx, cfg, nil)
}

func SetupWithEnv(ctx context.Context, cfg config.Config, env map[string]string) (Store, error) {
	settings, err := ResolveSettings(cfg, env)
	if err != nil {
		return nil, err
	}

	switch settings.BindingKind {
	case "filesystem_root":
		store, err := NewFilesystemStore(settings.RootPath)
		if err != nil {
			return nil, err
		}
		if cfg.Telemetry.Enabled {
			return InstrumentStore(store, cfg.Telemetry.Resource.ServiceVersion), nil
		}
		return store, nil
	case "managed_service":
		store, err := newS3Store(ctx, settings)
		if err != nil {
			return nil, err
		}
		if cfg.Telemetry.Enabled {
			return InstrumentStore(store, cfg.Telemetry.Resource.ServiceVersion), nil
		}
		return store, nil
	default:
		return nil, fmt.Errorf("setup object store: unsupported binding_kind %q", settings.BindingKind)
	}
}

func ResolveSettings(cfg config.Config, env map[string]string) (Settings, error) {
	switch cfg.Roots.ObjectStorage.BindingKind {
	case "filesystem_root":
		if cfg.Roots.ObjectStorage.Path == "" {
			return Settings{}, fmt.Errorf("resolve object-store settings: roots.object_storage.path is required")
		}
		return Settings{BindingKind: "filesystem_root", RootPath: cfg.Roots.ObjectStorage.Path}, nil
	case "managed_service":
		settings, err := resolveManagedServiceSettings(cfg.Roots.ObjectStorage.ServiceRef, env)
		if err != nil {
			return Settings{}, err
		}
		settings.BindingKind = "managed_service"
		return settings, nil
	default:
		return Settings{}, fmt.Errorf("resolve object-store settings: roots.object_storage.binding_kind must be configured before object-store setup")
	}
}

func newS3Store(ctx context.Context, settings Settings) (Store, error) {
	if err := validateBucketName(settings.Bucket); err != nil {
		return nil, err
	}
	client, err := minio.New(settings.Endpoint, &minio.Options{
		Creds:      credentials.NewStaticV4(settings.AccessKey, settings.SecretKey, ""),
		Secure:     settings.Secure,
		MaxRetries: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("create S3 client: %w", err)
	}

	store := &S3Store{client: client, bucket: settings.Bucket}
	exists, err := runWithRetry(ctx, OperationStartupValidation, objectMetadataTimeout, true, func(ctx context.Context) (bool, error) {
		return client.BucketExists(ctx, settings.Bucket)
	})
	if err != nil {
		return nil, fmt.Errorf("validate object-store bucket: %w", err)
	}
	if !exists {
		return nil, adapterError(OperationStartupValidation, ErrorCodeUnavailable, ReasonBucketMissing, true, "configured bucket is missing", nil)
	}
	if err := store.validateStartupCapabilities(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *S3Store) validateStartupCapabilities(ctx context.Context) error {
	_, err := runWithRetry(ctx, OperationStartupValidation, objectMetadataTimeout, true, func(ctx context.Context) (struct{}, error) {
		_, statErr := s.client.StatObject(ctx, s.bucket, ".cartulary/startup/capability-check", minio.StatObjectOptions{})
		if statErr == nil {
			return struct{}{}, nil
		}
		mapped := mapBackendError(OperationStartupValidation, statErr)
		if IsObjectNotFound(mapped) {
			return struct{}{}, nil
		}
		return struct{}{}, mapped
	})
	if err != nil {
		return err
	}
	_, err = s.client.PresignedPutObject(ctx, s.bucket, ".cartulary/startup/direct-put-check", time.Minute)
	if err != nil {
		return mapBackendError(OperationStartupValidation, err)
	}
	return nil
}

type S3Store struct {
	client *minio.Client
	bucket string
}

func (s *S3Store) UploadTarget(ctx context.Context, key string, expiresAt time.Time) (UploadTarget, error) {
	return s.CreateUploadTarget(ctx, UploadTargetRequest{Key: key, ByteSize: -1, ExpiresAt: expiresAt, Purpose: PurposeProductUpload})
}

func (s *S3Store) CreateUploadTarget(ctx context.Context, request UploadTargetRequest) (UploadTarget, error) {
	if err := validateUploadTargetRequest(request); err != nil {
		return UploadTarget{}, err
	}
	expiresIn := time.Until(request.ExpiresAt)
	target, err := runWithRetry(ctx, OperationCreateUploadTarget, uploadTargetTimeout, true, func(ctx context.Context) (UploadTarget, error) {
		targetURL, err := s.client.PresignedPutObject(ctx, s.bucket, request.Key, expiresIn)
		if err != nil {
			return UploadTarget{}, err
		}
		return UploadTarget{Href: targetURL.String(), Method: "PUT", Headers: map[string]string{}}, nil
	})
	if err != nil {
		return UploadTarget{}, err
	}
	return target, nil
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
	opts := minio.PutObjectOptions{ContentType: request.ContentType}
	if len(request.Metadata) > 0 {
		opts.UserMetadata = map[string]string(request.Metadata)
	}
	info, err := runWithRetry(ctx, OperationPutObject, objectMutationTimeout, false, func(ctx context.Context) (minio.UploadInfo, error) {
		return s.client.PutObject(ctx, s.bucket, request.Key, request.Body, request.Size, opts)
	})
	if err != nil {
		return PutObjectResult{}, err
	}
	return PutObjectResult{ETag: info.ETag, SizeBytes: info.Size, ContentType: request.ContentType, Metadata: request.Metadata}, nil
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
		return nil, ObjectInfo{}, adapterError(OperationGetObjectRange, ErrorCodeRangeNotSatisfiable, ReasonRangeInvalid, false, "object range not satisfiable", nil)
	}
	opts := minio.GetObjectOptions{}
	if request.RangeStart != nil {
		end := int64(0)
		if request.RangeEnd != nil {
			end = *request.RangeEnd
		}
		if err := opts.SetRange(*request.RangeStart, end); err != nil {
			return nil, ObjectInfo{}, mapBackendError(operation, err)
		}
	}
	object, cancel, err := s.getObjectStreamWithRetry(ctx, operation, request.Key, opts)
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	return closeObservedStream(operation, request.Key, cancelOnCloseReadCloser{ReadCloser: object, cancel: cancel}), info, nil
}

func (s *S3Store) getObjectStreamWithRetry(ctx context.Context, operation Operation, key string, opts minio.GetObjectOptions) (*minio.Object, context.CancelFunc, error) {
	attemptLimit := maxTotalAttempts
	var lastErr error
	for attempt := 1; attempt <= attemptLimit; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, nil, mapBackendError(operation, err)
		}
		attemptCtx, cancel := context.WithTimeout(ctx, objectReadTimeout)
		object, err := s.client.GetObject(attemptCtx, s.bucket, key, opts)
		if err == nil {
			return object, cancel, nil
		}
		cancel()
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
				return nil, nil, mapBackendError(operation, ctx.Err())
			}
		}
	}
	if adapterErr, ok := AsAdapterError(lastErr); ok && adapterErr.Retryable {
		return nil, nil, adapterError(operation, ErrorCodeRetryExhausted, ReasonRetryExhausted, false, "retryable object-store operation exhausted attempts", lastErr)
	}
	return nil, nil, lastErr
}

func (s *S3Store) StatObject(ctx context.Context, key string) (ObjectInfo, error) {
	return s.Head(ctx, HeadObjectRequest{Key: key, Purpose: PurposeProductRead})
}

func (s *S3Store) Head(ctx context.Context, request HeadObjectRequest) (ObjectInfo, error) {
	if err := validateHeadObjectRequest(request); err != nil {
		return ObjectInfo{}, err
	}
	stat, err := runWithRetry(ctx, OperationHeadObject, objectMetadataTimeout, true, func(ctx context.Context) (minio.ObjectInfo, error) {
		return s.client.StatObject(ctx, s.bucket, request.Key, minio.StatObjectOptions{})
	})
	if err != nil {
		return ObjectInfo{}, err
	}
	return ObjectInfo{Key: request.Key, Size: stat.Size, ContentType: stat.ContentType}, nil
}

func (s *S3Store) ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	result, err := s.ListPrefix(ctx, ListPrefixRequest{Prefix: prefix, Purpose: PurposeTestCleanup})
	if err != nil {
		return nil, err
	}
	return result.Objects, nil
}

func (s *S3Store) ListPrefix(ctx context.Context, request ListPrefixRequest) (ListPrefixResult, error) {
	if err := validateListPrefixRequest(request); err != nil {
		return ListPrefixResult{}, err
	}
	objects, err := runWithRetry(ctx, OperationListPrefix, objectMutationTimeout, true, func(ctx context.Context) ([]ObjectInfo, error) {
		objects := make([]ObjectInfo, 0)
		for objectInfo := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{Prefix: request.Prefix, Recursive: true}) {
			if objectInfo.Err != nil {
				return objects, objectInfo.Err
			}
			objects = append(objects, ObjectInfo{Key: objectInfo.Key, Size: objectInfo.Size, ContentType: objectInfo.ContentType})
		}
		return objects, nil
	})
	if err != nil {
		return ListPrefixResult{}, err
	}
	sort.Slice(objects, func(left, right int) bool {
		return objects[left].Key < objects[right].Key
	})
	return ListPrefixResult{Objects: objects}, nil
}

func (s *S3Store) DeleteObject(ctx context.Context, key string) error {
	return s.Delete(ctx, DeleteObjectRequest{Key: key, Purpose: PurposeTestCleanup})
}

func (s *S3Store) Delete(ctx context.Context, request DeleteObjectRequest) error {
	if err := validateDeleteObjectRequest(request); err != nil {
		return err
	}
	_, err := runWithRetry(ctx, OperationDeleteObject, objectMutationTimeout, true, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, s.client.RemoveObject(ctx, s.bucket, request.Key, minio.RemoveObjectOptions{})
	})
	return err
}

func (s *S3Store) EnsureBucketForDevTest(ctx context.Context, request EnsureBucketRequest) (EnsureBucketResult, error) {
	if err := validateEnsureBucketRequest(request); err != nil {
		return EnsureBucketResult{}, err
	}
	exists, err := runWithRetry(ctx, OperationEnsureBucketForDevTest, objectMetadataTimeout, true, func(ctx context.Context) (bool, error) {
		return s.client.BucketExists(ctx, s.bucket)
	})
	if err != nil {
		return EnsureBucketResult{}, err
	}
	if exists {
		return EnsureBucketResult{AlreadyExists: true}, nil
	}
	_, err = runWithRetry(ctx, OperationEnsureBucketForDevTest, objectMutationTimeout, true, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
	})
	if err != nil {
		return EnsureBucketResult{}, err
	}
	return EnsureBucketResult{Created: true}, nil
}

func (s *S3Store) Close() error {
	return nil
}

type FilesystemStore struct {
	root         *os.Root
	uploadTokens map[string]filesystemUploadTarget
	mu           sync.Mutex
}

type filesystemUploadTarget struct {
	Key       string
	ExpiresAt time.Time
}

type filesystemMetadata struct {
	ContentType string `json:"content_type,omitempty"`
}

func NewFilesystemStore(root string) (*FilesystemStore, error) {
	if root == "" {
		return nil, fmt.Errorf("create filesystem object store: root path is required")
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create filesystem object-store root: %w", err)
	}
	rootHandle, err := os.OpenRoot(filepath.Clean(root))
	if err != nil {
		return nil, fmt.Errorf("open filesystem object-store root: %w", err)
	}
	return &FilesystemStore{
		root:         rootHandle,
		uploadTokens: make(map[string]filesystemUploadTarget),
	}, nil
}

func (s *FilesystemStore) UploadTarget(_ context.Context, key string, expiresAt time.Time) (UploadTarget, error) {
	if _, err := s.resolvePath(key); err != nil {
		return UploadTarget{}, err
	}
	if !expiresAt.After(time.Now()) {
		return UploadTarget{}, fmt.Errorf("create filesystem upload target: expires_at must be in the future")
	}
	token, err := randomToken()
	if err != nil {
		return UploadTarget{}, err
	}
	s.mu.Lock()
	s.uploadTokens[token] = filesystemUploadTarget{Key: key, ExpiresAt: expiresAt}
	s.mu.Unlock()
	return UploadTarget{
		Href:    "/api/v1/object-uploads/" + url.PathEscape(token),
		Method:  "PUT",
		Headers: map[string]string{},
	}, nil
}

func (s *FilesystemStore) CompleteUploadTarget(ctx context.Context, token string, body io.Reader, contentType string) error {
	s.mu.Lock()
	target, ok := s.uploadTokens[token]
	if ok && time.Now().After(target.ExpiresAt) {
		delete(s.uploadTokens, token)
		ok = false
	}
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("complete filesystem upload target: unknown or expired upload token")
	}
	return s.PutObject(ctx, target.Key, body, -1, contentType)
}

func (s *FilesystemStore) PutObject(ctx context.Context, key string, body io.Reader, _ int64, contentType string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	relativePath, err := s.resolvePath(key)
	if err != nil {
		return err
	}
	if err := s.root.MkdirAll(filepath.Dir(relativePath), 0o700); err != nil {
		return fmt.Errorf("create filesystem object parent: %w", err)
	}
	tmpPath := relativePath + ".tmp"
	file, err := s.root.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create filesystem object: %w", err)
	}
	_, copyErr := io.Copy(file, body)
	closeErr := file.Close()
	if copyErr != nil {
		_ = s.root.Remove(tmpPath)
		return fmt.Errorf("write filesystem object: %w", copyErr)
	}
	if closeErr != nil {
		_ = s.root.Remove(tmpPath)
		return fmt.Errorf("close filesystem object: %w", closeErr)
	}
	if err := s.root.Rename(tmpPath, relativePath); err != nil {
		_ = s.root.Remove(tmpPath)
		return fmt.Errorf("commit filesystem object: %w", err)
	}
	if err := s.writeMetadata(relativePath, contentType); err != nil {
		return err
	}
	return nil
}

func (s *FilesystemStore) ReadObject(_ context.Context, key string, options ReadOptions) (io.ReadCloser, ObjectInfo, error) {
	relativePath, err := s.resolvePath(key)
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	file, err := s.root.Open(relativePath)
	if err != nil {
		return nil, ObjectInfo{}, fmt.Errorf("open filesystem object: %w", err)
	}
	info, err := s.statPath(key, relativePath)
	if err != nil {
		_ = file.Close()
		return nil, ObjectInfo{}, err
	}
	if options.RangeStart != nil {
		start := *options.RangeStart
		if _, err := file.Seek(start, io.SeekStart); err != nil {
			_ = file.Close()
			return nil, ObjectInfo{}, err
		}
		if options.RangeEnd != nil {
			length := *options.RangeEnd - start + 1
			if length < 0 {
				_ = file.Close()
				return nil, ObjectInfo{}, fmt.Errorf("read filesystem object: invalid byte range")
			}
			return closeObservedStream(OperationGetObjectRange, key, readLimitCloser{Reader: io.LimitReader(file, length), Closer: file}), info, nil
		}
	}
	operation := OperationGetObject
	if options.RangeStart != nil {
		operation = OperationGetObjectRange
	}
	return closeObservedStream(operation, key, file), info, nil
}

func (s *FilesystemStore) StatObject(_ context.Context, key string) (ObjectInfo, error) {
	relativePath, err := s.resolvePath(key)
	if err != nil {
		return ObjectInfo{}, err
	}
	return s.statPath(key, relativePath)
}

func (s *FilesystemStore) ListObjects(_ context.Context, prefix string) ([]ObjectInfo, error) {
	objects := make([]ObjectInfo, 0)
	if err := fs.WalkDir(s.root.FS(), ".", func(relativePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".meta.json") || strings.HasSuffix(entry.Name(), ".tmp") {
			return nil
		}
		key := filepath.ToSlash(relativePath)
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			return nil
		}
		info, err := s.statPath(key, relativePath)
		if err != nil {
			return err
		}
		objects = append(objects, info)
		return nil
	}); err != nil {
		return objects, fmt.Errorf("list filesystem objects: %w", err)
	}
	return objects, nil
}

func (s *FilesystemStore) DeleteObject(_ context.Context, key string) error {
	relativePath, err := s.resolvePath(key)
	if err != nil {
		return err
	}
	if err := s.root.Remove(relativePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete filesystem object: %w", err)
	}
	if err := s.root.Remove(metadataPath(relativePath)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete filesystem object metadata: %w", err)
	}
	return nil
}

func (s *FilesystemStore) CreateUploadTarget(ctx context.Context, request UploadTargetRequest) (UploadTarget, error) {
	if err := validateUploadTargetRequest(request); err != nil {
		return UploadTarget{}, err
	}
	target, err := s.UploadTarget(ctx, request.Key, request.ExpiresAt)
	if err != nil {
		return UploadTarget{}, mapBackendError(OperationCreateUploadTarget, err)
	}
	return target, nil
}

func (s *FilesystemStore) Put(ctx context.Context, request PutObjectRequest) (PutObjectResult, error) {
	if err := validatePutObjectRequest(request); err != nil {
		return PutObjectResult{}, err
	}
	if err := s.PutObject(ctx, request.Key, request.Body, request.Size, request.ContentType); err != nil {
		return PutObjectResult{}, mapBackendError(OperationPutObject, err)
	}
	return PutObjectResult{SizeBytes: request.Size, ContentType: request.ContentType, Metadata: request.Metadata}, nil
}

func (s *FilesystemStore) Head(ctx context.Context, request HeadObjectRequest) (ObjectInfo, error) {
	if err := validateHeadObjectRequest(request); err != nil {
		return ObjectInfo{}, err
	}
	info, err := s.StatObject(ctx, request.Key)
	if err != nil {
		return ObjectInfo{}, mapBackendError(OperationHeadObject, err)
	}
	return info, nil
}

func (s *FilesystemStore) Get(ctx context.Context, request GetObjectRequest) (io.ReadCloser, ObjectInfo, error) {
	if err := validateGetObjectRequest(request); err != nil {
		return nil, ObjectInfo{}, err
	}
	operation := OperationGetObject
	if request.RangeStart != nil {
		operation = OperationGetObjectRange
		info, err := s.Head(ctx, HeadObjectRequest{Key: request.Key, Purpose: request.Purpose})
		if err != nil {
			return nil, ObjectInfo{}, err
		}
		if *request.RangeStart >= info.Size {
			return nil, ObjectInfo{}, adapterError(OperationGetObjectRange, ErrorCodeRangeNotSatisfiable, ReasonRangeInvalid, false, "object range not satisfiable", nil)
		}
	}
	object, info, err := s.ReadObject(ctx, request.Key, ReadOptions{RangeStart: request.RangeStart, RangeEnd: request.RangeEnd})
	if err != nil {
		return nil, ObjectInfo{}, mapBackendError(operation, err)
	}
	return object, info, nil
}

func (s *FilesystemStore) ListPrefix(ctx context.Context, request ListPrefixRequest) (ListPrefixResult, error) {
	if err := validateListPrefixRequest(request); err != nil {
		return ListPrefixResult{}, err
	}
	objects, err := s.ListObjects(ctx, request.Prefix)
	if err != nil {
		return ListPrefixResult{}, mapBackendError(OperationListPrefix, err)
	}
	sort.Slice(objects, func(left, right int) bool {
		return objects[left].Key < objects[right].Key
	})
	return ListPrefixResult{Objects: objects}, nil
}

func (s *FilesystemStore) Delete(ctx context.Context, request DeleteObjectRequest) error {
	if err := validateDeleteObjectRequest(request); err != nil {
		return err
	}
	if err := s.DeleteObject(ctx, request.Key); err != nil {
		return mapBackendError(OperationDeleteObject, err)
	}
	return nil
}

func (s *FilesystemStore) EnsureBucketForDevTest(_ context.Context, request EnsureBucketRequest) (EnsureBucketResult, error) {
	if err := validateEnsureBucketRequest(request); err != nil {
		return EnsureBucketResult{}, err
	}
	return EnsureBucketResult{AlreadyExists: true}, nil
}

func (s *FilesystemStore) resolvePath(key string) (string, error) {
	if key == "" || strings.ContainsRune(key, '\x00') {
		return "", fmt.Errorf("resolve filesystem object key: key is required")
	}
	if filepath.IsAbs(key) {
		return "", fmt.Errorf("resolve filesystem object key %q: absolute keys are not allowed", key)
	}
	cleaned := filepath.Clean(filepath.FromSlash(key))
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("resolve filesystem object key %q: key escapes object-store root", key)
	}
	return cleaned, nil
}

func (s *FilesystemStore) statPath(key string, relativePath string) (ObjectInfo, error) {
	stat, err := s.root.Stat(relativePath)
	if err != nil {
		return ObjectInfo{}, fmt.Errorf("stat filesystem object: %w", err)
	}
	if stat.IsDir() {
		return ObjectInfo{}, fmt.Errorf("stat filesystem object: object key %q resolved to a directory", key)
	}
	return ObjectInfo{Key: key, Size: stat.Size(), ContentType: s.readContentType(relativePath)}, nil
}

func (s *FilesystemStore) writeMetadata(relativePath string, contentType string) error {
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		return nil
	}
	if _, _, err := mime.ParseMediaType(contentType); err != nil {
		contentType = "application/octet-stream"
	}
	payload, err := json.Marshal(filesystemMetadata{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("marshal filesystem object metadata: %w", err)
	}
	if err := s.root.WriteFile(metadataPath(relativePath), payload, 0o600); err != nil {
		return fmt.Errorf("write filesystem object metadata: %w", err)
	}
	return nil
}

func (s *FilesystemStore) readContentType(relativePath string) string {
	payload, err := s.root.ReadFile(metadataPath(relativePath))
	if err != nil {
		return ""
	}
	var metadata filesystemMetadata
	if err := json.Unmarshal(payload, &metadata); err != nil {
		return ""
	}
	return metadata.ContentType
}

func metadataPath(path string) string {
	return path + ".meta.json"
}

func (s *FilesystemStore) Close() error {
	return s.root.Close()
}

type readLimitCloser struct {
	io.Reader
	io.Closer
}

type cancelOnCloseReadCloser struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (reader cancelOnCloseReadCloser) Close() error {
	err := reader.ReadCloser.Close()
	if reader.cancel != nil {
		reader.cancel()
	}
	return err
}

func resolveManagedServiceSettings(serviceRef string, env map[string]string) (Settings, error) {
	keys, err := EnvKeysForServiceRef(serviceRef)
	if err != nil {
		return Settings{}, err
	}

	settings := Settings{
		Endpoint:  lookupEnvValue(env, keys.Endpoint),
		AccessKey: lookupEnvValue(env, keys.AccessKey),
		SecretKey: lookupEnvValue(env, keys.SecretKey),
		Bucket:    lookupEnvValue(env, keys.Bucket),
	}
	if settings.Endpoint == "" {
		return Settings{}, fmt.Errorf("missing object-store endpoint for managed service %q (%s)", serviceRef, keys.Endpoint)
	}
	if settings.AccessKey == "" {
		return Settings{}, fmt.Errorf("missing object-store access key for managed service %q (%s)", serviceRef, keys.AccessKey)
	}
	if settings.SecretKey == "" {
		return Settings{}, fmt.Errorf("missing object-store secret key for managed service %q (%s)", serviceRef, keys.SecretKey)
	}
	if settings.Bucket == "" {
		return Settings{}, fmt.Errorf("missing object-store bucket for managed service %q (%s)", serviceRef, keys.Bucket)
	}

	if secureValue := lookupEnvValue(env, keys.Secure); secureValue != "" {
		secure, err := strconv.ParseBool(secureValue)
		if err != nil {
			return Settings{}, fmt.Errorf("parse %s: %w", keys.Secure, err)
		}
		settings.Secure = secure
	}

	return settings, nil
}

func EnvKeysForServiceRef(serviceRef string) (ServiceRefEnvKeys, error) {
	normalized := NormalizeServiceRef(serviceRef)
	if normalized == "" {
		return ServiceRefEnvKeys{}, fmt.Errorf("resolve managed object-store env keys: service_ref must contain at least one letter or digit")
	}

	return ServiceRefEnvKeys{
		Endpoint:  "CARTULARY_S3_" + normalized + "_ENDPOINT",
		AccessKey: "CARTULARY_S3_" + normalized + "_ACCESS_KEY_ID",
		SecretKey: "CARTULARY_S3_" + normalized + "_SECRET_ACCESS_KEY",
		Secure:    "CARTULARY_S3_" + normalized + "_SECURE",
		Bucket:    "CARTULARY_S3_" + normalized + "_BUCKET",
	}, nil
}

func NormalizeServiceRef(value string) string {
	var builder strings.Builder
	previousUnderscore := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(unicode.ToUpper(r))
			previousUnderscore = false
		case !previousUnderscore:
			builder.WriteByte('_')
			previousUnderscore = true
		}
	}

	return strings.Trim(builder.String(), "_")
}

func lookupEnv(env map[string]string, key string) (string, bool) {
	if env != nil {
		value, ok := env[key]
		return value, ok
	}

	return os.LookupEnv(key)
}

func lookupEnvValue(env map[string]string, key string) string {
	value, _ := lookupEnv(env, key)
	return value
}

func randomToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate upload token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}
