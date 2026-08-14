package performancefixture

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	RuntimeSchemaID    = "cartulary.performance_fixture_runtime.v1"
	LargeGridProfileID = "ac043_large_grid_snapshot_v1"
	RuntimeBundleName  = "performance-fixture-runtime.json"
)

type BackgroundAccount struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RuntimeBundle struct {
	SchemaID           string              `json:"schema_id"`
	FixtureProfileID   string              `json:"fixture_profile_id"`
	SnapshotKey        string              `json:"snapshot_key"`
	BackgroundAccounts []BackgroundAccount `json:"background_accounts"`
}

func GenerateRuntimeBundle(snapshotKey string) (RuntimeBundle, error) {
	return generateRuntimeBundle(snapshotKey, rand.Reader)
}

func generateRuntimeBundle(snapshotKey string, entropy io.Reader) (RuntimeBundle, error) {
	if err := validateSnapshotKey(snapshotKey); err != nil {
		return RuntimeBundle{}, err
	}
	if entropy == nil {
		return RuntimeBundle{}, errors.New("performance fixture credential entropy is required")
	}
	bundle := RuntimeBundle{
		SchemaID:           RuntimeSchemaID,
		FixtureProfileID:   LargeGridProfileID,
		SnapshotKey:        snapshotKey,
		BackgroundAccounts: make([]BackgroundAccount, 24),
	}
	for index := range bundle.BackgroundAccounts {
		emailEntropy := make([]byte, 16)
		passwordEntropy := make([]byte, 32)
		if _, err := io.ReadFull(entropy, emailEntropy); err != nil {
			return RuntimeBundle{}, fmt.Errorf("read performance fixture email entropy: %w", err)
		}
		if _, err := io.ReadFull(entropy, passwordEntropy); err != nil {
			return RuntimeBundle{}, fmt.Errorf("read performance fixture password entropy: %w", err)
		}
		bundle.BackgroundAccounts[index] = BackgroundAccount{
			Email:    fmt.Sprintf("ac043-%s@example.test", hex.EncodeToString(emailEntropy)),
			Password: "Ac043!" + base64.RawURLEncoding.EncodeToString(passwordEntropy),
		}
	}
	return bundle, nil
}

func WriteRuntimeBundle(root string, bundle RuntimeBundle) (string, error) {
	if err := validateRuntimeBundle(bundle); err != nil {
		return "", err
	}
	if err := createPrivateRoot(root); err != nil {
		return "", err
	}
	path := filepath.Join(root, RuntimeBundleName)
	if err := writeExclusiveJSON(path, bundle); err != nil {
		_ = os.Remove(root)
		return "", err
	}
	return path, nil
}

func ReadRuntimeBundle(path string) (RuntimeBundle, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return RuntimeBundle{}, fmt.Errorf("inspect performance fixture runtime bundle: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		return RuntimeBundle{}, fmt.Errorf("performance fixture runtime bundle must be a regular 0600 file")
	}
	raw, err := os.ReadFile(path) // #nosec G304 -- caller supplies a lease-private bundle path and file identity is checked above.
	if err != nil {
		return RuntimeBundle{}, fmt.Errorf("read performance fixture runtime bundle: %w", err)
	}
	var bundle RuntimeBundle
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&bundle); err != nil {
		return RuntimeBundle{}, fmt.Errorf("decode performance fixture runtime bundle: %w", err)
	}
	if err := validateRuntimeBundle(bundle); err != nil {
		return RuntimeBundle{}, err
	}
	return bundle, nil
}

func CopyRuntimeBundle(sourcePath string, destinationRoot string) (string, error) {
	bundle, err := ReadRuntimeBundle(sourcePath)
	if err != nil {
		return "", err
	}
	return WriteRuntimeBundle(destinationRoot, bundle)
}

func RemoveRuntimeBundle(root string) error {
	if strings.TrimSpace(root) == "" || !filepath.IsAbs(root) {
		return errors.New("performance fixture runtime root must be an absolute path")
	}
	for _, name := range []string{RuntimeBundleName, RuntimeBundleName + ".tmp"} {
		if err := os.Remove(filepath.Join(root, name)); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove performance fixture runtime bundle: %w", err)
		}
	}
	if err := os.Remove(root); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove performance fixture runtime root: %w", err)
	}
	return nil
}

func ValidateReceiptRedaction(result Result, state *BuildState) error {
	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}
	text := string(payload)
	for _, account := range state.RuntimeBundle.BackgroundAccounts {
		for _, secret := range []string{account.Email, account.Password} {
			if secret != "" && strings.Contains(text, secret) {
				return errors.New("performance fixture semantic receipt contains runtime credential material")
			}
		}
	}
	for _, userID := range state.BackgroundUserIDs {
		if userID != "" && strings.Contains(text, userID) {
			return errors.New("performance fixture semantic receipt contains a background user identifier")
		}
	}
	if state.IncidentID != "" && strings.Contains(text, state.IncidentID) {
		return errors.New("performance fixture semantic receipt contains an incident identifier")
	}
	return nil
}

func validateRuntimeBundle(bundle RuntimeBundle) error {
	if bundle.SchemaID != RuntimeSchemaID || bundle.FixtureProfileID != LargeGridProfileID {
		return errors.New("performance fixture runtime bundle identity is unsupported")
	}
	if err := validateSnapshotKey(bundle.SnapshotKey); err != nil {
		return err
	}
	if len(bundle.BackgroundAccounts) != 24 {
		return fmt.Errorf("performance fixture runtime bundle has %d background accounts, want 24", len(bundle.BackgroundAccounts))
	}
	seen := map[string]struct{}{}
	for index, account := range bundle.BackgroundAccounts {
		if !strings.Contains(account.Email, "@") || len(account.Email) > 254 || len(account.Password) < 24 || len(account.Password) > 256 {
			return fmt.Errorf("performance fixture runtime account %d is malformed", index)
		}
		if _, exists := seen[account.Email]; exists {
			return fmt.Errorf("performance fixture runtime account %d duplicates an email", index)
		}
		seen[account.Email] = struct{}{}
	}
	return nil
}

func createPrivateRoot(root string) error {
	if strings.TrimSpace(root) == "" || !filepath.IsAbs(root) {
		return errors.New("performance fixture runtime root must be an absolute path")
	}
	if err := os.Mkdir(root, 0o700); err != nil {
		return fmt.Errorf("create performance fixture runtime root: %w", err)
	}
	info, err := os.Lstat(root)
	if err != nil {
		return fmt.Errorf("inspect performance fixture runtime root: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o700 {
		return errors.New("performance fixture runtime root must be a private 0700 directory")
	}
	return nil
}

func writeExclusiveJSON(path string, value any) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("performance fixture runtime bundle already exists")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect performance fixture runtime bundle destination: %w", err)
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode performance fixture runtime bundle: %w", err)
	}
	temporary := path + ".tmp"
	file, err := os.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) // #nosec G304 -- the path is below an already validated private runtime root.
	if err != nil {
		return fmt.Errorf("create performance fixture runtime bundle: %w", err)
	}
	succeeded := false
	defer func() {
		_ = file.Close()
		if !succeeded {
			_ = os.Remove(temporary)
		}
	}()
	if _, err := file.Write(append(payload, '\n')); err != nil {
		return fmt.Errorf("write performance fixture runtime bundle: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync performance fixture runtime bundle: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close performance fixture runtime bundle: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("publish performance fixture runtime bundle: %w", err)
	}
	succeeded = true
	return nil
}
