package objectstore

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/url"
	"os"
	pathpkg "path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

const (
	EndpointEnv        = "CARTULARY_S3_ENDPOINT"
	AccessKeyEnv       = "CARTULARY_S3_ACCESS_KEY_ID"
	SecretKeyEnv       = "CARTULARY_S3_SECRET_ACCESS_KEY"
	SecureEnv          = "CARTULARY_S3_SECURE"
	BucketEnv          = "CARTULARY_S3_BUCKET"
	RegionEnv          = "CARTULARY_S3_REGION"
	RootCertificateEnv = "CARTULARY_S3_ROOT_CERTIFICATE_PATH"
)

type Settings struct {
	BindingKind         string
	RootPath            string
	Endpoint            string
	AccessKey           string
	SecretKey           string
	Secure              bool
	Bucket              string
	Region              string
	RootCertificatePath string
}

type Binding struct {
	BindingKind string
	RootPath    string
	ServiceRef  string
}

type Instrumentation struct {
	Enabled        bool
	ServiceVersion string
}

type ServiceRefEnvKeys struct {
	Endpoint            string
	AccessKey           string
	SecretKey           string
	Secure              string
	Bucket              string
	Region              string
	RootCertificatePath string
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

func Setup(ctx context.Context, settings Settings, instrumentation Instrumentation) (Store, error) {
	switch settings.BindingKind {
	case "filesystem_root":
		store, err := NewFilesystemStore(settings.RootPath)
		if err != nil {
			return nil, err
		}
		if instrumentation.Enabled {
			return InstrumentStore(store, instrumentation.ServiceVersion), nil
		}
		return store, nil
	case "managed_service":
		store, err := newS3Store(ctx, settings)
		if err != nil {
			return nil, err
		}
		if instrumentation.Enabled {
			return InstrumentStore(store, instrumentation.ServiceVersion), nil
		}
		return store, nil
	default:
		return nil, fmt.Errorf("setup object store: unsupported binding_kind %q", settings.BindingKind)
	}
}

func ResolveSettings(binding Binding, env map[string]string) (Settings, error) {
	switch binding.BindingKind {
	case "filesystem_root":
		if binding.RootPath == "" {
			return Settings{}, fmt.Errorf("resolve object-store settings: roots.object_storage.path is required")
		}
		return Settings{BindingKind: "filesystem_root", RootPath: binding.RootPath}, nil
	case "managed_service":
		settings, err := resolveManagedServiceSettings(binding.ServiceRef, env)
		if err != nil {
			return Settings{}, err
		}
		settings.BindingKind = "managed_service"
		return settings, nil
	default:
		return Settings{}, fmt.Errorf("resolve object-store settings: roots.object_storage.binding_kind must be configured before object-store setup")
	}
}

type FilesystemStore struct {
	root         *rootedfs.Root
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
	rootHandle, err := rootedfs.OpenOrCreate(root)
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
	reference, err := s.resolvePath(key)
	if err != nil {
		return err
	}
	if err := s.makeParent(reference); err != nil {
		return err
	}
	if err := s.root.AtomicReplace(ctx, reference, func(writer io.Writer) error {
		_, copyErr := io.Copy(writer, body)
		return copyErr
	}); err != nil {
		return fmt.Errorf("commit filesystem object: %w", err)
	}
	if err := s.writeMetadata(ctx, reference, contentType); err != nil {
		return err
	}
	return nil
}

func (s *FilesystemStore) ReadObject(_ context.Context, key string, options ReadOptions) (io.ReadCloser, ObjectInfo, error) {
	reference, err := s.resolvePath(key)
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	file, metadata, err := s.root.OpenRegular(reference)
	if err != nil {
		return nil, ObjectInfo{}, fmt.Errorf("open filesystem object: %w", err)
	}
	info := ObjectInfo{Key: key, Size: metadata.Size, ContentType: s.readContentType(reference)}
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
	reference, err := s.resolvePath(key)
	if err != nil {
		return ObjectInfo{}, err
	}
	return s.statPath(key, reference)
}

func (s *FilesystemStore) ListObjects(_ context.Context, prefix string) ([]ObjectInfo, error) {
	objects := make([]ObjectInfo, 0)
	entries, err := s.root.ListRegular()
	if err != nil {
		return objects, fmt.Errorf("list filesystem objects: %w", err)
	}
	for _, entry := range entries {
		key := entry.Reference.String()
		if strings.HasSuffix(key, ".meta.json") || strings.HasSuffix(key, ".tmp") {
			continue
		}
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			continue
		}
		objects = append(objects, ObjectInfo{
			Key:         key,
			Size:        entry.Metadata.Size,
			ContentType: s.readContentType(entry.Reference),
		})
	}
	return objects, nil
}

func (s *FilesystemStore) DeleteObject(_ context.Context, key string) error {
	reference, err := s.resolvePath(key)
	if err != nil {
		return err
	}
	if err := s.root.RemoveRegular(reference); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("delete filesystem object: %w", err)
	}
	metadataReference := metadataReference(reference)
	if err := s.root.RemoveRegular(metadataReference); err != nil && !errors.Is(err, fs.ErrNotExist) {
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

func (s *FilesystemStore) resolvePath(key string) (rootedfs.Reference, error) {
	reference, err := rootedfs.ParseReference(key)
	if err != nil {
		return rootedfs.Reference{}, fmt.Errorf("resolve filesystem object key: invalid logical reference: %w", err)
	}
	return reference, nil
}

func (s *FilesystemStore) statPath(key string, reference rootedfs.Reference) (ObjectInfo, error) {
	file, metadata, err := s.root.OpenRegular(reference)
	if err != nil {
		return ObjectInfo{}, fmt.Errorf("stat filesystem object: %w", err)
	}
	if err := file.Close(); err != nil {
		return ObjectInfo{}, fmt.Errorf("stat filesystem object: %w", err)
	}
	return ObjectInfo{Key: key, Size: metadata.Size, ContentType: s.readContentType(reference)}, nil
}

func (s *FilesystemStore) writeMetadata(ctx context.Context, reference rootedfs.Reference, contentType string) error {
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
	metadata := metadataReference(reference)
	if err := s.makeParent(metadata); err != nil {
		return err
	}
	if err := s.root.AtomicReplace(ctx, metadata, func(writer io.Writer) error {
		_, writeErr := writer.Write(payload)
		return writeErr
	}); err != nil {
		return fmt.Errorf("write filesystem object metadata: %w", err)
	}
	return nil
}

func (s *FilesystemStore) readContentType(reference rootedfs.Reference) string {
	payload, _, err := s.root.ReadRegular(metadataReference(reference), 4096)
	if err != nil {
		return ""
	}
	var metadata filesystemMetadata
	if err := json.Unmarshal(payload, &metadata); err != nil {
		return ""
	}
	return metadata.ContentType
}

func (s *FilesystemStore) makeParent(reference rootedfs.Reference) error {
	parent := pathpkg.Dir(reference.String())
	if parent == "." {
		return nil
	}
	parentReference, err := rootedfs.ParseReference(parent)
	if err != nil {
		return fmt.Errorf("resolve filesystem object parent: %w", err)
	}
	if err := s.root.MakePrivateDir(parentReference); err != nil {
		return fmt.Errorf("create filesystem object parent: %w", err)
	}
	return nil
}

func metadataReference(reference rootedfs.Reference) rootedfs.Reference {
	return rootedfs.MustParseReference(reference.String() + ".meta.json")
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
		Secure:              true,
		Endpoint:            lookupEnvValue(env, keys.Endpoint),
		AccessKey:           lookupEnvValue(env, keys.AccessKey),
		SecretKey:           lookupEnvValue(env, keys.SecretKey),
		Bucket:              lookupEnvValue(env, keys.Bucket),
		Region:              lookupEnvValue(env, keys.Region),
		RootCertificatePath: lookupEnvValue(env, keys.RootCertificatePath),
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

	if value, present := lookupEnv(env, keys.Secure); present && value != "true" {
		return Settings{}, fmt.Errorf("managed object store requires verified TLS")
	}

	return settings, nil
}

func EnvKeysForServiceRef(serviceRef string) (ServiceRefEnvKeys, error) {
	normalized := NormalizeServiceRef(serviceRef)
	if normalized == "" {
		return ServiceRefEnvKeys{}, fmt.Errorf("resolve managed object-store env keys: service_ref must contain at least one letter or digit")
	}

	return ServiceRefEnvKeys{
		Endpoint:            "CARTULARY_S3_" + normalized + "_ENDPOINT",
		AccessKey:           "CARTULARY_S3_" + normalized + "_ACCESS_KEY_ID",
		SecretKey:           "CARTULARY_S3_" + normalized + "_SECRET_ACCESS_KEY",
		Secure:              "CARTULARY_S3_" + normalized + "_SECURE",
		Bucket:              "CARTULARY_S3_" + normalized + "_BUCKET",
		Region:              "CARTULARY_S3_" + normalized + "_REGION",
		RootCertificatePath: "CARTULARY_S3_" + normalized + "_ROOT_CERTIFICATE_PATH",
	}, nil
}

func NormalizeServiceRef(value string) string {
	var builder strings.Builder
	previousUnderscore := false
	for i := 0; i < len(value); i++ {
		c := value[i]
		switch {
		case c >= 'a' && c <= 'z':
			builder.WriteByte(c - ('a' - 'A'))
			previousUnderscore = false
		case c >= 'A' && c <= 'Z' || c >= '0' && c <= '9':
			builder.WriteByte(c)
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
