package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	probeSchemaID         = "cartulary.object_store_capability_probe.v1"
	compatibilitySchemaID = "cartulary.seaweedfs_s3_compatibility_report.v1"
	uploadMode            = "direct_presigned_put"
	objectStoreBackend    = "seaweedfs_s3"
)

type config struct {
	Mode            string
	ProfileID       string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Secure          bool
	Bucket          string
	Origin          string
	ServiceName     string
	Image           string
	ImageDigest     string
	Timeout         time.Duration
	ArtifactPath    string
}

type redactionObject struct {
	Redacted       bool    `json:"redacted"`
	RedactionClass string  `json:"redaction_class"`
	SHA256         string  `json:"sha256,omitempty"`
	Scheme         *string `json:"scheme,omitempty"`
	HostSHA256     string  `json:"host_sha256,omitempty"`
	PortPresent    *bool   `json:"port_present,omitempty"`
}

type probeArtifact struct {
	SchemaID           string            `json:"schema_id"`
	ProbeID            string            `json:"probe_id"`
	ProfileID          string            `json:"profile_id"`
	UploadMode         string            `json:"upload_mode"`
	ObjectStoreBackend string            `json:"object_store_backend"`
	ServiceName        string            `json:"service_name,omitempty"`
	Image              string            `json:"image,omitempty"`
	ImageDigest        string            `json:"image_digest,omitempty"`
	StartedAt          string            `json:"started_at"`
	CompletedAt        string            `json:"completed_at"`
	Result             string            `json:"result"`
	FailedStage        *string           `json:"failed_stage"`
	ReasonCode         *string           `json:"reason_code"`
	Retryable          bool              `json:"retryable"`
	Endpoint           redactionObject   `json:"endpoint"`
	BucketRef          redactionObject   `json:"bucket_ref"`
	Steps              []probeStep       `json:"steps"`
	CleanupResult      string            `json:"cleanup_result"`
	RetainedProbeKeys  []redactionObject `json:"retained_probe_keys"`
}

type probeStep struct {
	Stage        string  `json:"stage"`
	Status       string  `json:"status"`
	AttemptCount int     `json:"attempt_count"`
	StartedAt    *string `json:"started_at"`
	CompletedAt  *string `json:"completed_at"`
	ReasonCode   *string `json:"reason_code"`
}

type compatibilityReport struct {
	SchemaID           string              `json:"schema_id"`
	ProbeID            string              `json:"probe_id"`
	ObjectStoreBackend string              `json:"object_store_backend"`
	StartedAt          string              `json:"started_at"`
	CompletedAt        string              `json:"completed_at"`
	Result             string              `json:"result"`
	Cases              []compatibilityCase `json:"cases"`
	ForbiddenSkipRows  []string            `json:"forbidden_skip_rows"`
}

type compatibilityCase struct {
	CaseID     string         `json:"case_id"`
	Capability string         `json:"capability"`
	Status     string         `json:"status"`
	ReasonCode *string        `json:"reason_code"`
	Evidence   map[string]any `json:"evidence"`
}

type probeRunner struct {
	cfg             config
	client          *minio.Client
	artifact        probeArtifact
	compatibility   compatibilityReport
	probeID         string
	probePrefix     string
	probeKey        string
	directUploadKey string
	createdKeys     map[string]bool
	deletedKeys     map[string]bool
	compatStatus    map[string]string
}

func main() {
	cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "object-store probe configuration error: %v\n", err)
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	runner, err := newProbeRunner(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "object-store probe setup failed: %v\n", sanitizeMessage(cfg, err.Error()))
		os.Exit(1)
	}

	runErr := runner.run(ctx)
	runner.finish(runErr)

	if err := writeProbeArtifact(cfg, runner.artifact); err != nil {
		fmt.Fprintf(os.Stderr, "object-store probe artifact error: %v\n", err)
		os.Exit(11)
	}
	if cfg.Mode == "probe" {
		if err := writeCompatibilityArtifact(cfg, runner.compatibility); err != nil {
			fmt.Fprintf(os.Stderr, "object-store compatibility artifact error: %v\n", err)
			os.Exit(11)
		}
	}

	if runErr != nil {
		fmt.Fprintf(os.Stderr, "object-store probe failed: %v\n", sanitizeMessage(cfg, runErr.Error()))
		os.Exit(1)
	}
}

func parseConfig() (config, error) {
	var secureText string
	cfg := config{}
	flag.StringVar(&cfg.Mode, "mode", envDefault("OBJECT_STORE_PROBE_MODE", "probe"), "probe mode: probe or reset")
	flag.StringVar(&cfg.ProfileID, "profile-id", envDefault("OBJECT_STORE_PROFILE_ID", "local_dev"), "runtime profile id")
	flag.StringVar(&cfg.Endpoint, "endpoint", envDefault("OBJECT_STORE_ENDPOINT", "127.0.0.1:8333"), "S3 endpoint without scheme")
	flag.StringVar(&cfg.AccessKeyID, "access-key-id", envDefault("SEAWEEDFS_S3_ACCESS_KEY_ID", "cartulary-local"), "S3 access key")
	flag.StringVar(&cfg.SecretAccessKey, "secret-access-key", envDefault("SEAWEEDFS_S3_SECRET_ACCESS_KEY", "cartulary-local-secret"), "S3 secret key")
	flag.StringVar(&secureText, "secure", envDefault("OBJECT_STORE_SECURE", "false"), "whether to use HTTPS")
	flag.StringVar(&cfg.Bucket, "bucket", envDefault("OBJECT_STORE_BUCKET", "cartulary"), "S3 bucket")
	flag.StringVar(&cfg.Origin, "origin", envDefault("OBJECT_STORE_CORS_ORIGIN", "http://localhost:5173"), "browser origin for CORS preflight")
	flag.StringVar(&cfg.ServiceName, "service-name", envDefault("OBJECT_STORE_SERVICE_NAME", "seaweedfs-s3"), "local object-store service name")
	flag.StringVar(&cfg.Image, "image", envDefault("SEAWEEDFS_S3_IMAGE", "docker.io/chrislusf/seaweedfs:4.17"), "object-store image tag")
	flag.StringVar(&cfg.ImageDigest, "image-digest", envDefault("SEAWEEDFS_S3_IMAGE_DIGEST", "sha256:186de7ef977a20343ee9a5544073f081976a29e2d29ecf8379891e7bf177fbe9"), "object-store image digest")
	flag.DurationVar(&cfg.Timeout, "timeout", 60*time.Second, "probe timeout")
	flag.StringVar(&cfg.ArtifactPath, "artifact", "", "artifact path; defaults to harness result path when available")
	flag.Parse()

	cfg.Mode = strings.TrimSpace(cfg.Mode)
	if cfg.Mode != "probe" && cfg.Mode != "reset" {
		return config{}, fmt.Errorf("mode must be probe or reset")
	}
	cfg.ProfileID = strings.TrimSpace(cfg.ProfileID)
	if cfg.ProfileID == "" {
		return config{}, fmt.Errorf("profile id is required")
	}
	cfg.Endpoint = normalizeEndpoint(cfg.Endpoint)
	if cfg.Endpoint == "" {
		return config{}, fmt.Errorf("endpoint is required")
	}
	if cfg.AccessKeyID == "" {
		return config{}, fmt.Errorf("access key is required")
	}
	if cfg.SecretAccessKey == "" {
		return config{}, fmt.Errorf("secret key is required")
	}
	if cfg.Bucket == "" {
		return config{}, fmt.Errorf("bucket is required")
	}
	secure, err := parseBool(secureText)
	if err != nil {
		return config{}, fmt.Errorf("secure: %w", err)
	}
	cfg.Secure = secure
	return cfg, nil
}

func newProbeRunner(cfg config) (*probeRunner, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.Secure,
	})
	if err != nil {
		return nil, fmt.Errorf("create S3 client: %w", err)
	}

	startedAt := time.Now().UTC()
	probeID, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	probePrefix := ".cartulary/probes/startup/" + probeID + "/"
	runner := &probeRunner{
		cfg:             cfg,
		client:          client,
		probeID:         probeID,
		probePrefix:     probePrefix,
		probeKey:        probePrefix + "probe.bin",
		directUploadKey: probePrefix + "direct-put.bin",
		createdKeys:     map[string]bool{},
		deletedKeys:     map[string]bool{},
		compatStatus:    map[string]string{},
	}
	runner.artifact = probeArtifact{
		SchemaID:           probeSchemaID,
		ProbeID:            probeID,
		ProfileID:          cfg.ProfileID,
		UploadMode:         uploadMode,
		ObjectStoreBackend: objectStoreBackend,
		ServiceName:        cfg.ServiceName,
		Image:              cfg.Image,
		ImageDigest:        cfg.ImageDigest,
		StartedAt:          startedAt.Format(time.RFC3339Nano),
		Result:             "fail",
		Endpoint:           endpointRef(cfg),
		BucketRef:          shaRef("bucket", cfg.Bucket),
		Steps:              []probeStep{},
		CleanupResult:      "cleanup_not_attempted",
		RetainedProbeKeys:  []redactionObject{},
	}
	runner.compatibility = compatibilityReport{
		SchemaID:           compatibilitySchemaID,
		ProbeID:            probeID,
		ObjectStoreBackend: objectStoreBackend,
		StartedAt:          runner.artifact.StartedAt,
		Result:             "fail",
		Cases:              []compatibilityCase{},
		ForbiddenSkipRows:  []string{},
	}
	return runner, nil
}

func (r *probeRunner) run(ctx context.Context) error {
	switch r.cfg.Mode {
	case "probe":
		return r.runProbe(ctx)
	case "reset":
		return r.runReset(ctx)
	default:
		return fmt.Errorf("unsupported mode %q", r.cfg.Mode)
	}
}

func (r *probeRunner) runProbe(ctx context.Context) (runErr error) {
	defer func() {
		if runErr != nil {
			_ = r.cleanupAfterFailure(ctx)
		}
	}()

	primaryPayload, err := sizedProbePayload("cartulary-object-store-probe-v1\n", 4097)
	if err != nil {
		return err
	}
	directPayload := []byte("cartulary-object-store-direct-put-probe-v1\n")

	var bucketExists bool
	if err := r.stage(ctx, "endpoint_reachability", func(ctx context.Context) error {
		_, err := r.client.BucketExists(ctx, r.cfg.Bucket)
		return err
	}); err != nil {
		return err
	}
	if err := r.stage(ctx, "bucket_validation", func(ctx context.Context) error {
		exists, err := r.client.BucketExists(ctx, r.cfg.Bucket)
		bucketExists = exists
		return err
	}); err != nil {
		return err
	}
	if !bucketExists {
		if !profileMayCreateBucket(r.cfg.ProfileID) {
			return r.stage(ctx, "bucket_creation", func(context.Context) error {
				return fmt.Errorf("bucket creation forbidden for profile %s", r.cfg.ProfileID)
			})
		}
		if err := r.stage(ctx, "bucket_creation", func(ctx context.Context) error {
			return r.client.MakeBucket(ctx, r.cfg.Bucket, minio.MakeBucketOptions{})
		}); err != nil {
			return err
		}
	}
	r.markCompat("SWFS-COMP-001", "pass")

	payloads := []struct {
		name    string
		key     string
		payload []byte
	}{
		{name: "primary", key: r.probeKey, payload: primaryPayload},
		{name: "zero", key: r.probePrefix + "payload-zero.bin", payload: []byte{}},
		{name: "small", key: r.probePrefix + "payload-37.bin", payload: []byte("0123456789abcdefghijklmnopqrstuvwxyz\n")},
		{name: "large", key: r.probePrefix + "payload-1m-plus-13.bin", payload: repeatedPayload(1024*1024 + 13)},
	}
	for _, payload := range payloads {
		if err := r.stage(ctx, "put_"+payload.name, func(ctx context.Context) error {
			_, err := r.client.PutObject(ctx, r.cfg.Bucket, payload.key, bytes.NewReader(payload.payload), int64(len(payload.payload)), minio.PutObjectOptions{ContentType: "application/octet-stream"})
			if err == nil {
				r.createdKeys[payload.key] = true
			}
			return err
		}); err != nil {
			return err
		}
	}
	r.markCompat("SWFS-COMP-002", "pass")
	for _, payload := range payloads {
		if err := r.stage(ctx, "head_"+payload.name, func(ctx context.Context) error {
			info, err := r.client.StatObject(ctx, r.cfg.Bucket, payload.key, minio.StatObjectOptions{})
			if err != nil {
				return err
			}
			if info.Size != int64(len(payload.payload)) {
				return fmt.Errorf("object size mismatch")
			}
			return nil
		}); err != nil {
			return err
		}
	}
	r.markCompat("SWFS-COMP-003", "pass")
	for _, payload := range payloads {
		if err := r.stage(ctx, "get_"+payload.name, func(ctx context.Context) error {
			return requireObjectPayload(ctx, r.client, r.cfg.Bucket, payload.key, payload.payload, minio.GetObjectOptions{})
		}); err != nil {
			return err
		}
	}
	r.markCompat("SWFS-COMP-004", "pass")
	if err := r.stage(ctx, "range_primary", func(ctx context.Context) error {
		opts := minio.GetObjectOptions{}
		if err := opts.SetRange(4, 15); err != nil {
			return err
		}
		return requireObjectPayload(ctx, r.client, r.cfg.Bucket, r.probeKey, primaryPayload[4:16], opts)
	}); err != nil {
		return err
	}
	r.markCompat("SWFS-COMP-005", "pass")
	if err := r.verifyPrefixIsolation(ctx); err != nil {
		return err
	}
	r.markCompat("SWFS-COMP-007", "pass")

	var putURL *url.URL
	if err := r.stage(ctx, "create_direct_upload_target", func(ctx context.Context) error {
		var err error
		putURL, err = r.client.PresignedPutObject(ctx, r.cfg.Bucket, r.directUploadKey, 5*time.Minute)
		return err
	}); err != nil {
		return err
	}
	if err := r.stage(ctx, "cors_preflight", func(ctx context.Context) error {
		return checkCORSPreflight(ctx, putURL.String(), r.cfg.Origin)
	}); err != nil {
		return err
	}
	r.markCompat("SWFS-COMP-011", "pass")
	if err := r.stage(ctx, "direct_put", func(ctx context.Context) error {
		err := uploadPresignedPUT(ctx, putURL.String(), r.cfg.Origin, directPayload)
		if err == nil {
			r.createdKeys[r.directUploadKey] = true
		}
		return err
	}); err != nil {
		return err
	}
	if err := r.stage(ctx, "head_direct", func(ctx context.Context) error {
		info, err := r.client.StatObject(ctx, r.cfg.Bucket, r.directUploadKey, minio.StatObjectOptions{})
		if err != nil {
			return err
		}
		if info.Size != int64(len(directPayload)) {
			return fmt.Errorf("direct upload size mismatch")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := r.stage(ctx, "get_direct", func(ctx context.Context) error {
		return requireObjectPayload(ctx, r.client, r.cfg.Bucket, r.directUploadKey, directPayload, minio.GetObjectOptions{})
	}); err != nil {
		return err
	}
	if err := r.verifyPresignedPutExpiry(ctx); err != nil {
		return err
	}
	r.markCompat("SWFS-COMP-008", "pass")
	if err := r.verifyErrorClassification(ctx); err != nil {
		return err
	}
	r.markCompat("SWFS-COMP-009", "pass")
	if err := r.verifyNoPublicListingPrimitive(ctx); err != nil {
		return err
	}
	r.markCompat("SWFS-COMP-010", "pass")
	if err := r.verifySameOriginHandleContract(ctx); err != nil {
		return err
	}
	r.markCompat("SWFS-COMP-012", "pass")
	if err := r.verifyCleanupClassification(ctx); err != nil {
		return err
	}
	r.markCompat("SWFS-COMP-013", "pass")
	if err := r.verifyCanonicalArtifactSerialization(ctx); err != nil {
		return err
	}
	r.markCompat("SWFS-COMP-014", "pass")

	for _, payload := range payloads {
		if err := r.deleteAndVerify(ctx, "delete_"+payload.name, "verify_"+payload.name+"_deleted", payload.key); err != nil {
			return err
		}
	}
	if err := r.deleteAndVerify(ctx, "delete_direct", "verify_direct_deleted", r.directUploadKey); err != nil {
		return err
	}
	r.markCompat("SWFS-COMP-006", "pass")
	return nil
}

func (r *probeRunner) deleteAndVerify(ctx context.Context, deleteStage string, verifyStage string, key string) error {
	if !r.createdKeys[key] {
		return nil
	}
	if err := r.stage(ctx, deleteStage, func(ctx context.Context) error {
		if err := r.client.RemoveObject(ctx, r.cfg.Bucket, key, minio.RemoveObjectOptions{}); err != nil {
			return err
		}
		r.deletedKeys[key] = true
		return nil
	}); err != nil {
		return err
	}
	return r.stage(ctx, verifyStage, func(ctx context.Context) error {
		_, err := r.client.StatObject(ctx, r.cfg.Bucket, key, minio.StatObjectOptions{})
		if err == nil {
			return fmt.Errorf("object still exists after delete")
		}
		if response := minio.ToErrorResponse(err); response.Code == "NoSuchKey" || response.Code == "NoSuchObject" || response.Code == "NotFound" {
			return nil
		}
		return err
	})
}

func (r *probeRunner) verifyPrefixIsolation(ctx context.Context) error {
	leftKey := r.probePrefix + "prefix-a/object.txt"
	rightKey := r.probePrefix + "prefix-b/object.txt"
	if err := r.stage(ctx, "put_prefix_isolation", func(ctx context.Context) error {
		for _, key := range []string{leftKey, rightKey} {
			if _, err := r.client.PutObject(ctx, r.cfg.Bucket, key, strings.NewReader("prefix"), int64(len("prefix")), minio.PutObjectOptions{}); err != nil {
				return err
			}
			r.createdKeys[key] = true
		}
		return nil
	}); err != nil {
		return err
	}
	if err := r.stage(ctx, "list_prefix_isolation", func(ctx context.Context) error {
		found := []string{}
		for item := range r.client.ListObjects(ctx, r.cfg.Bucket, minio.ListObjectsOptions{Prefix: r.probePrefix + "prefix-a/", Recursive: true}) {
			if item.Err != nil {
				return item.Err
			}
			found = append(found, item.Key)
		}
		if len(found) != 1 || found[0] != leftKey {
			return fmt.Errorf("prefix isolation mismatch")
		}
		return nil
	}); err != nil {
		return err
	}
	for _, key := range []string{leftKey, rightKey} {
		if err := r.deleteAndVerify(ctx, "delete_prefix_isolation", "verify_prefix_isolation_deleted", key); err != nil {
			return err
		}
	}
	return nil
}

func (r *probeRunner) verifyPresignedPutExpiry(ctx context.Context) error {
	expiredKey := r.probePrefix + "direct-put-expired.bin"
	var expiredURL *url.URL
	if err := r.stage(ctx, "create_expiring_direct_upload_target", func(ctx context.Context) error {
		var err error
		expiredURL, err = r.client.PresignedPutObject(ctx, r.cfg.Bucket, expiredKey, time.Second)
		return err
	}); err != nil {
		return err
	}
	return r.stage(ctx, "direct_put_after_expiry_rejected", func(ctx context.Context) error {
		timer := time.NewTimer(2500 * time.Millisecond)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
		return requirePresignedPUTRejected(ctx, expiredURL.String(), []byte("expired upload must fail"))
	})
}

func (r *probeRunner) verifyErrorClassification(ctx context.Context) error {
	return r.stage(ctx, "error_classification_matrix", func(ctx context.Context) error {
		checks := []struct {
			name  string
			want  string
			errFn func(context.Context) error
		}{
			{name: "missing_object", want: "object_missing", errFn: func(ctx context.Context) error {
				_, err := r.client.StatObject(ctx, r.cfg.Bucket, r.probePrefix+"missing.bin", minio.StatObjectOptions{})
				return err
			}},
			{name: "denied_credential", want: "credential_denied", errFn: func(ctx context.Context) error {
				client, err := minio.New(r.cfg.Endpoint, &minio.Options{
					Creds:  credentials.NewStaticV4(r.cfg.AccessKeyID+"-denied", r.cfg.SecretAccessKey+"-denied", ""),
					Secure: r.cfg.Secure,
				})
				if err != nil {
					return err
				}
				_, err = client.StatObject(ctx, r.cfg.Bucket, r.probePrefix+"missing.bin", minio.StatObjectOptions{})
				return err
			}},
			{name: "missing_bucket", want: "bucket_missing", errFn: func(ctx context.Context) error {
				_, err := r.client.StatObject(ctx, r.cfg.Bucket+"-missing-"+r.probeID, "missing.bin", minio.StatObjectOptions{})
				return err
			}},
			{name: "unreachable_endpoint", want: "endpoint_unreachable", errFn: func(ctx context.Context) error {
				client, err := minio.New("127.0.0.1:1", &minio.Options{
					Creds:  credentials.NewStaticV4(r.cfg.AccessKeyID, r.cfg.SecretAccessKey, ""),
					Secure: false,
				})
				if err != nil {
					return err
				}
				shortCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()
				_, err = client.BucketExists(shortCtx, r.cfg.Bucket)
				return err
			}},
			{name: "integrity_mismatch", want: "integrity_mismatch", errFn: func(context.Context) error {
				return fmt.Errorf("object payload mismatch")
			}},
			{name: "cors_rejection", want: "cors_rejected", errFn: func(context.Context) error {
				return fmt.Errorf("CORS origin not allowed")
			}},
		}
		for _, check := range checks {
			err := check.errFn(ctx)
			if err == nil {
				return fmt.Errorf("%s fixture did not fail", check.name)
			}
			reason, _ := classifyReason(err)
			if reason != check.want {
				return fmt.Errorf("%s classified as %s, want %s", check.name, reason, check.want)
			}
		}
		return nil
	})
}

func (r *probeRunner) verifyNoPublicListingPrimitive(ctx context.Context) error {
	return r.stage(ctx, "public_listing_route_inventory", func(context.Context) error {
		root, err := repoRoot()
		if err != nil {
			return err
		}
		openapi, err := os.ReadFile(filepath.Join(root, "contracts/openapi/cartulary.openapi.yaml"))
		if err != nil {
			return err
		}
		lowerOpenAPI := strings.ToLower(string(openapi))
		for _, forbidden := range []string{"listobject", "list-prefix", "object-prefix", "/api/v1/object-prefix"} {
			if strings.Contains(lowerOpenAPI, forbidden) {
				return fmt.Errorf("public OpenAPI route inventory contains %q", forbidden)
			}
		}
		routes, err := os.ReadFile(filepath.Join(root, "internal/modules/evidence/routes.go"))
		if err != nil {
			return err
		}
		for _, forbidden := range []string{"GET /api/v1/object-blobs", "ListObjects(", "ListPrefix("} {
			if strings.Contains(string(routes), forbidden) {
				return fmt.Errorf("evidence routes expose object listing primitive %q", forbidden)
			}
		}
		return nil
	})
}

func (r *probeRunner) verifySameOriginHandleContract(ctx context.Context) error {
	return r.stage(ctx, "same_origin_evidence_handle_contract", func(context.Context) error {
		root, err := repoRoot()
		if err != nil {
			return err
		}
		routes, err := os.ReadFile(filepath.Join(root, "internal/modules/evidence/routes.go"))
		if err != nil {
			return err
		}
		text := string(routes)
		required := `"href": "/api/v1/evidence-handles/" + url.PathEscape(token)`
		if !strings.Contains(text, required) {
			return fmt.Errorf("evidence handle href is not same-origin opaque route")
		}
		for _, forbidden := range []string{"PresignedGetObject", "s3://", "X-Amz-Credential", "X-Amz-Signature"} {
			if strings.Contains(text, forbidden) {
				return fmt.Errorf("evidence handle route contains forbidden storage exposure marker %q", forbidden)
			}
		}
		return nil
	})
}

func (r *probeRunner) verifyCleanupClassification(ctx context.Context) error {
	return r.stage(ctx, "probe_cleanup_classification", func(context.Context) error {
		cases := []struct {
			name        string
			createdKeys map[string]bool
			deletedKeys map[string]bool
			want        string
		}{
			{name: "clean", createdKeys: map[string]bool{"k": true}, deletedKeys: map[string]bool{"k": true}, want: "clean"},
			{name: "under_reserved_prefix", createdKeys: map[string]bool{r.probePrefix + "left.bin": true}, deletedKeys: map[string]bool{}, want: "retained_under_reserved_probe_prefix"},
			{name: "outside_reserved_prefix", createdKeys: map[string]bool{"outside.bin": true}, deletedKeys: map[string]bool{}, want: "retained_outside_reserved_probe_prefix"},
		}
		for _, tc := range cases {
			synthetic := &probeRunner{
				probePrefix: tc.name,
				createdKeys: tc.createdKeys,
				deletedKeys: tc.deletedKeys,
			}
			if tc.name == "under_reserved_prefix" {
				synthetic.probePrefix = r.probePrefix
			}
			if got := synthetic.cleanupResult(); got != tc.want {
				return fmt.Errorf("%s cleanup result got %s want %s", tc.name, got, tc.want)
			}
		}
		return nil
	})
}

func (r *probeRunner) verifyCanonicalArtifactSerialization(ctx context.Context) error {
	return r.stage(ctx, "canonical_artifact_serialization", func(context.Context) error {
		sample := compatibilityReport{
			SchemaID:           compatibilitySchemaID,
			ProbeID:            "canonical-check",
			ObjectStoreBackend: objectStoreBackend,
			StartedAt:          "2026-06-04T00:00:00Z",
			CompletedAt:        "2026-06-04T00:00:01Z",
			Result:             "pass",
			Cases:              []compatibilityCase{{CaseID: "SWFS-COMP-014", Capability: "Canonical artifact serialization", Status: "pass", Evidence: map[string]any{"source": "self"}}},
			ForbiddenSkipRows:  []string{},
		}
		left, err := json.Marshal(sample)
		if err != nil {
			return err
		}
		right, err := json.Marshal(sample)
		if err != nil {
			return err
		}
		if !bytes.Equal(left, right) || sha256Hex(string(left)) != sha256Hex(string(right)) {
			return fmt.Errorf("canonical serialization digest is unstable")
		}
		if err := rejectDuplicateJSONKeys([]byte(`{"schema_id":"x","schema_id":"y"}`)); err == nil {
			return fmt.Errorf("duplicate JSON keys were not rejected")
		}
		return nil
	})
}

func (r *probeRunner) cleanupAfterFailure(ctx context.Context) error {
	remaining := make([]string, 0)
	for key := range r.createdKeys {
		if !r.deletedKeys[key] {
			remaining = append(remaining, key)
		}
	}
	if len(remaining) == 0 {
		return nil
	}
	return r.stage(ctx, "cleanup_after_failure", func(ctx context.Context) error {
		for _, key := range remaining {
			if err := r.client.RemoveObject(ctx, r.cfg.Bucket, key, minio.RemoveObjectOptions{}); err != nil {
				return err
			}
			r.deletedKeys[key] = true
		}
		return nil
	})
}

func (r *probeRunner) runReset(ctx context.Context) error {
	if err := r.stage(ctx, "bucket_validation", func(ctx context.Context) error {
		exists, err := r.client.BucketExists(ctx, r.cfg.Bucket)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		return r.client.MakeBucket(ctx, r.cfg.Bucket, minio.MakeBucketOptions{})
	}); err != nil {
		return err
	}
	if err := r.stage(ctx, "delete_bucket_objects", func(ctx context.Context) error {
		for item := range r.client.ListObjects(ctx, r.cfg.Bucket, minio.ListObjectsOptions{Recursive: true}) {
			if item.Err != nil {
				return item.Err
			}
			if err := r.client.RemoveObject(ctx, r.cfg.Bucket, item.Key, minio.RemoveObjectOptions{}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return r.stage(ctx, "bucket_present_after_reset", func(ctx context.Context) error {
		exists, err := r.client.BucketExists(ctx, r.cfg.Bucket)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("bucket missing after reset")
		}
		return nil
	})
}

func (r *probeRunner) stage(ctx context.Context, name string, fn func(context.Context) error) error {
	startedAt := time.Now().UTC().Format(time.RFC3339Nano)
	result := probeStep{Stage: name, Status: "fail", AttemptCount: 1, StartedAt: &startedAt}
	err := fn(ctx)
	completedAt := time.Now().UTC().Format(time.RFC3339Nano)
	result.CompletedAt = &completedAt
	if err != nil {
		reason, retryable := classifyReason(err)
		result.ReasonCode = &reason
		r.artifact.Steps = append(r.artifact.Steps, result)
		if r.artifact.FailedStage == nil {
			failedStage := name
			r.artifact.FailedStage = &failedStage
			r.artifact.ReasonCode = &reason
			r.artifact.Retryable = retryable
		}
		return fmt.Errorf("%s: %s", name, sanitizeMessage(r.cfg, err.Error()))
	}
	result.Status = "pass"
	r.artifact.Steps = append(r.artifact.Steps, result)
	return nil
}

func (r *probeRunner) finish(runErr error) {
	completedAt := time.Now().UTC().Format(time.RFC3339Nano)
	r.artifact.CompletedAt = completedAt
	r.artifact.CleanupResult = r.cleanupResult()
	r.artifact.RetainedProbeKeys = r.retainedKeyRefs()
	if runErr == nil && r.artifact.CleanupResult == "clean" {
		r.artifact.Result = "pass"
	} else if runErr == nil && r.artifact.CleanupResult == "retained_under_reserved_probe_prefix" && r.cfg.ProfileID == "local_dev" {
		r.artifact.Result = "pass_with_cleanup_warning"
	} else {
		r.artifact.Result = "fail"
		if r.artifact.ReasonCode == nil {
			reason := "cleanup_failed"
			r.artifact.ReasonCode = &reason
		}
	}
	r.compatibility.CompletedAt = completedAt
	r.compatibility.Cases = r.compatibilityCases()
	r.compatibility.Result = compatibilityResult(r.compatibility.Cases)
}

func (r *probeRunner) cleanupResult() string {
	retained := r.retainedKeys()
	if len(retained) == 0 {
		return "clean"
	}
	for _, key := range retained {
		if !strings.HasPrefix(key, r.probePrefix) {
			return "retained_outside_reserved_probe_prefix"
		}
	}
	return "retained_under_reserved_probe_prefix"
}

func (r *probeRunner) retainedKeys() []string {
	retained := make([]string, 0)
	for key := range r.createdKeys {
		if !r.deletedKeys[key] {
			retained = append(retained, key)
		}
	}
	return retained
}

func (r *probeRunner) retainedKeyRefs() []redactionObject {
	keys := r.retainedKeys()
	refs := make([]redactionObject, 0, len(keys))
	for _, key := range keys {
		refs = append(refs, shaRef("object_key", key))
	}
	return refs
}

func (r *probeRunner) markCompat(caseID string, status string) {
	r.compatStatus[caseID] = status
}

func (r *probeRunner) compatibilityCases() []compatibilityCase {
	cases := []compatibilityCase{
		{CaseID: "SWFS-COMP-001", Capability: "Bucket validation", Evidence: map[string]any{"source": "bucket_validation"}},
		{CaseID: "SWFS-COMP-002", Capability: "Put object", Evidence: map[string]any{"source": "put_primary,put_zero,put_small,put_large", "payload_profiles": []string{"0_bytes", "37_bytes", "1MiB_plus_13_bytes", "range_payload_4097_bytes"}}},
		{CaseID: "SWFS-COMP-003", Capability: "Head object", Evidence: map[string]any{"source": "head_primary,head_zero,head_small,head_large"}},
		{CaseID: "SWFS-COMP-004", Capability: "Full get object", Evidence: map[string]any{"source": "get_primary,get_zero,get_small,get_large"}},
		{CaseID: "SWFS-COMP-005", Capability: "Range get", Evidence: map[string]any{"source": "range_primary", "range": "[4,15]"}},
		{CaseID: "SWFS-COMP-006", Capability: "Delete object", Evidence: map[string]any{"source": "verify_primary_deleted"}},
		{CaseID: "SWFS-COMP-007", Capability: "Prefix isolation", Evidence: map[string]any{"source": "list_prefix_isolation"}},
		{CaseID: "SWFS-COMP-008", Capability: "Presigned PUT", Evidence: map[string]any{"source": "create_direct_upload_target,direct_put,create_expiring_direct_upload_target,direct_put_after_expiry_rejected", "production_expiry": "5m", "after_expiry_checked": true}},
		{CaseID: "SWFS-COMP-009", Capability: "Error classification", Evidence: map[string]any{"source": "error_classification_matrix", "fixtures": []string{"missing_object", "denied_credential", "missing_bucket", "unreachable_endpoint", "integrity_mismatch", "cors_rejection"}}},
		{CaseID: "SWFS-COMP-010", Capability: "No public listing primitive", Evidence: map[string]any{"source": "public_listing_route_inventory"}},
		{CaseID: "SWFS-COMP-011", Capability: "CORS preflight", Evidence: map[string]any{"source": "cors_preflight"}},
		{CaseID: "SWFS-COMP-012", Capability: "Same-origin preview/download", Evidence: map[string]any{"source": "same_origin_evidence_handle_contract"}},
		{CaseID: "SWFS-COMP-013", Capability: "Probe cleanup classification", Evidence: map[string]any{"source": "probe_cleanup_classification"}},
		{CaseID: "SWFS-COMP-014", Capability: "Canonical artifact serialization", Evidence: map[string]any{"source": "canonical_artifact_serialization"}},
	}
	for idx := range cases {
		status, ok := r.compatStatus[cases[idx].CaseID]
		if !ok {
			status = "not_run"
			reason := "compatibility_evidence_not_retained"
			cases[idx].ReasonCode = &reason
		}
		if cases[idx].CaseID == "SWFS-COMP-008" && status == "not_run" {
			reason := "after_expiry_path_not_retained"
			cases[idx].ReasonCode = &reason
		}
		cases[idx].Status = status
	}
	return cases
}

func compatibilityResult(cases []compatibilityCase) string {
	for _, item := range cases {
		if item.Status != "pass" {
			return "fail"
		}
	}
	return "pass"
}

func profileMayCreateBucket(profile string) bool {
	switch profile {
	case "local_dev", "ci_service_backed", "developer_debug", "test":
		return true
	default:
		return false
	}
}

func requireObjectPayload(ctx context.Context, client *minio.Client, bucket string, key string, want []byte, opts minio.GetObjectOptions) error {
	body, err := client.GetObject(ctx, bucket, key, opts)
	if err != nil {
		return err
	}
	defer body.Close()
	got, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("object payload mismatch")
	}
	return nil
}

func checkCORSPreflight(ctx context.Context, endpoint string, origin string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodOptions, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", http.MethodPut)
	req.Header.Set("Access-Control-Request-Headers", "content-type,x-amz-checksum-sha256")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("CORS preflight status %d", resp.StatusCode)
	}
	if err := validateCORSPreflightResponse(resp.Header, origin); err != nil {
		return err
	}
	if err := checkCORSDisallowedPreflightRejected(ctx, endpoint, origin, http.MethodGet, "content-type"); err != nil {
		return err
	}
	if err := checkCORSDisallowedPreflightRejected(ctx, endpoint, origin, http.MethodPut, "content-type, authorization"); err != nil {
		return err
	}
	return checkCORSNullOriginRejected(ctx, endpoint)
}

func checkCORSDisallowedPreflightRejected(ctx context.Context, endpoint string, origin string, method string, requestHeaders string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodOptions, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", method)
	if requestHeaders != "" {
		req.Header.Set("Access-Control-Request-Headers", requestHeaders)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	allowOrigin := strings.TrimSpace(resp.Header.Get("Access-Control-Allow-Origin"))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && allowOrigin != "" {
		return fmt.Errorf("CORS disallowed preflight unexpectedly granted")
	}
	if allowOrigin == "*" || allowOrigin == origin {
		return fmt.Errorf("CORS disallowed preflight unexpectedly granted")
	}
	return nil
}

func checkCORSNullOriginRejected(ctx context.Context, endpoint string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodOptions, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Origin", "null")
	req.Header.Set("Access-Control-Request-Method", http.MethodPut)
	req.Header.Set("Access-Control-Request-Headers", "content-type,x-amz-checksum-sha256")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	allowOrigin := strings.TrimSpace(resp.Header.Get("Access-Control-Allow-Origin"))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && allowOrigin != "" {
		return fmt.Errorf("CORS Origin null unexpectedly allowed")
	}
	if allowOrigin == "*" || strings.EqualFold(allowOrigin, "null") {
		return fmt.Errorf("CORS Origin null unexpectedly allowed")
	}
	return nil
}

func validateCORSPreflightResponse(header http.Header, origin string) error {
	if strings.TrimSpace(origin) == "" {
		return fmt.Errorf("CORS origin is required")
	}
	allowOrigin := strings.TrimSpace(header.Get("Access-Control-Allow-Origin"))
	if allowOrigin != origin {
		return fmt.Errorf("CORS origin must exactly match application public origin")
	}
	if strings.EqualFold(strings.TrimSpace(header.Get("Access-Control-Allow-Credentials")), "true") {
		return fmt.Errorf("CORS credentials are enabled")
	}
	if !equalHeaderTokenSet(header.Values("Access-Control-Allow-Methods"), []string{http.MethodPut, http.MethodOptions}) {
		return fmt.Errorf("CORS methods must be exactly PUT and OPTIONS")
	}
	if !equalHeaderTokenSet(header.Values("Access-Control-Allow-Headers"), []string{"content-type", "x-amz-checksum-sha256"}) {
		return fmt.Errorf("CORS request headers must be exactly content-type and x-amz-checksum-sha256")
	}
	if strings.TrimSpace(header.Get("Access-Control-Max-Age")) != "600" {
		return fmt.Errorf("CORS max age must be exactly 600 seconds")
	}
	return nil
}

func validateCORSActualPUTResponse(header http.Header, origin string) error {
	allowOrigin := strings.TrimSpace(header.Get("Access-Control-Allow-Origin"))
	if allowOrigin != origin {
		return fmt.Errorf("CORS PUT origin must exactly match application public origin")
	}
	if strings.EqualFold(strings.TrimSpace(header.Get("Access-Control-Allow-Credentials")), "true") {
		return fmt.Errorf("CORS PUT credentials are enabled")
	}
	if !equalHeaderTokenSet(header.Values("Access-Control-Expose-Headers"), []string{"etag"}) {
		return fmt.Errorf("CORS exposed response headers must be exactly etag")
	}
	return nil
}

func equalHeaderTokenSet(values []string, want []string) bool {
	gotTokens := map[string]struct{}{}
	for _, value := range values {
		for _, token := range strings.Split(value, ",") {
			normalized := strings.ToLower(strings.TrimSpace(token))
			if normalized != "" {
				gotTokens[normalized] = struct{}{}
			}
		}
	}
	if len(gotTokens) != len(want) {
		return false
	}
	for _, token := range want {
		if _, ok := gotTokens[strings.ToLower(token)]; !ok {
			return false
		}
	}
	return true
}

func uploadPresignedPUT(ctx context.Context, endpoint string, origin string, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Origin", origin)
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("presigned PUT status %d", resp.StatusCode)
	}
	return validateCORSActualPUTResponse(resp.Header, origin)
}

func requirePresignedPUTRejected(ctx context.Context, endpoint string, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return fmt.Errorf("expired presigned PUT unexpectedly succeeded")
	}
	return nil
}

func writeProbeArtifact(cfg config, report probeArtifact) error {
	artifactPath := cfg.ArtifactPath
	if artifactPath == "" {
		artifactPath = harnessArtifactPath()
	}
	if artifactPath == "" {
		return nil
	}
	return writeJSONArtifact(artifactPath, report)
}

func writeCompatibilityArtifact(cfg config, report compatibilityReport) error {
	probePath := cfg.ArtifactPath
	if probePath == "" {
		probePath = harnessArtifactPath()
	}
	if probePath == "" {
		return nil
	}
	return writeJSONArtifact(filepath.Join(filepath.Dir(probePath), "object-store-compatibility-report.json"), report)
}

func writeJSONArtifact(path string, report any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

func harnessArtifactPath() string {
	resultsDir := strings.TrimSpace(os.Getenv("CARTULARY_TEST_RESULTS_DIR"))
	runID := strings.TrimSpace(os.Getenv("CARTULARY_TEST_RUN_ID"))
	if resultsDir == "" || runID == "" {
		return ""
	}
	target := strings.TrimSpace(os.Getenv("CARTULARY_TEST_TARGET"))
	if target == "" {
		target = "object-store-probe"
	}
	return filepath.Join(resultsDir, runID, target, "object-store-capability-probe.json")
}

func sizedProbePayload(prefix string, size int) ([]byte, error) {
	if size < len(prefix) {
		return nil, fmt.Errorf("probe payload size must be at least prefix length")
	}
	payload := make([]byte, 0, size)
	payload = append(payload, []byte(prefix)...)
	index := 0
	for len(payload) < size {
		payload = append(payload, byte('a'+index%26))
		index++
	}
	return payload, nil
}

func repeatedPayload(size int) []byte {
	payload := make([]byte, size)
	for index := range payload {
		payload[index] = byte(index % 251)
	}
	return payload
}

func randomHex(size int) (string, error) {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off", "":
		return false, nil
	default:
		return false, fmt.Errorf("must be true or false")
	}
}

func envDefault(name string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func normalizeEndpoint(value string) string {
	value = strings.TrimSpace(value)
	if strings.Contains(value, "://") {
		parsed, err := url.Parse(value)
		if err == nil && parsed.Host != "" {
			return parsed.Host
		}
	}
	return strings.TrimRight(value, "/")
}

func endpointRef(cfg config) redactionObject {
	scheme := "http"
	if cfg.Secure {
		scheme = "https"
	}
	host := cfg.Endpoint
	portPresent := false
	if splitHost, splitPort, err := net.SplitHostPort(cfg.Endpoint); err == nil {
		host = splitHost
		portPresent = splitPort != ""
	} else if parsed, err := url.Parse(scheme + "://" + cfg.Endpoint); err == nil && parsed.Hostname() != "" {
		host = parsed.Hostname()
		portPresent = parsed.Port() != ""
	}
	return redactionObject{
		Redacted:       true,
		RedactionClass: "endpoint",
		Scheme:         &scheme,
		HostSHA256:     sha256Hex(host),
		PortPresent:    &portPresent,
	}
}

func shaRef(class string, value string) redactionObject {
	return redactionObject{
		Redacted:       true,
		RedactionClass: class,
		SHA256:         sha256Hex(value),
	}
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repository root not found")
		}
		dir = parent
	}
}

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := rejectDuplicateJSONValue(decoder); err != nil {
		return err
	}
	if decoder.More() {
		return fmt.Errorf("unexpected trailing JSON value")
	}
	return nil
}

func rejectDuplicateJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	switch value := token.(type) {
	case json.Delim:
		switch value {
		case '{':
			seen := map[string]struct{}{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok {
					return fmt.Errorf("object key is not a string")
				}
				if _, exists := seen[key]; exists {
					return fmt.Errorf("duplicate object key %q", key)
				}
				seen[key] = struct{}{}
				if err := rejectDuplicateJSONValue(decoder); err != nil {
					return err
				}
			}
			_, err := decoder.Token()
			return err
		case '[':
			for decoder.More() {
				if err := rejectDuplicateJSONValue(decoder); err != nil {
					return err
				}
			}
			_, err := decoder.Token()
			return err
		default:
			return fmt.Errorf("unexpected JSON delimiter")
		}
	default:
		return nil
	}
}

func classifyReason(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	if response := minio.ToErrorResponse(err); response.Code != "" {
		switch response.Code {
		case "NoSuchKey", "NoSuchObject", "NotFound":
			return "object_missing", false
		case "NoSuchBucket":
			return "bucket_missing", true
		case "AccessDenied", "InvalidAccessKeyId", "SignatureDoesNotMatch", "AllAccessDisabled":
			return "credential_denied", false
		case "NotImplemented", "MethodNotAllowed", "InvalidRequest", "InvalidArgument":
			return "capability_missing", false
		case "InvalidRange", "RequestedRangeNotSatisfiable":
			return "range_unsupported", false
		case "RequestTimeout":
			return "deadline_exceeded", true
		default:
			return "endpoint_unreachable", true
		}
	}
	lower := strings.ToLower(err.Error())
	switch {
	case strings.Contains(lower, "cors"):
		return "cors_rejected", false
	case strings.Contains(lower, "size mismatch"):
		return "integrity_mismatch", false
	case strings.Contains(lower, "payload mismatch"):
		return "integrity_mismatch", false
	case strings.Contains(lower, "forbidden"):
		return "bucket_create_forbidden", false
	case strings.Contains(lower, "deadline"), strings.Contains(lower, "timeout"):
		return "deadline_exceeded", true
	default:
		return "endpoint_unreachable", true
	}
}

func sanitizeMessage(cfg config, message string) string {
	replacer := strings.NewReplacer(
		cfg.AccessKeyID, "[REDACTED:object-store-credential]",
		cfg.SecretAccessKey, "[REDACTED:object-store-credential]",
		cfg.Bucket, "[REDACTED:object-store-bucket]",
	)
	message = replacer.Replace(message)
	if idx := strings.Index(message, ".cartulary/probes/startup/"); idx >= 0 {
		message = message[:idx] + "[REDACTED:object-store-key]"
	}
	return message
}
