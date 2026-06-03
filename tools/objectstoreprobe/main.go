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
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const schemaID = "cartulary.object_store_capability_probe.v1"

type config struct {
	Mode            string
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

type artifact struct {
	SchemaID      string        `json:"schema_id"`
	Result        string        `json:"result"`
	Mode          string        `json:"mode"`
	Backend       string        `json:"backend"`
	ServiceName   string        `json:"service_name"`
	EndpointKind  string        `json:"endpoint_kind"`
	Endpoint      string        `json:"endpoint"`
	BucketSHA256  string        `json:"bucket_sha256"`
	Image         string        `json:"image"`
	ImageDigest   string        `json:"image_digest"`
	ProbeObjectID string        `json:"probe_object_id,omitempty"`
	StartedAt     string        `json:"started_at"`
	CompletedAt   string        `json:"completed_at"`
	Stages        []stageResult `json:"stages"`
	Error         string        `json:"error,omitempty"`
}

type stageResult struct {
	Name        string `json:"name"`
	Result      string `json:"result"`
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
	Message     string `json:"message,omitempty"`
}

type probeRunner struct {
	cfg             config
	client          *minio.Client
	artifact        artifact
	probeKey        string
	directUploadKey string
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
	if runErr != nil {
		runner.artifact.Result = "fail"
		runner.artifact.Error = sanitizeMessage(cfg, runErr.Error())
	} else {
		runner.artifact.Result = "pass"
	}
	runner.artifact.CompletedAt = time.Now().UTC().Format(time.RFC3339Nano)

	if err := writeArtifact(cfg, runner.artifact); err != nil {
		fmt.Fprintf(os.Stderr, "object-store probe artifact error: %v\n", err)
		os.Exit(11)
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
	flag.StringVar(&cfg.Endpoint, "endpoint", envDefault("OBJECT_STORE_ENDPOINT", "127.0.0.1:8333"), "S3 endpoint without scheme")
	flag.StringVar(&cfg.AccessKeyID, "access-key-id", envDefault("SEAWEEDFS_S3_ACCESS_KEY_ID", "cartulary-local"), "S3 access key")
	flag.StringVar(&cfg.SecretAccessKey, "secret-access-key", envDefault("SEAWEEDFS_S3_SECRET_ACCESS_KEY", "cartulary-local-secret"), "S3 secret key")
	flag.StringVar(&secureText, "secure", envDefault("OBJECT_STORE_SECURE", "false"), "whether to use HTTPS")
	flag.StringVar(&cfg.Bucket, "bucket", envDefault("OBJECT_STORE_BUCKET", "cartulary"), "S3 bucket")
	flag.StringVar(&cfg.Origin, "origin", envDefault("OBJECT_STORE_CORS_ORIGIN", "http://127.0.0.1:5173"), "browser origin for CORS preflight")
	flag.StringVar(&cfg.ServiceName, "service-name", envDefault("OBJECT_STORE_SERVICE_NAME", "seaweedfs-s3"), "local object-store service name")
	flag.StringVar(&cfg.Image, "image", envDefault("SEAWEEDFS_S3_IMAGE", "docker.io/chrislusf/seaweedfs:4.17"), "object-store image tag")
	flag.StringVar(&cfg.ImageDigest, "image-digest", envDefault("SEAWEEDFS_S3_IMAGE_DIGEST", "sha256:186de7ef977a20343ee9a5544073f081976a29e2d29ecf8379891e7bf177fbe9"), "object-store image digest")
	flag.DurationVar(&cfg.Timeout, "timeout", 20*time.Second, "probe timeout")
	flag.StringVar(&cfg.ArtifactPath, "artifact", "", "artifact path; defaults to harness result path when available")
	flag.Parse()

	cfg.Mode = strings.TrimSpace(cfg.Mode)
	if cfg.Mode != "probe" && cfg.Mode != "reset" {
		return config{}, fmt.Errorf("mode must be probe or reset")
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
	id, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	probeKey := "cartulary/local-probe/" + id + "/primary"
	directUploadKey := "cartulary/local-probe/" + id + "/direct-put"

	return &probeRunner{
		cfg:             cfg,
		client:          client,
		probeKey:        probeKey,
		directUploadKey: directUploadKey,
		artifact: artifact{
			SchemaID:      schemaID,
			Result:        "fail",
			Mode:          cfg.Mode,
			Backend:       "seaweedfs_s3",
			ServiceName:   cfg.ServiceName,
			EndpointKind:  "s3",
			Endpoint:      artifactEndpoint(cfg),
			BucketSHA256:  sha256Text(cfg.Bucket),
			Image:         cfg.Image,
			ImageDigest:   cfg.ImageDigest,
			ProbeObjectID: sha256Text(probeKey + "\n" + directUploadKey),
			StartedAt:     startedAt.Format(time.RFC3339Nano),
			Stages:        []stageResult{},
		},
	}, nil
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

func (r *probeRunner) runProbe(ctx context.Context) error {
	primaryPayload, err := probePayload("cartulary-object-store-probe-v1\n")
	if err != nil {
		return err
	}
	directPayload, err := probePayload("cartulary-object-store-direct-put-probe-v1\n")
	if err != nil {
		return err
	}

	if err := r.stage(ctx, "bucket_create_or_exists", func(ctx context.Context) error {
		return ensureBucket(ctx, r.client, r.cfg.Bucket)
	}); err != nil {
		return err
	}
	if err := r.stage(ctx, "put_object", func(ctx context.Context) error {
		_, err := r.client.PutObject(ctx, r.cfg.Bucket, r.probeKey, bytes.NewReader(primaryPayload), int64(len(primaryPayload)), minio.PutObjectOptions{ContentType: "application/octet-stream"})
		return err
	}); err != nil {
		return err
	}
	if err := r.stage(ctx, "head_object", func(ctx context.Context) error {
		info, err := r.client.StatObject(ctx, r.cfg.Bucket, r.probeKey, minio.StatObjectOptions{})
		if err != nil {
			return err
		}
		if info.Size != int64(len(primaryPayload)) {
			return fmt.Errorf("object size mismatch")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := r.stage(ctx, "get_object", func(ctx context.Context) error {
		body, err := r.client.GetObject(ctx, r.cfg.Bucket, r.probeKey, minio.GetObjectOptions{})
		if err != nil {
			return err
		}
		defer body.Close()
		got, err := io.ReadAll(body)
		if err != nil {
			return err
		}
		if !bytes.Equal(got, primaryPayload) {
			return fmt.Errorf("object payload mismatch")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := r.stage(ctx, "range_read", func(ctx context.Context) error {
		opts := minio.GetObjectOptions{}
		if err := opts.SetRange(0, 15); err != nil {
			return err
		}
		body, err := r.client.GetObject(ctx, r.cfg.Bucket, r.probeKey, opts)
		if err != nil {
			return err
		}
		defer body.Close()
		got, err := io.ReadAll(body)
		if err != nil {
			return err
		}
		if !bytes.Equal(got, primaryPayload[:16]) {
			return fmt.Errorf("range payload mismatch")
		}
		return nil
	}); err != nil {
		return err
	}

	var putURL *url.URL
	if err := r.stage(ctx, "presigned_put_url", func(ctx context.Context) error {
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
	if err := r.stage(ctx, "presigned_put_upload", func(ctx context.Context) error {
		return uploadPresignedPUT(ctx, putURL.String(), directPayload)
	}); err != nil {
		return err
	}
	if err := r.stage(ctx, "direct_put_head_object", func(ctx context.Context) error {
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
	return r.stage(ctx, "delete_probe_objects", func(ctx context.Context) error {
		if err := r.client.RemoveObject(ctx, r.cfg.Bucket, r.probeKey, minio.RemoveObjectOptions{}); err != nil {
			return err
		}
		return r.client.RemoveObject(ctx, r.cfg.Bucket, r.directUploadKey, minio.RemoveObjectOptions{})
	})
}

func (r *probeRunner) runReset(ctx context.Context) error {
	if err := r.stage(ctx, "bucket_create_or_exists", func(ctx context.Context) error {
		return ensureBucket(ctx, r.client, r.cfg.Bucket)
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
	result := stageResult{Name: name, Result: "fail", StartedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	err := fn(ctx)
	result.CompletedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if err != nil {
		result.Message = sanitizeMessage(r.cfg, err.Error())
		r.artifact.Stages = append(r.artifact.Stages, result)
		return fmt.Errorf("%s: %s", name, result.Message)
	}
	result.Result = "pass"
	r.artifact.Stages = append(r.artifact.Stages, result)
	return nil
}

func ensureBucket(ctx context.Context, client *minio.Client, bucket string) error {
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
}

func checkCORSPreflight(ctx context.Context, endpoint string, origin string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodOptions, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", http.MethodPut)
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("CORS preflight status %d", resp.StatusCode)
	}
	allowOrigin := strings.TrimSpace(resp.Header.Get("Access-Control-Allow-Origin"))
	if allowOrigin != "*" && allowOrigin != origin {
		return fmt.Errorf("CORS origin not allowed")
	}
	if strings.EqualFold(strings.TrimSpace(resp.Header.Get("Access-Control-Allow-Credentials")), "true") {
		return fmt.Errorf("CORS credentials are enabled")
	}
	methods := strings.ToUpper(resp.Header.Get("Access-Control-Allow-Methods"))
	if !strings.Contains(methods, http.MethodPut) {
		return fmt.Errorf("CORS PUT method not allowed")
	}
	return nil
}

func uploadPresignedPUT(ctx context.Context, endpoint string, payload []byte) error {
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
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("presigned PUT status %d", resp.StatusCode)
	}
	return nil
}

func writeArtifact(cfg config, report artifact) error {
	artifactPath := cfg.ArtifactPath
	if artifactPath == "" {
		artifactPath = harnessArtifactPath()
	}
	if artifactPath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(artifactPath), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(artifactPath, data, 0o600)
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

func probePayload(prefix string) ([]byte, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return nil, err
	}
	payload := []byte(prefix)
	payload = append(payload, random...)
	return payload, nil
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

func artifactEndpoint(cfg config) string {
	scheme := "http"
	if cfg.Secure {
		scheme = "https"
	}
	return scheme + "://" + cfg.Endpoint
}

func sha256Text(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func sanitizeMessage(cfg config, message string) string {
	replacer := strings.NewReplacer(
		cfg.AccessKeyID, "[REDACTED:object-store-credential]",
		cfg.SecretAccessKey, "[REDACTED:object-store-credential]",
		cfg.Bucket, "[REDACTED:object-store-bucket]",
	)
	message = replacer.Replace(message)
	for _, token := range []string{"cartulary/local-probe/"} {
		if idx := strings.Index(message, token); idx >= 0 {
			message = message[:idx] + "[REDACTED:object-store-key]"
		}
	}
	return message
}
