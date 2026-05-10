package pgschema

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
)

const (
	runnerIdentity = "cartulary-postgres-migrate/goose/v3.27.0"
)

// Hash returns a stable hash of the migration inputs that define the current
// test schema template.
func Hash() (string, error) {
	entries := make([]string, 0)
	if err := fs.WalkDir(dbmigrations.Files, dbmigrations.EmbeddedPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			return nil
		}
		entries = append(entries, path)
		return nil
	}); err != nil {
		return "", fmt.Errorf("walk embedded migrations: %w", err)
	}
	sort.Strings(entries)

	hash := sha256.New()
	_, _ = hash.Write([]byte(runnerIdentity))
	_, _ = hash.Write([]byte{0})
	for _, path := range entries {
		data, err := fs.ReadFile(dbmigrations.Files, path)
		if err != nil {
			return "", fmt.Errorf("read embedded migration %s: %w", path, err)
		}
		_, _ = hash.Write([]byte(path))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(data)
		_, _ = hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func MustHash() string {
	hash, err := Hash()
	if err != nil {
		panic(err)
	}
	return hash
}

func ShortHash(hash string) string {
	hash = strings.TrimSpace(hash)
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
