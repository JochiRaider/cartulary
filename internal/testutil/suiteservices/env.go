package suiteservices

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	ActiveEnv                     = "CARTULARY_TEST_SERVICES_ACTIVE"
	SuiteIDEnv                    = "CARTULARY_TEST_SUITE_ID"
	TargetEnv                     = "CARTULARY_TEST_TARGET"
	HarnessServiceDependenciesEnv = "CARTULARY_HARNESS_SERVICE_DEPENDENCIES"
	SuiteRuntimeRootEnv           = "CARTULARY_HARNESS_SUITE_RUNTIME_ROOT"
	SuiteRuntimeLeaseIDEnv        = "CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID"
	SuiteRuntimeRunIDEnv          = "CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID"

	PGAdminDSNEnv    = "CARTULARY_PGTEST_ADMIN_DSN"
	PGDSNTemplateEnv = "CARTULARY_PGTEST_DSN_TEMPLATE"
	PGTemplateDBEnv  = "CARTULARY_PGTEST_TEMPLATE_DB"
	PGSchemaHashEnv  = "CARTULARY_PGTEST_SCHEMA_HASH"
	PostgresDSNEnv   = "CARTULARY_POSTGRES_DSN"

	S3EndpointEnv    = "CARTULARY_S3TEST_ENDPOINT"
	S3AccessKeyEnv   = "CARTULARY_S3TEST_ACCESS_KEY_ID"
	S3SecretKeyEnv   = "CARTULARY_S3TEST_SECRET_ACCESS_KEY"
	S3SecureEnv      = "CARTULARY_S3TEST_SECURE"
	S3ProbeBucketEnv = "CARTULARY_S3TEST_PROBE_BUCKET"

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

// ResolveSuiteRuntimeDir returns the private, external directory used for
// secret-capable suite state. It intentionally has no fallback beneath the
// repository or retained result tree.
func ResolveSuiteRuntimeDir(env map[string]string) (string, bool, error) {
	suiteID := SuiteID(env)
	if suiteID == "" {
		return "", false, nil
	}
	configured := strings.TrimSpace(LookupEnvValue(env, SuiteRuntimeRootEnv))
	if configured == "" || !filepath.IsAbs(configured) {
		return "", false, fmt.Errorf("%s must name an absolute external directory", SuiteRuntimeRootEnv)
	}
	root := filepath.Clean(configured)
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", false, fmt.Errorf("resolve suite runtime root: %w", err)
	}
	if canonical != root {
		return "", false, fmt.Errorf("suite runtime root must not traverse symlinks")
	}
	info, err := os.Lstat(root)
	if err != nil {
		return "", false, fmt.Errorf("inspect suite runtime root: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o700 {
		return "", false, fmt.Errorf("suite runtime root must be a non-symlink owner-only 0700 directory")
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && int(stat.Uid) != os.Getuid() {
		return "", false, fmt.Errorf("suite runtime root must be owned by the current user")
	}
	repoRoot, err := FindRepoRoot()
	if err != nil {
		return "", false, err
	}
	resultsRoot, err := ResolveResultsRoot(env)
	if err != nil {
		return "", false, err
	}
	if pathContained(repoRoot, root) || pathContained(resultsRoot, root) {
		return "", false, fmt.Errorf("suite runtime root must be outside repository and retained result roots")
	}
	if err := validateSuiteRuntimeOwner(root, env); err != nil {
		return "", false, err
	}
	privateDir := filepath.Join(root, "test-services")
	if _, err := os.Lstat(privateDir); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(privateDir, 0o700); err != nil {
			return "", false, fmt.Errorf("create private suite service directory: %w", err)
		}
	} else if err != nil {
		return "", false, fmt.Errorf("inspect private suite service directory: %w", err)
	}
	privateInfo, err := os.Lstat(privateDir)
	if err != nil || !privateInfo.IsDir() || privateInfo.Mode()&os.ModeSymlink != 0 || privateInfo.Mode().Perm() != 0o700 {
		return "", false, fmt.Errorf("private suite service directory must be a non-symlink owner-only 0700 directory")
	}
	if stat, ok := privateInfo.Sys().(*syscall.Stat_t); ok && int(stat.Uid) != os.Getuid() {
		return "", false, fmt.Errorf("private suite service directory must be owned by the current user")
	}
	return privateDir, true, nil
}

func validateSuiteRuntimeOwner(root string, env map[string]string) error {
	ownerPath := filepath.Join(root, "runtime-owner.json")
	info, err := os.Lstat(ownerPath)
	if err != nil {
		return fmt.Errorf("inspect suite runtime owner marker: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o600 {
		return fmt.Errorf("suite runtime owner marker must be a non-symlink owner-only 0600 file")
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && int(stat.Uid) != os.Getuid() {
		return fmt.Errorf("suite runtime owner marker must be owned by the current user")
	}
	file, err := os.Open(ownerPath) // #nosec G304 -- ownerPath is contained below the validated non-symlink suite root.
	if err != nil {
		return fmt.Errorf("open suite runtime owner marker: %w", err)
	}
	defer file.Close()
	var owner struct {
		SchemaID  string `json:"schema_id"`
		LeaseID   string `json:"lease_id"`
		RunID     string `json:"run_id"`
		OwnerUID  int    `json:"owner_uid"`
		CreatedAt string `json:"created_at"`
	}
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&owner); err != nil {
		return fmt.Errorf("decode suite runtime owner marker: %w", err)
	}
	expectedLease := strings.TrimSpace(LookupEnvValue(env, SuiteRuntimeLeaseIDEnv))
	if owner.SchemaID != "cartulary.harness_suite_runtime_owner.v1" ||
		expectedLease == "" || owner.LeaseID != expectedLease ||
		owner.RunID != strings.TrimSpace(LookupEnvValue(env, SuiteRuntimeRunIDEnv)) || owner.OwnerUID != os.Getuid() {
		return fmt.Errorf("suite runtime owner marker does not match the active lease")
	}
	if _, err := time.Parse(time.RFC3339Nano, owner.CreatedAt); err != nil {
		return fmt.Errorf("suite runtime owner marker has invalid creation time: %w", err)
	}
	return nil
}

func pathContained(parent string, child string) bool {
	relative, err := filepath.Rel(filepath.Clean(parent), filepath.Clean(child))
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
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
