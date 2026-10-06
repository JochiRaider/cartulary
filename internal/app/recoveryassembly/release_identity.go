package recoveryassembly

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

// Packaged operations bind the complete immutable release manifest after checking
// its executable receipt. Unpackaged owner tests bind their actual executable.
func OperatorReleaseIdentity() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return operatorReleaseIdentity(executable, "/etc/cartulary/release-manifest.json")
}
func operatorReleaseIdentity(executablePath, manifestPath string) (string, error) {
	root, err := rootedfs.Open(filepath.Dir(executablePath))
	if err != nil {
		return "", err
	}
	defer root.Close()
	executable, _, err := root.OpenRegular(rootedfs.MustParseReference(filepath.Base(executablePath)))
	if err != nil {
		return "", err
	}
	defer executable.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, executable); err != nil {
		return "", err
	}
	executableDigest := hex.EncodeToString(hash.Sum(nil))
	manifestRoot, err := rootedfs.Open(filepath.Dir(manifestPath))
	if errors.Is(err, os.ErrNotExist) {
		return executableDigest, nil
	}
	if err != nil {
		return "", err
	}
	defer manifestRoot.Close()
	manifestFile, _, err := manifestRoot.OpenRegular(rootedfs.MustParseReference(filepath.Base(manifestPath)))
	if errors.Is(err, os.ErrNotExist) {
		return executableDigest, nil
	}
	if err != nil {
		return "", err
	}
	defer manifestFile.Close()
	body, err := io.ReadAll(io.LimitReader(manifestFile, (16<<20)+1))
	if err != nil {
		return "", err
	}
	if len(body) > 16<<20 {
		return "", errors.New("release manifest exceeds bound")
	}
	if _, err := canonicaljson.Canonicalize(body); err != nil {
		return "", err
	}
	var manifest struct {
		SchemaID string          `json:"schema_id"`
		Platform string          `json:"platform"`
		Assets   json.RawMessage `json:"assets"`
		Images   json.RawMessage `json:"images"`
		Binaries map[string]struct {
			SHA256  string `json:"sha256"`
			Receipt string `json:"receipt"`
		} `json:"binaries"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return "", err
	}
	if manifest.SchemaID != "cartulary.local_release_manifest.v1" || manifest.Platform != "linux/amd64" || manifest.Binaries["operator"].SHA256 != executableDigest || len(manifest.Binaries) != 3 {
		return "", errors.New("release executable identity mismatch")
	}
	digest := sha256.Sum256(body)
	return hex.EncodeToString(digest[:]), nil
}
