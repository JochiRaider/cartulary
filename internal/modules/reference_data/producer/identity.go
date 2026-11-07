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
	name, err := os.Executable()
	if err != nil {
		return Builder{}, err
	}
	file, err := os.Open(name)
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
	return Builder{ID: "cartulary.attack.source-producer", Version: "native-v2", SHA256: Digest(b), Size: int64(len(b)), GoVersion: version}, nil
}
