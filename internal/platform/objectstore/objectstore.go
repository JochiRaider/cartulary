package objectstore

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"example.com/todo/cartulary/internal/platform/config"
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

func Setup(ctx context.Context, cfg config.Config) (*minio.Client, error) {
	return SetupWithEnv(ctx, cfg, nil)
}

func SetupWithEnv(ctx context.Context, cfg config.Config, env map[string]string) (*minio.Client, error) {
	_ = ctx

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

	// TODO: derive endpoint and credentials from deployment configuration instead of fixed local defaults.
	return client, nil
}

func ResolveSettings(cfg config.Config, env map[string]string) (Settings, error) {
	_ = cfg

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

func lookupEnv(env map[string]string, key string) (string, bool) {
	if env != nil {
		value, ok := env[key]
		return value, ok
	}

	return os.LookupEnv(key)
}
