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

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
)

const RuntimeBundleName = "performance-fixture-runtime.json"

type RuntimeCredential struct {
	Principal string `json:"principal"`
	Secret    string `json:"secret"`
}

type RuntimeCredentialSet struct {
	SetID          string              `json:"set_id"`
	CredentialKind string              `json:"credential_kind"`
	Credentials    []RuntimeCredential `json:"credentials"`
}

type RuntimeBundle struct {
	SchemaID         string                 `json:"schema_id"`
	FixtureProfileID string                 `json:"fixture_profile_id"`
	SnapshotKey      string                 `json:"snapshot_key"`
	CredentialSets   []RuntimeCredentialSet `json:"credential_sets"`
}

func (bundle RuntimeBundle) Credentials(setID string) ([]RuntimeCredential, bool) {
	for _, set := range bundle.CredentialSets {
		if set.SetID == setID {
			return append([]RuntimeCredential(nil), set.Credentials...), true
		}
	}
	return nil, false
}

func GenerateRuntimeBundle(profile performancefixtureprofile.Profile, snapshotKey string) (RuntimeBundle, error) {
	return generateRuntimeBundle(profile, snapshotKey, rand.Reader)
}

func generateRuntimeBundle(profile performancefixtureprofile.Profile, snapshotKey string, entropy io.Reader) (RuntimeBundle, error) {
	if err := validateSnapshotKey(snapshotKey); err != nil {
		return RuntimeBundle{}, err
	}
	if entropy == nil {
		return RuntimeBundle{}, errors.New("performance fixture credential entropy is required")
	}
	bundle := RuntimeBundle{
		SchemaID:         profile.ArtifactPolicy.RuntimeSchemaID,
		FixtureProfileID: profile.FixtureProfileID,
		SnapshotKey:      snapshotKey,
		CredentialSets:   make([]RuntimeCredentialSet, len(profile.RuntimeCredentialSets)),
	}
	for setIndex, descriptor := range profile.RuntimeCredentialSets {
		if descriptor.CredentialKind != "email_password" || strings.TrimSpace(descriptor.SetID) == "" || descriptor.AccountCount < 1 {
			return RuntimeBundle{}, errors.New("performance fixture runtime credential set is unsupported")
		}
		set := RuntimeCredentialSet{
			SetID:          descriptor.SetID,
			CredentialKind: descriptor.CredentialKind,
			Credentials:    make([]RuntimeCredential, descriptor.AccountCount),
		}
		for credentialIndex := range set.Credentials {
			principalEntropy := make([]byte, 16)
			secretEntropy := make([]byte, 32)
			if _, err := io.ReadFull(entropy, principalEntropy); err != nil {
				return RuntimeBundle{}, fmt.Errorf("read performance fixture principal entropy: %w", err)
			}
			if _, err := io.ReadFull(entropy, secretEntropy); err != nil {
				return RuntimeBundle{}, fmt.Errorf("read performance fixture secret entropy: %w", err)
			}
			set.Credentials[credentialIndex] = RuntimeCredential{
				Principal: fmt.Sprintf("fixture-%s@example.test", hex.EncodeToString(principalEntropy)),
				Secret:    "Fixture!" + base64.RawURLEncoding.EncodeToString(secretEntropy),
			}
		}
		bundle.CredentialSets[setIndex] = set
	}
	return bundle, nil
}

func WriteRuntimeBundle(profile performancefixtureprofile.Profile, root string, bundle RuntimeBundle) (string, error) {
	if err := ValidateRuntimeBundle(profile, bundle); err != nil {
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

func ReadRuntimeBundle(profile performancefixtureprofile.Profile, path string) (RuntimeBundle, error) {
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
	if err := ValidateRuntimeBundle(profile, bundle); err != nil {
		return RuntimeBundle{}, err
	}
	return bundle, nil
}

func CopyRuntimeBundle(profile performancefixtureprofile.Profile, sourcePath string, destinationRoot string) (string, error) {
	bundle, err := ReadRuntimeBundle(profile, sourcePath)
	if err != nil {
		return "", err
	}
	return WriteRuntimeBundle(profile, destinationRoot, bundle)
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
	for _, set := range state.RuntimeBundle.CredentialSets {
		for _, credential := range set.Credentials {
			for _, secret := range []string{credential.Principal, credential.Secret} {
				if secret != "" && strings.Contains(text, secret) {
					return errors.New("performance fixture semantic receipt contains runtime credential material")
				}
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

func ValidateRuntimeBundle(profile performancefixtureprofile.Profile, bundle RuntimeBundle) error {
	if bundle.SchemaID != profile.ArtifactPolicy.RuntimeSchemaID || bundle.FixtureProfileID != profile.FixtureProfileID {
		return errors.New("performance fixture runtime bundle identity is unsupported")
	}
	if err := validateSnapshotKey(bundle.SnapshotKey); err != nil {
		return err
	}
	if len(bundle.CredentialSets) != len(profile.RuntimeCredentialSets) {
		return fmt.Errorf("performance fixture runtime bundle has %d credential sets, want %d", len(bundle.CredentialSets), len(profile.RuntimeCredentialSets))
	}
	seenPrincipals := map[string]struct{}{}
	seenSets := map[string]struct{}{}
	for setIndex, set := range bundle.CredentialSets {
		descriptor := profile.RuntimeCredentialSets[setIndex]
		if set.SetID != descriptor.SetID || set.CredentialKind != descriptor.CredentialKind || len(set.Credentials) != descriptor.AccountCount {
			return fmt.Errorf("performance fixture runtime credential set %d is inconsistent", setIndex)
		}
		if _, exists := seenSets[set.SetID]; exists {
			return fmt.Errorf("performance fixture runtime credential set %d is duplicated", setIndex)
		}
		seenSets[set.SetID] = struct{}{}
		for credentialIndex, credential := range set.Credentials {
			if !strings.Contains(credential.Principal, "@") || len(credential.Principal) > 254 || len(credential.Secret) < 24 || len(credential.Secret) > 4096 {
				return fmt.Errorf("performance fixture runtime credential %d in set %d is malformed", credentialIndex, setIndex)
			}
			if _, exists := seenPrincipals[credential.Principal]; exists {
				return fmt.Errorf("performance fixture runtime credential %d in set %d duplicates a principal", credentialIndex, setIndex)
			}
			seenPrincipals[credential.Principal] = struct{}{}
		}
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
