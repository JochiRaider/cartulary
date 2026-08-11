// Package sourcecatalog owns the private, immutable representation of the
// Database Migrations source catalog. Repository boundary policy limits this
// bridge to the Database Migrations owner and the PostgreSQL test harness.
package sourcecatalog

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing/fstest"

	"github.com/pressly/goose/v3"
	gooselock "github.com/pressly/goose/v3/lock"
)

var migrationFilenamePattern = regexp.MustCompile(`^([0-9]{5})_[a-z0-9]+(?:_[a-z0-9]+)*\.sql$`)

const migrationAdvisoryLockID = int64(4097083626)

// Entry is a defensive metadata projection of one validated migration.
type Entry struct {
	Version      int64
	Filename     string
	SHA256       string
	HasGooseUp   bool
	HasGooseDown bool
}

// Catalog is an immutable validated snapshot. All storage and identity state
// remains private to this package.
type Catalog struct {
	files           fstest.MapFS
	entries         []catalogEntry
	lineageID       string
	lineageBoundary string
}

type catalogEntry struct {
	metadata Entry
	body     []byte
}

type catalogFS struct {
	files fstest.MapFS
}

func (catalog catalogFS) Open(name string) (fs.File, error) {
	return catalog.files.Open(name)
}

// Build copies and validates a complete migration catalog without database
// access. Lineage arguments are supplied only by the owning package.
func Build(fsys fs.FS, root string, lineageID string, lineageBoundary string) (*Catalog, error) {
	if fsys == nil {
		return nil, errors.New("migration source filesystem is nil")
	}
	if root == "" || !fs.ValidPath(root) || path.Clean(root) != root {
		return nil, errors.New("migration source root is invalid")
	}
	lineageID = strings.TrimSpace(lineageID)
	lineageBoundary = strings.TrimSpace(lineageBoundary)
	if lineageID == "" || lineageBoundary == "" {
		return nil, errors.New("migration source lineage metadata is required")
	}

	directoryEntries, err := fs.ReadDir(fsys, root)
	if err != nil {
		return nil, fmt.Errorf("read migration source root: %w", err)
	}
	if len(directoryEntries) == 0 {
		return nil, errors.New("migration source catalog is empty")
	}

	files := make(fstest.MapFS, len(directoryEntries))
	entries := make([]catalogEntry, 0, len(directoryEntries))
	seenVersions := make(map[int64]string, len(directoryEntries))
	for _, entry := range directoryEntries {
		if entry.IsDir() || entry.Type()&fs.ModeSymlink != 0 {
			return nil, fmt.Errorf("migration source contains unexpected entry %q", entry.Name())
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return nil, fmt.Errorf("inspect migration source entry %q: %w", entry.Name(), infoErr)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("migration source contains unexpected entry %q", entry.Name())
		}
		match := migrationFilenamePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			return nil, fmt.Errorf("migration source filename %q is invalid", entry.Name())
		}
		version, parseErr := strconv.ParseInt(match[1], 10, 64)
		if parseErr != nil || version <= 0 {
			return nil, fmt.Errorf("migration source version in %q is invalid", entry.Name())
		}
		if previous, exists := seenVersions[version]; exists {
			return nil, fmt.Errorf("migration source version %05d is duplicated by %q and %q", version, previous, entry.Name())
		}

		name := path.Join(root, entry.Name())
		body, readErr := fs.ReadFile(fsys, name)
		if readErr != nil {
			return nil, fmt.Errorf("read migration source file %q: %w", entry.Name(), readErr)
		}
		if markerErr := validateMigrationMarkers(entry.Name(), body); markerErr != nil {
			return nil, markerErr
		}

		copied := append([]byte(nil), body...)
		digest := sha256.Sum256(copied)
		metadata := Entry{
			Version:      version,
			Filename:     entry.Name(),
			SHA256:       hex.EncodeToString(digest[:]),
			HasGooseUp:   true,
			HasGooseDown: true,
		}
		files[entry.Name()] = &fstest.MapFile{Data: copied, Mode: 0o444}
		entries = append(entries, catalogEntry{metadata: metadata, body: copied})
		seenVersions[version] = entry.Name()
	}

	sort.Slice(entries, func(left, right int) bool {
		if entries[left].metadata.Version == entries[right].metadata.Version {
			return entries[left].metadata.Filename < entries[right].metadata.Filename
		}
		return entries[left].metadata.Version < entries[right].metadata.Version
	})
	for index, entry := range entries {
		want := int64(index + 1)
		if entry.metadata.Version != want {
			return nil, fmt.Errorf("migration source versions are not contiguous: got %05d want %05d", entry.metadata.Version, want)
		}
	}

	return &Catalog{
		files:           files,
		entries:         entries,
		lineageID:       lineageID,
		lineageBoundary: lineageBoundary,
	}, nil
}

func validateMigrationMarkers(name string, body []byte) error {
	upCount := 0
	downCount := 0
	upSeen := false
	downSeen := false
	statementDepth := 0
	for _, rawLine := range strings.Split(string(body), "\n") {
		line := strings.TrimSpace(rawLine)
		if !strings.HasPrefix(line, "-- +goose") {
			continue
		}
		switch line {
		case "-- +goose Up":
			upCount++
			upSeen = true
			if downSeen || statementDepth != 0 {
				return fmt.Errorf("migration source file %q has an invalid Up marker", name)
			}
		case "-- +goose Down":
			downCount++
			downSeen = true
			if !upSeen || statementDepth != 0 {
				return fmt.Errorf("migration source file %q has an invalid Down marker", name)
			}
		case "-- +goose NO TRANSACTION":
			if statementDepth != 0 {
				return fmt.Errorf("migration source file %q has a misplaced NO TRANSACTION directive", name)
			}
		case "-- +goose StatementBegin":
			if !upSeen || statementDepth != 0 {
				return fmt.Errorf("migration source file %q has an unbalanced StatementBegin directive", name)
			}
			statementDepth = 1
		case "-- +goose StatementEnd":
			if statementDepth != 1 {
				return fmt.Errorf("migration source file %q has an unbalanced StatementEnd directive", name)
			}
			statementDepth = 0
		default:
			return fmt.Errorf("migration source file %q has unsupported directive %q", name, line)
		}
	}
	if upCount != 1 || downCount != 1 {
		return fmt.Errorf("migration source file %q must contain exactly one Up and one Down marker", name)
	}
	if statementDepth != 0 {
		return fmt.Errorf("migration source file %q has an unbalanced statement block", name)
	}
	return nil
}

func Validate(catalog *Catalog) error {
	if catalog == nil || catalog.files == nil || len(catalog.entries) == 0 || catalog.lineageID == "" || catalog.lineageBoundary == "" {
		return errors.New("migration source is invalid")
	}
	return nil
}

func Inspect(catalog *Catalog) ([]Entry, error) {
	if err := Validate(catalog); err != nil {
		return nil, err
	}
	entries := make([]Entry, len(catalog.entries))
	for index, entry := range catalog.entries {
		entries[index] = entry.metadata
	}
	return entries, nil
}

func HeadVersion(catalog *Catalog) int64 {
	if catalog == nil || len(catalog.entries) == 0 {
		return 0
	}
	return catalog.entries[len(catalog.entries)-1].metadata.Version
}

func LineageID(catalog *Catalog) string {
	if catalog == nil {
		return ""
	}
	return catalog.lineageID
}

func LineageBoundary(catalog *Catalog) string {
	if catalog == nil {
		return ""
	}
	return catalog.lineageBoundary
}

// NewProvider constructs an invocation-local PostgreSQL Goose provider over
// the immutable catalog with the owner lock policy and global registration
// disabled.
func NewProvider(db *sql.DB, catalog *Catalog, locker gooselock.SessionLocker) (*goose.Provider, error) {
	if db == nil {
		return nil, errors.New("migration provider database is nil")
	}
	if err := Validate(catalog); err != nil {
		return nil, err
	}
	if locker == nil {
		return nil, errors.New("migration provider session locker is nil")
	}
	return goose.NewProvider(
		goose.DialectPostgres,
		db,
		catalogFS{files: catalog.files},
		goose.WithDisableGlobalRegistry(true),
		goose.WithLogger(log.New(io.Discard, "", 0)),
		goose.WithSessionLocker(locker),
		goose.WithTableName("public.goose_db_version"),
	)
}

func NewSessionLocker() (gooselock.SessionLocker, error) {
	return gooselock.NewPostgresSessionLocker(
		gooselock.WithLockID(migrationAdvisoryLockID),
		gooselock.WithLockTimeout(1, 300),
		gooselock.WithUnlockTimeout(1, 30),
	)
}

func SchemaHash(catalog *Catalog, runnerIdentity string) (string, error) {
	if err := Validate(catalog); err != nil {
		return "", err
	}
	if strings.TrimSpace(runnerIdentity) == "" {
		return "", errors.New("migration schema hash runner identity is required")
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte(runnerIdentity))
	_, _ = hash.Write([]byte{0})
	for _, entry := range catalog.entries {
		_, _ = hash.Write([]byte(entry.metadata.Filename))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(entry.body)
		_, _ = hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
