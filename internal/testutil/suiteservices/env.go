package suiteservices

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	ActiveEnv  = "CARTULARY_TEST_SERVICES_ACTIVE"
	SuiteIDEnv = "CARTULARY_TEST_SUITE_ID"
	TargetEnv  = "CARTULARY_TEST_TARGET"

	PGAdminDSNEnv    = "CARTULARY_PGTEST_ADMIN_DSN"
	PGDSNTemplateEnv = "CARTULARY_PGTEST_DSN_TEMPLATE"
	PGTemplateDBEnv  = "CARTULARY_PGTEST_TEMPLATE_DB"
	PGSchemaHashEnv  = "CARTULARY_PGTEST_SCHEMA_HASH"

	S3EndpointEnv  = "CARTULARY_S3TEST_ENDPOINT"
	S3AccessKeyEnv = "CARTULARY_S3TEST_ACCESS_KEY_ID"
	S3SecretKeyEnv = "CARTULARY_S3TEST_SECRET_ACCESS_KEY"
	S3SecureEnv    = "CARTULARY_S3TEST_SECURE"

	testResultsDirEnv = "CARTULARY_TEST_RESULTS_DIR"
	testRunIDEnv      = "CARTULARY_TEST_RUN_ID"
)

func LookupEnv(env map[string]string, key string) (string, bool) {
	if env != nil {
		value, ok := env[key]
		return value, ok
	}
	return os.LookupEnv(key)
}

func LookupEnvValue(env map[string]string, key string) string {
	value, _ := LookupEnv(env, key)
	return value
}

func SuiteActive(env map[string]string) bool {
	return strings.TrimSpace(LookupEnvValue(env, ActiveEnv)) == "1"
}

func SuiteID(env map[string]string) string {
	return strings.TrimSpace(LookupEnvValue(env, SuiteIDEnv))
}

func SuiteHash(env map[string]string) string {
	return ShortHash(SuiteID(env), 8)
}

func ProcessHash() string {
	return ShortHash(fmt.Sprintf("%s|%d", strings.Join(os.Args, "\x1f"), os.Getpid()), 8)
}

func ShortHash(value string, length int) string {
	if length <= 0 {
		return ""
	}
	if strings.TrimSpace(value) == "" {
		value = "cartulary"
	}
	sum := sha1.Sum([]byte(value))
	encoded := hex.EncodeToString(sum[:])
	if length >= len(encoded) {
		return encoded
	}
	return encoded[:length]
}

func RelativeResultsRoot(env map[string]string) string {
	value := strings.TrimSpace(LookupEnvValue(env, testResultsDirEnv))
	if value != "" {
		return value
	}
	return filepath.Join(".cartulary", "test-results")
}

func ResolveResultsRoot(env map[string]string) (string, error) {
	configured := RelativeResultsRoot(env)
	if filepath.IsAbs(configured) {
		if err := validateResultsRootSecurity(configured, true); err != nil {
			return "", err
		}
		return configured, nil
	}

	repoRoot, err := FindRepoRoot()
	if err != nil {
		return "", err
	}
	resolved := filepath.Join(repoRoot, configured)
	if err := validateResultsRootSecurity(resolved, false); err != nil {
		return "", err
	}
	return resolved, nil
}

func validateResultsRootSecurity(resolved string, custom bool) error {
	existing := resolved
	if _, err := os.Stat(existing); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat result root %s: %w", resolved, err)
		}
		existing = filepath.Dir(resolved)
	}
	info, err := os.Stat(existing)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat result root parent %s: %w", existing, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("result root parent %s is not a directory", existing)
	}
	if custom && info.Mode().Perm()&0o002 != 0 && info.Mode()&os.ModeSticky == 0 {
		return fmt.Errorf("result root %s must not be world-writable without sticky bit", existing)
	}
	if existing == resolved && !custom {
		if err := os.Chmod(resolved, 0o700); err != nil {
			return fmt.Errorf("narrow result root permissions for %s: %w", resolved, err)
		}
	}
	return nil
}

func ResolveRunID(env map[string]string) string {
	value := strings.TrimSpace(LookupEnvValue(env, testRunIDEnv))
	if value != "" {
		return value
	}
	return "adhoc"
}

func ResolveSuiteArtifactDir(env map[string]string) (string, bool, error) {
	suiteID := SuiteID(env)
	if suiteID == "" {
		return "", false, nil
	}

	resultsRoot, err := ResolveResultsRoot(env)
	if err != nil {
		return "", false, err
	}

	artifactDir, err := resultsRootSubdir(resultsRoot, ResolveRunID(env), "_shared", "test-services", suiteID)
	if err != nil {
		return "", false, err
	}

	return artifactDir, true, nil
}

func resultsRootSubdir(resultsRoot string, parts ...string) (string, error) {
	cleanRoot := filepath.Clean(resultsRoot)
	for _, part := range parts {
		if filepath.IsAbs(part) {
			return "", fmt.Errorf("results artifact path component %q must be relative", part)
		}
	}
	target := filepath.Join(append([]string{cleanRoot}, parts...)...)
	relative, err := filepath.Rel(cleanRoot, target)
	if err != nil {
		return "", fmt.Errorf("resolve results artifact path: %w", err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("results artifact path %q escapes results root %q", target, cleanRoot)
	}
	return target, nil
}

func FindRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	for current := wd; ; current = filepath.Dir(current) {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("resolve repo root from %s: go.mod not found", wd)
		}
	}
}

func ParseBool(value string) (bool, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false, nil
	}
	parsed, err := strconv.ParseBool(trimmed)
	if err != nil {
		return false, fmt.Errorf("parse bool %q: %w", value, err)
	}
	return parsed, nil
}
