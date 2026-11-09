package producer

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"runtime/debug"
)

func randomSuffix() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type Builder struct {
	ID        string `json:"builder_id"`
	Version   string `json:"builder_version"`
	SHA256    string `json:"builder_source_sha256"`
	Size      int64  `json:"size_bytes"`
	GoVersion string `json:"go_version"`
}

func BuilderIdentity() (Builder, error) {
	// The Linux producer identifies the running image, even if its pathname
	// has been replaced or unlinked since launch.
	file, err := os.Open("/proc/self/exe")
	if err != nil {
		return Builder{}, err
	}
	defer file.Close()
	b, err := io.ReadAll(file)
	if err != nil {
		return Builder{}, err
	}
	version := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		version = info.GoVersion
	}
	return Builder{ID: "cartulary.reference-pack.attack", Version: "dataset-v1", SHA256: Digest(b), Size: int64(len(b)), GoVersion: version}, nil
}
