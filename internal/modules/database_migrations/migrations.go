package database_migrations

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"io/fs"
	"log"

	gooselock "github.com/pressly/goose/v3/lock"

	"github.com/JochiRaider/cartulary/internal/modules/database_migrations/sourcecatalog"
)

var errNilMigrateContext = errors.New("postgres migrate: nil context")

const (
	canonicalLineageID       = "cartulary.prod_ddl_rebaseline.v1"
	canonicalLineageBoundary = "prod_ddl_rebaseline_v1"
)

// Source is the owner-private immutable migration catalog representation. Its
// fields remain private in sourcecatalog; boundary policy restricts bridge use.
type Source = sourcecatalog.Catalog

// SourceInspection is a defensive metadata-only projection of a validated
// migration source. It contains no SQL bytes or source locators.
type SourceInspection struct {
	Entries      []SourceInspectionEntry
	MinVersion   int64
	MaxVersion   int64
	VersionCount int
}

type SourceInspectionEntry struct {
	Version      int64
	Filename     string
	SHA256       string
	HasGooseUp   bool
	HasGooseDown bool
}

// BuildCanonicalEmbedded copies and validates the canonical embedded catalog.
// Lineage identity is fixed inside the owner and cannot be selected by callers.
func BuildCanonicalEmbedded(fsys fs.FS, root string) (*Source, error) {
	return buildSource(fsys, root, canonicalLineageID, canonicalLineageBoundary)
}

func buildSource(fsys fs.FS, root string, lineageID string, lineageBoundary string) (*Source, error) {
	catalog, err := sourcecatalog.Build(fsys, root, lineageID, lineageBoundary)
	if err != nil {
		return nil, err
	}
	return catalog, nil
}

func InspectSource(source *Source) (SourceInspection, error) {
	if err := validateSource(source); err != nil {
		return SourceInspection{}, err
	}
	entries, err := sourcecatalog.Inspect(source)
	if err != nil {
		return SourceInspection{}, err
	}
	inspection := SourceInspection{Entries: make([]SourceInspectionEntry, len(entries)), VersionCount: len(entries)}
	for index, entry := range entries {
		inspection.Entries[index] = SourceInspectionEntry{
			Version:      entry.Version,
			Filename:     entry.Filename,
			SHA256:       entry.SHA256,
			HasGooseUp:   entry.HasGooseUp,
			HasGooseDown: entry.HasGooseDown,
		}
	}
	if len(entries) > 0 {
		inspection.MinVersion = entries[0].Version
		inspection.MaxVersion = entries[len(entries)-1].Version
	}
	return inspection, nil
}

func SchemaHash(source *Source, runnerIdentity string) (string, error) {
	if err := validateSource(source); err != nil {
		return "", err
	}
	return sourcecatalog.SchemaHash(source, runnerIdentity)
}

func validateSource(source *Source) error {
	if source == nil {
		return errors.New("migration source is invalid")
	}
	return sourcecatalog.Validate(source)
}

func sourceHeadVersion(source *Source) int64 {
	if source == nil {
		return 0
	}
	return sourcecatalog.HeadVersion(source)
}

func sourceHasVersion(source *Source, version int64) bool {
	return source != nil && sourcecatalog.HasVersion(source, version)
}

func sourceLineageID(source *Source) string {
	if source == nil {
		return ""
	}
	return sourcecatalog.LineageID(source)
}

func sourceLineageBoundary(source *Source) string {
	if source == nil {
		return ""
	}
	return sourcecatalog.LineageBoundary(source)
}

// Apply advances the database to the validated source head.
func Apply(ctx context.Context, db *sql.DB, source *Source) (retErr error) {
	if ctx == nil {
		return newMigrationFailure(reasonMigrationContextInvalid, errNilMigrateContext)
	}
	if err := ctx.Err(); err != nil {
		return newMigrationFailure(reasonSchemaMigrationExecutionFailed, err)
	}
	if err := validateSource(source); err != nil {
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

func runGooseProvider(ctx context.Context, db *sql.DB, source *Source, locker gooselock.SessionLocker) (retErr error) {
	defer recoverProviderPanic(&retErr)
	provider, err := sourcecatalog.NewProvider(db, source, log.New(io.Discard, "", 0), locker)
	if err != nil {
		return err
	}
	_, err = provider.Up(ctx)
	if err != nil && ctx.Err() != nil {
		return newMigrationFailure(reasonSchemaMigrationExecutionFailed, ctx.Err())
	}
	return err
}
