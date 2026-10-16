package performancefixturelifecycle

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

type PreparedFixture struct {
	ObjectSnapshot        *ObjectSnapshotSummary
	DatabaseName          string
	DSN                   string
	Bucket                string
	S3Endpoint            string
	S3AccessKey           string
	S3SecretKey           string
	S3Secure              bool
	S3RootCertificatePath string
	FixtureProfileID      string
	SnapshotKey           string
	BuilderUnitID         string
	RowID                 string
	PredicateID           string
	CloneLeaseID          string
	CloneOrdinal          int
	RuntimeBundlePath     string
	RuntimeBundleRoot     string
}

type LeaseMetadata struct {
	ObjectSnapshot    *ObjectSnapshotSummary
	DatabaseName      string
	Bucket            string
	FixtureProfileID  string
	SnapshotKey       string
	BuilderUnitID     string
	RowID             string
	PredicateID       string
	CloneLeaseID      string
	CloneOrdinal      int
	RuntimeBundleRoot string
}

type preparationFailure struct {
	cause   error
	cleanup []CleanupResult
}

func (failure *preparationFailure) Error() string { return failure.cause.Error() }
func (failure *preparationFailure) Unwrap() error { return failure.cause }

type CleanupPorts struct {
	CleanupSessions func(context.Context, map[string]string, string) error
	DetectLeaks     func(context.Context, LeaseMetadata, map[string]string) error
	CleanupDatabase func(context.Context, LeaseMetadata, map[string]string) error
	CleanupBucket   func(context.Context, LeaseMetadata, map[string]string) error
}

func Prepare(ctx context.Context, env map[string]string) (PreparedFixture, error) {
	profileID := strings.TrimSpace(env["CARTULARY_FIXTURE_PROFILE_ID"])
	profile, err := ResolveProfile(profileID)
	if err != nil {
		return PreparedFixture{}, err
	}
	build, err := readBuildArtifact(env, profile)
	if err != nil {
		return PreparedFixture{}, err
	}
	return prepareWithProfile(ctx, env, profile, build)
}

func prepareWithProfile(ctx context.Context, env map[string]string, profile performancefixtureprofile.Profile, build BuildArtifact) (_ PreparedFixture, retErr error) {
	if err := validatePreparedBuild(env, profile, build); err != nil {
		return PreparedFixture{}, err
	}

	key := strings.TrimSpace(env["CARTULARY_FIXTURE_SNAPSHOT_KEY"])
	builderID := strings.TrimSpace(env["CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID"])
	rowID := strings.TrimSpace(env["CARTULARY_FIXTURE_ROW_ID"])
	predicateID := strings.TrimSpace(env["CARTULARY_FIXTURE_PREDICATE_ID"])
	leaseID := strings.TrimSpace(env["CARTULARY_FIXTURE_CLONE_LEASE_ID"])
	ordinal, err := strconv.Atoi(strings.TrimSpace(env["CARTULARY_FIXTURE_CLONE_ORDINAL"]))
	if !lowerHexDigest(key) || builderID != "fixture_snapshot:default:"+profile.FixtureProfileID+":"+key ||
		err != nil || ordinal < 1 || !safeCatalogIdentity(rowID) || !predicateAdmitted(profile, predicateID) || leaseID == "" {
		return PreparedFixture{}, errors.New("performance browser clone requires complete admitted profile identity")
	}
	adminDSN := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGAdminDSNEnv))
	suiteID := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.SuiteIDEnv))
	templateName := templateName(suiteID, key)
	if err := validateTemplate(ctx, adminDSN, templateName, templateOwner(suiteID, key)); err != nil {
		return PreparedFixture{}, err
	}
	cloneName, err := newCloneName()
	if err != nil {
		return PreparedFixture{}, err
	}
	if err := cloneDatabase(ctx, adminDSN, templateName, cloneName); err != nil {
		return PreparedFixture{}, err
	}
	// All allocations remain owned here until complete restoration admits the
	// consumer. Cancellation cannot suppress cleanup of a partial destination.
	var s3Harness *s3test.Harness
	var bucket, destinationRoot string
	defer func() {
		if retErr == nil {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		bucketOutcome, credentialOutcome := "not_acquired", "not_acquired"
		if destinationRoot != "" {
			err := fixture.RemoveRuntimeBundle(destinationRoot)
			credentialOutcome = cleanupOutcome(err)
			retErr = errors.Join(retErr, err)
		}
		if bucket != "" {
			err := s3Harness.CleanupBucket(cleanupCtx, bucket)
			bucketOutcome = cleanupOutcome(err)
			retErr = errors.Join(retErr, err)
		}
		databaseErr := dropDatabase(cleanupCtx, adminDSN, cloneName)
		retErr = &preparationFailure{cause: errors.Join(retErr, databaseErr), cleanup: cleanupResults(bucketOutcome, credentialOutcome, cleanupOutcome(databaseErr), "not_acquired", "not_acquired")}
	}()
	dsn, err := replaceDatabaseInDSN(adminDSN, cloneName)
	if err != nil {
		return PreparedFixture{}, err
	}
	s3Harness, err = s3test.StartShared(ctx)
	if err != nil {
		return PreparedFixture{}, fmt.Errorf("attach suite object-store: %w", err)
	}
	bucket, err = s3Harness.BootstrapBucket(ctx, "web-e2e")
	if err != nil {
		return PreparedFixture{}, err
	}
	templateRoot, err := templateRuntimeRoot(env, key)
	if err != nil {
		return PreparedFixture{}, err
	}
	objects, err := objectstore.Setup(ctx, objectstore.Settings{
		BindingKind: "managed_service", Endpoint: s3Harness.Endpoint,
		AccessKey: s3Harness.AccessKey, SecretKey: s3Harness.SecretKey,
		Secure: s3Harness.Secure, Bucket: bucket, RootCertificatePath: s3Harness.RootCertificatePath(),
	}, objectstore.Instrumentation{})
	if err != nil {
		return PreparedFixture{}, err
	}
	restoreErr := restoreObjectSnapshot(ctx, templateRoot, key, *build.ObjectSnapshot, objects)
	closeErr := objects.Close()
	if err := errors.Join(restoreErr, closeErr); err != nil {
		return PreparedFixture{}, err
	}
	candidateRoot, err := cloneRuntimeRoot(env, leaseID)
	if err != nil {
		return PreparedFixture{}, err
	}
	if err := os.MkdirAll(filepath.Dir(candidateRoot), 0o700); err != nil {
		return PreparedFixture{}, err
	}
	bundlePath, err := fixture.CopyRuntimeBundle(profile, filepath.Join(templateRoot, fixture.RuntimeBundleName), candidateRoot)
	if err != nil {
		return PreparedFixture{}, err
	}
	destinationRoot = candidateRoot
	proof := *build.ObjectSnapshot
	return PreparedFixture{
		DatabaseName:          cloneName,
		ObjectSnapshot:        &proof,
		DSN:                   dsn,
		Bucket:                bucket,
		S3Endpoint:            s3Harness.Endpoint,
		S3AccessKey:           s3Harness.AccessKey,
		S3SecretKey:           s3Harness.SecretKey,
		S3Secure:              s3Harness.Secure,
		S3RootCertificatePath: s3Harness.RootCertificatePath(),
		FixtureProfileID:      profile.FixtureProfileID,
		SnapshotKey:           key,
		BuilderUnitID:         builderID,
		RowID:                 rowID,
		PredicateID:           predicateID,
		CloneLeaseID:          leaseID,
		CloneOrdinal:          ordinal,
		RuntimeBundlePath:     bundlePath,
		RuntimeBundleRoot:     destinationRoot,
	}, nil
}

func CleanupLease(ctx context.Context, env map[string]string, metadata LeaseMetadata, ports CleanupPorts) error {
	profile, err := ResolveProfile(metadata.FixtureProfileID)
	if err != nil {
		return err
	}
	return cleanupLeaseWithProfile(ctx, env, profile, metadata, ports)
}

func cleanupLeaseWithProfile(ctx context.Context, env map[string]string, profile performancefixtureprofile.Profile, metadata LeaseMetadata, ports CleanupPorts) error {
	if ports.CleanupSessions == nil || ports.DetectLeaks == nil || ports.CleanupDatabase == nil || ports.CleanupBucket == nil {
		return errors.New("performance fixture cleanup ports are incomplete")
	}
	leaseIdentity, err := opaqueLeaseIdentity(metadata.CloneLeaseID)
	if err != nil {
		return err
	}
	processClean := env["CARTULARY_FIXTURE_PROCESS_CLEANUP_COMPLETE"] == "1"
	artifact := leaseArtifact{
		SchemaID:          profile.ArtifactPolicy.LeaseSchemaID,
		ObjectSnapshot:    metadata.ObjectSnapshot,
		FixtureProfileID:  profile.FixtureProfileID,
		SnapshotKey:       metadata.SnapshotKey,
		BuilderUnitID:     metadata.BuilderUnitID,
		RowID:             metadata.RowID,
		PredicateID:       metadata.PredicateID,
		LeaseIdentity:     leaseIdentity,
		CloneOrdinal:      metadata.CloneOrdinal,
		CreationState:     "created",
		IsolationResult:   "isolated",
		RedactionPolicyID: profile.RedactionPolicy.PolicyID,
		FinalizedAt:       time.Now().UTC().Format(time.RFC3339Nano),
	}
	sessionErr := ports.CleanupSessions(ctx, env, metadata.DatabaseName)
	credentialErr := fixture.RemoveRuntimeBundle(metadata.RuntimeBundleRoot)
	leakErr := ports.DetectLeaks(ctx, metadata, env)
	var databaseErr error
	if leakErr == nil {
		databaseErr = ports.CleanupDatabase(ctx, metadata, env)
	} else {
		databaseErr = leakErr
	}
	bucketErr := ports.CleanupBucket(ctx, metadata, env)
	artifact.CleanupResults = cleanupResults(
		cleanupOutcome(bucketErr),
		cleanupOutcome(credentialErr),
		cleanupOutcome(databaseErr),
		boolCleanupOutcome(processClean),
		cleanupOutcome(sessionErr),
	)
	var processErr error
	if !processClean {
		processErr = errors.New("performance fixture process cleanup is incomplete")
	}
	if sessionErr != nil || credentialErr != nil || databaseErr != nil || bucketErr != nil || processErr != nil {
		artifact.CleanupState = "failed"
		switch {
		case sessionErr != nil:
			artifact.FailureCode = "session_cleanup_failed"
		case credentialErr != nil:
			artifact.FailureCode = "credential_cleanup_failed"
		case databaseErr != nil:
			artifact.FailureCode = "database_cleanup_failed"
		case bucketErr != nil:
			artifact.FailureCode = "bucket_cleanup_failed"
		default:
			artifact.FailureCode = "process_cleanup_failed"
		}
		return errors.Join(sessionErr, credentialErr, databaseErr, bucketErr, processErr, writeLeaseArtifact(env, profile, artifact))
	}
	artifact.CleanupState = "complete"
	return writeLeaseArtifact(env, profile, artifact)
}

func cleanupOutcome(err error) string {
	if err != nil {
		return "failed"
	}
	return "complete"
}

func boolCleanupOutcome(complete bool) string {
	if complete {
		return "complete"
	}
	return "failed"
}

func RevokeSessions(ctx context.Context, env map[string]string, databaseName string) error {
	adminDSN := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGAdminDSNEnv))
	dsn, err := replaceDatabaseInDSN(adminDSN, databaseName)
	if err != nil {
		return err
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `UPDATE user_sessions SET revoked_at = COALESCE(revoked_at, now())`); err != nil {
		closeErr := db.Close()
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "3D000" {
			// A previous owned cleanup already dropped this exact clone.
			return closeErr
		}
		return errors.Join(fmt.Errorf("revoke performance fixture sessions: %w", err), closeErr)
	}
	return db.Close()
}
