package database_migrations

import (
	"context"
	"database/sql"
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

var errNilMigrateContext = errors.New("postgres migrate: nil context")

var migrationFilenamePattern = regexp.MustCompile(`^([0-9]{5})_[a-z0-9]+(?:_[a-z0-9]+)*\.sql$`)

// Source is an immutable validated snapshot of one migration catalog and its
// lineage identity. Its storage and metadata are intentionally private.
type Source struct {
	catalog         fstest.MapFS
	versions        []int64
	lineageID       string
	lineageBoundary string
}

// NewSource copies and validates a migration catalog without database access.
func NewSource(fsys fs.FS, root string, lineageID string, lineageBoundary string) (Source, error) {
	if fsys == nil {
		return Source{}, errors.New("migration source filesystem is nil")
	}
	if root == "" || !fs.ValidPath(root) || path.Clean(root) != root {
		return Source{}, errors.New("migration source root is invalid")
	}
	lineageID = strings.TrimSpace(lineageID)
	lineageBoundary = strings.TrimSpace(lineageBoundary)
	if lineageID == "" || lineageBoundary == "" {
		return Source{}, errors.New("migration source lineage metadata is required")
	}

	entries, err := fs.ReadDir(fsys, root)
	if err != nil {
		return Source{}, fmt.Errorf("read migration source root: %w", err)
	}
	if len(entries) == 0 {
		return Source{}, errors.New("migration source catalog is empty")
	}

	catalog := make(fstest.MapFS, len(entries))
	versions := make([]int64, 0, len(entries))
	seenVersions := make(map[int64]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || entry.Type()&fs.ModeSymlink != 0 {
			return Source{}, fmt.Errorf("migration source contains unexpected entry %q", entry.Name())
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return Source{}, fmt.Errorf("inspect migration source entry %q: %w", entry.Name(), infoErr)
		}
		if !info.Mode().IsRegular() {
			return Source{}, fmt.Errorf("migration source contains unexpected entry %q", entry.Name())
		}
		match := migrationFilenamePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			return Source{}, fmt.Errorf("migration source filename %q is invalid", entry.Name())
		}
		version, parseErr := strconv.ParseInt(match[1], 10, 64)
		if parseErr != nil || version <= 0 {
			return Source{}, fmt.Errorf("migration source version in %q is invalid", entry.Name())
		}
		if previous, exists := seenVersions[version]; exists {
			return Source{}, fmt.Errorf("migration source version %05d is duplicated by %q and %q", version, previous, entry.Name())
		}

		name := path.Join(root, entry.Name())
		body, readErr := fs.ReadFile(fsys, name)
		if readErr != nil {
			return Source{}, fmt.Errorf("read migration source file %q: %w", entry.Name(), readErr)
		}
		if markerErr := validateMigrationMarkers(entry.Name(), body); markerErr != nil {
			return Source{}, markerErr
		}

		copied := append([]byte(nil), body...)
		catalog[entry.Name()] = &fstest.MapFile{Data: copied, Mode: 0o444}
		versions = append(versions, version)
		seenVersions[version] = entry.Name()
	}

	sort.Slice(versions, func(left, right int) bool { return versions[left] < versions[right] })
	for index, version := range versions {
		want := int64(index + 1)
		if version != want {
			return Source{}, fmt.Errorf("migration source versions are not contiguous: got %05d want %05d", version, want)
		}
	}

	return Source{
		catalog:         catalog,
		versions:        append([]int64(nil), versions...),
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

func (source Source) validate() error {
	if source.catalog == nil || len(source.versions) == 0 || source.lineageID == "" || source.lineageBoundary == "" {
		return errors.New("migration source is invalid")
	}
	return nil
}

func (source Source) headVersion() int64 {
	if len(source.versions) == 0 {
		return 0
	}
	return source.versions[len(source.versions)-1]
}

func (source Source) hasVersion(version int64) bool {
	index := version - 1
	return index >= 0 && index < int64(len(source.versions)) && source.versions[index] == version
}

// Apply advances the database to the validated source head.
func Apply(ctx context.Context, db *sql.DB, source Source) (retErr error) {
	if ctx == nil {
		return newMigrationFailure(reasonMigrationContextInvalid, errNilMigrateContext)
	}
	if err := ctx.Err(); err != nil {
		return newMigrationFailure(reasonSchemaMigrationExecutionFailed, err)
	}
	if err := source.validate(); err != nil {
		return newMigrationFailure(reasonMigrationSourceInvalid, err)
	}
	if db == nil {
		return newMigrationFailure(reasonMigrationDatabaseUnavailable, nil)
	}

	locker, err := newMigrationSessionLocker()
	if err != nil {
		return newMigrationFailure(reasonMigrationLockAcquisitionFailed, err)
	}
	workNeeded := false
	if err := withLockedMigrationSession(ctx, db, locker, func(ctx context.Context, conn *sql.Conn) error {
		var classifyErr error
		workNeeded, classifyErr = migrationWorkNeeded(ctx, conn, source)
		return classifyErr
	}); err != nil {
		return err
	}

	if workNeeded {
		providerLocker := &validatingSessionLocker{
			delegate: locker,
			validate: func(ctx context.Context, conn *sql.Conn) error {
				_, classifyErr := migrationWorkNeeded(ctx, conn, source)
				return classifyErr
			},
			lockTimeout:   migrationLockTimeout,
			unlockTimeout: migrationUnlockTimeout,
		}
		if err := runGooseProvider(ctx, db, source, providerLocker); err != nil {
			return normalizeProviderFailure(err)
		}
	}

	if err := withLockedMigrationSession(ctx, db, locker, func(ctx context.Context, conn *sql.Conn) error {
		return verifyMigrationPostcondition(ctx, conn, source)
	}); err != nil {
		var failure MigrationFailure
		if errors.As(err, &failure) && failure.ReasonCode() == reasonMigrationLockAcquisitionFailed {
			return err
		}
		return normalizeMigrationFailure(err, reasonSchemaMigrationPostcondition)
	}
	return nil
}

func runGooseProvider(ctx context.Context, db *sql.DB, source Source, locker gooselock.SessionLocker) (retErr error) {
	defer recoverProviderPanic(&retErr)
	provider, err := newGooseProvider(db, source.catalog, log.New(io.Discard, "", 0), locker)
	if err != nil {
		return err
	}
	_, err = provider.Up(ctx)
	if err != nil && ctx.Err() != nil {
		return newMigrationFailure(reasonSchemaMigrationExecutionFailed, ctx.Err())
	}
	return err
}

func newGooseProvider(db *sql.DB, sourceFS fs.FS, logger goose.Logger, sessionLockers ...gooselock.SessionLocker) (*goose.Provider, error) {
	options := []goose.ProviderOption{
		goose.WithDisableGlobalRegistry(true),
		goose.WithLogger(logger),
	}
	if len(sessionLockers) > 0 && sessionLockers[0] != nil {
		options = append(options, goose.WithSessionLocker(sessionLockers[0]))
	}
	return goose.NewProvider(
		goose.DialectPostgres,
		db,
		sourceFS,
		options...,
	)
}
