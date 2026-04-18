package objectstore

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    bool
	Bucket    string
}

type ServiceRefEnvKeys struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    string
	Bucket    string
}

func Setup(ctx context.Context, cfg config.Config) (*minio.Client, error) {
	return SetupWithEnv(ctx, cfg, nil)
}

func SetupWithEnv(ctx context.Context, cfg config.Config, env map[string]string) (*minio.Client, error) {
	settings, err := ResolveSettings(cfg, env)
	if err != nil {
		return nil, err
	}

	client, err := minio.New(settings.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(settings.AccessKey, settings.SecretKey, ""),
		Secure: settings.Secure,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	exists, err := client.BucketExists(ctx, settings.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check object store bucket %q: %w", settings.Bucket, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, settings.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create object store bucket %q: %w", settings.Bucket, err)
		}
	}

	return client, nil
}

func ResolveSettings(cfg config.Config, env map[string]string) (Settings, error) {
	switch cfg.Roots.ObjectStorage.BindingKind {
	case "filesystem_root":
		return resolveFilesystemRootSettings(env)
	case "managed_service":
		return resolveManagedServiceSettings(cfg.Roots.ObjectStorage.ServiceRef, env)
	default:
		return Settings{}, fmt.Errorf("resolve object-store settings: roots.object_storage.binding_kind must be configured before object-store setup")
	}
}

func resolveFilesystemRootSettings(env map[string]string) (Settings, error) {
	settings := Settings{
		Endpoint:  "localhost:9000",
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		Secure:    false,
		Bucket:    "cartulary",
	}

	if value, ok := lookupEnv(env, EndpointEnv); ok && value != "" {
		settings.Endpoint = value
	}
	if value, ok := lookupEnv(env, AccessKeyEnv); ok && value != "" {
		settings.AccessKey = value
	}
	if value, ok := lookupEnv(env, SecretKeyEnv); ok && value != "" {
		settings.SecretKey = value
	}
	if value, ok := lookupEnv(env, BucketEnv); ok && value != "" {
		settings.Bucket = value
	}
	if value, ok := lookupEnv(env, SecureEnv); ok && value != "" {
		secure, err := strconv.ParseBool(value)
		if err != nil {
			return Settings{}, fmt.Errorf("parse %s: %w", SecureEnv, err)
		}
		settings.Secure = secure
	}

	return settings, nil
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
	normalized := normalizeServiceRef(serviceRef)
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

func normalizeServiceRef(value string) string {
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
