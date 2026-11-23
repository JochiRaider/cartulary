package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

type templateOperations struct {
	open       func(context.Context, string, postgres.Purpose) (*sql.DB, error)
	initialize func(context.Context, *sql.DB, string, *database_migrations.Source) error
	close      func(*sql.DB) error
	seal       func(context.Context, string, string) error
}

func defaultTemplateOperations() templateOperations {
	return templateOperations{OpenPurposeDatabase, InitializeFreshDatabase, (*sql.DB).Close, SealTemplateDatabase}
}

// Template request validation runs in this order: context, owner admission,
// then identity. Each failure is a caller error that acquires no attempt.
var (
	errTemplateContextRequired = errors.New("postgres template preparation requires a context")
	errTemplateOwnerRequired   = errors.New("postgres template preparation requires an admitted physical owner")
	errInvalidTemplateIdentity = errors.New("invalid postgres template identity")
)

type templateAttempt struct {
	name     string
	done     chan struct{}
	database *TestDatabase
	err      error
}

// PrepareTemplate builds the physical owner's immutable ordinary-fixture
// template. An empty name uses the canonical suite/schema/provisioning identity.
// Borrowers cannot prepare or replace the template advertised by their owner.
// A failed attempt is terminal; recovery requires a fresh physical owner.
func (owner *OwnedCluster) PrepareTemplate(ctx context.Context, name string) (*TestDatabase, error) {
	if ctx == nil {
		return nil, errTemplateContextRequired
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if owner == nil || owner.Container == nil || owner.Fixture() == nil {
		return nil, errTemplateOwnerRequired
	}
	h := owner.fixture
	if name == "" {
		name = templateDatabaseName(h.suiteHash, h.schemaHash)
	}
	if name != sanitizeIdentifier(name) || len(name) > 63 {
		return nil, errInvalidTemplateIdentity
	}
	owner.preparationMu.Lock()
	if attempt := owner.preparation; attempt != nil {
		owner.preparationMu.Unlock()
		if name != attempt.name {
			return nil, errors.New("postgres owner already has a different template attempt")
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-attempt.done:
			if attempt.err != nil {
				return nil, attempt.err
			}
			if err := h.admitPreparation(); err != nil {
				return nil, err
			}
			copy := *attempt.database
			return &copy, nil
		}
	}
	if err := h.admitPreparation(); err != nil {
		owner.preparationMu.Unlock()
		return nil, err
	}
	attempt := &templateAttempt{name: name, done: make(chan struct{})}
	owner.preparation = attempt
	owner.preparationMu.Unlock()
	preparation, cancel := context.WithTimeout(ctx, TemplatePreparationTimeout)
	defer cancel()
	database, err := owner.prepareTemplate(preparation, name)
	// The recipe returns only after independent handle and disposal settlement.
	if err != nil {
		h.quarantined.Store(true)
	}
	owner.preparationMu.Lock()
	attempt.err = err
	if database != nil {
		copy := *database
		attempt.database = &copy
	}
	close(attempt.done)
	owner.preparationMu.Unlock()
	return database, err
}

func (owner *OwnedCluster) prepareTemplate(ctx context.Context, name string) (_ *TestDatabase, retErr error) {
	h := owner.fixture
	ops := defaultTemplateOperations()
	if owner.templateOps != nil {
		ops = *owner.templateOps
	}
	attribution := fixtureAttribution{PostgresFixturePolicy: "suite_template"}
	if owner.seedTemplate != "" {
		// The seed build created and migrated this template through the fresh
		// path below. Adoption replaces both; sealing stays per generation.
		if err := h.adoptDatabase(ctx, name, owner.seedTemplate, suiteservices.FixtureReuseSuiteTemplate, attribution); err != nil {
			return nil, fmt.Errorf("adopt postgres seed template: %w", err)
		}
	} else if err := h.createDatabase(ctx, name, "", suiteservices.FixtureReuseSuiteTemplate, attribution); err != nil {
		return nil, fmt.Errorf("create postgres template: %w", err)
	}
	// Confirmed CREATE or adoption is already in both ownership ledgers before
	// any further step. Failure never publishes h.templateDB or a ready binding.
	defer func() {
		if retErr != nil {
			retErr = errors.Join(retErr, h.disposeDatabase(ctx, name, suiteservices.FixtureReuseSuiteTemplate, attribution))
		}
	}()
	if owner.seedTemplate == "" {
		if err := migrateTemplate(ctx, h, ops, name, attribution); err != nil {
			return nil, err
		}
	}
	if err := ops.seal(ctx, h.adminDSN, name); err != nil {
		return nil, fmt.Errorf("seal postgres template: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := h.admitPreparation(); err != nil {
		return nil, err
	}
	h.templateMu.Lock()
	h.templateDB = name
	h.templateMu.Unlock()
	return &TestDatabase{Name: name, DSN: h.dsnFor(name)}, nil
}

// prepareSeedTemplate creates and migrates a seed builder's template through
// the fresh path. It is not sealed: each adopting generation seals its own
// copy, so a failed publication can still drop it. The builder's clean
// shutdown checkpoint settles the migration before the image is committed.
func (owner *OwnedCluster) prepareSeedTemplate(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, TemplatePreparationTimeout)
	defer cancel()
	h := owner.fixture
	attribution := fixtureAttribution{PostgresFixturePolicy: "suite_template"}
	if err := h.createDatabase(ctx, seedTemplateName, "", suiteservices.FixtureReuseSuiteTemplate, attribution); err != nil {
		return fmt.Errorf("create postgres seed template: %w", err)
	}
	return migrateTemplate(ctx, h, defaultTemplateOperations(), seedTemplateName, attribution)
}

// migrateTemplate replays the canonical chain into a freshly created template
// and returns only after its migration handle is closed.
func migrateTemplate(ctx context.Context, h *Harness, ops templateOperations, name string, attribution fixtureAttribution) (retErr error) {
	dsn := h.dsnFor(name)
	db, err := ops.open(ctx, dsn, postgres.PurposeMigration)
	if err != nil {
		return fmt.Errorf("open postgres template: %w", err)
	}
	closed := false
	closeHandle := func() error {
		if closed {
			return nil
		}
		closed = true
		if err := ops.close(db); err != nil {
			h.quarantined.Store(true)
			return fmt.Errorf("close postgres template: %w", err)
		}
		return nil
	}
	defer func() { retErr = errors.Join(retErr, closeHandle()) }()
	source, err := dbmigrations.Source()
	if err != nil {
		return fmt.Errorf("load canonical migration source: %w", err)
	}
	duration, err := suiteservices.MeasureMigration(ctx, h.binding.environment(), func() error { return ops.initialize(ctx, db, dsn, source) })
	if err != nil {
		return fmt.Errorf("migrate postgres template: %w", err)
	}
	if err := h.recordEvent(suiteservices.Event{Type: suiteservices.EventPostgresDBMigrated, Name: name, Kind: "template",
		Details: h.postgresPreparationDetails(suiteservices.PostgresPreparationTemplate, name, suiteservices.FixtureReuseSuiteTemplate, attribution, duration)}); err != nil {
		return err
	}
	return closeHandle()
}
