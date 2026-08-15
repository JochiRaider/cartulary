package performancefixturelifecycle

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func templateName(suiteID string, key string) string {
	if !lowerHexDigest(key) {
		return ""
	}
	return "ct_pfs_" + suiteservices.ShortHash(suiteID, 8) + "_" + key[:12]
}

func templateOwner(suiteID string, key string) string {
	return "cartulary.performance-fixture:" + suiteservices.ShortHash(suiteID, 16) + ":" + key
}

func templateRuntimeRoot(env map[string]string, key string) (string, error) {
	privateRoot, ok, err := suiteservices.ResolveSuiteRuntimeDir(env)
	if err != nil {
		return "", err
	}
	if !ok || !lowerHexDigest(key) {
		return "", errors.New("performance fixture requires a private suite runtime directory and canonical snapshot key")
	}
	return filepath.Join(privateRoot, "performance-fixtures", "templates", suiteservices.ShortHash(suiteservices.SuiteID(env), 16), key), nil
}

func cloneRuntimeRoot(env map[string]string, leaseID string) (string, error) {
	privateRoot, ok, err := suiteservices.ResolveSuiteRuntimeDir(env)
	if err != nil {
		return "", err
	}
	if !ok || strings.TrimSpace(leaseID) == "" {
		return "", errors.New("performance fixture clone requires a private suite runtime directory and lease identity")
	}
	return filepath.Join(privateRoot, "performance-fixtures", "clones", suiteservices.ShortHash(suiteservices.SuiteID(env), 16), suiteservices.ShortHash(leaseID, 24)), nil
}

func removeTemplateRuntime(env map[string]string, key string) error {
	runtimeRoot, err := templateRuntimeRoot(env, key)
	if err != nil {
		return err
	}
	return performancefixture.RemoveRuntimeBundle(runtimeRoot)
}

func newCloneName() (string, error) {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", err
	}
	return "ct_" + hex.EncodeToString(data[:]) + "_web_e2e", nil
}

func lowerHexDigest(value string) bool {
	if len(value) != 64 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func safePostgresIdentifier(value string) bool {
	if value == "" || len(value) > 63 {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			continue
		}
		return false
	}
	return true
}

func safeCatalogIdentity(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' || char == '.' {
			continue
		}
		return false
	}
	return true
}

func WriteImmutableJSON(file string, value any) error {
	if !filepath.IsAbs(file) {
		return errors.New("immutable artifact path must be absolute")
	}
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		return err
	}
	handle, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) // #nosec G304 -- caller supplies a canonical run-relative artifact path.
	if err != nil {
		return err
	}
	succeeded := false
	defer func() {
		_ = handle.Close()
		if !succeeded {
			_ = os.Remove(file)
		}
	}()
	if _, err := handle.Write(append(payload, '\n')); err != nil {
		return err
	}
	if err := handle.Sync(); err != nil {
		return err
	}
	if err := handle.Close(); err != nil {
		return err
	}
	succeeded = true
	return nil
}
