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

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

type PreparedFixture struct {
	DatabaseName      string
	DSN               string
	Bucket            string
	S3Endpoint        string
	S3AccessKey       string
	S3SecretKey       string
	S3Secure          bool
	FixtureProfileID  string
	SnapshotKey       string
	BuilderUnitID     string
	RowID             string
	PredicateID       string
	CloneLeaseID      string
	CloneOrdinal      int
	RuntimeBundlePath string
	RuntimeBundleRoot string
}

type LeaseMetadata struct {
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
	return prepareWithProfile(ctx, env, profile)
}

func prepareWithProfile(ctx context.Context, env map[string]string, profile performancefixtureprofile.Profile) (PreparedFixture, error) {
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
	dsn, err := replaceDatabaseInDSN(adminDSN, cloneName)
	if err != nil {
		_ = dropDatabase(context.Background(), adminDSN, cloneName)
		return PreparedFixture{}, err
	}
	s3Harness, err := s3test.StartShared(ctx)
	if err != nil {
		_ = dropDatabase(context.Background(), adminDSN, cloneName)
		return PreparedFixture{}, fmt.Errorf("attach suite object-store: %w", err)
	}
	bucket, err := s3Harness.BootstrapBucket(ctx, "web-e2e")
	if err != nil {
		_ = dropDatabase(context.Background(), adminDSN, cloneName)
		return PreparedFixture{}, err
	}
	templateRuntimeRoot, err := templateRuntimeRoot(env, key)
	if err != nil {
		_ = dropDatabase(context.Background(), adminDSN, cloneName)
		_ = s3Harness.CleanupBucket(context.Background(), bucket)
		return PreparedFixture{}, err
	}
	destinationRoot, err := cloneRuntimeRoot(env, leaseID)
	if err != nil {
		_ = dropDatabase(context.Background(), adminDSN, cloneName)
		_ = s3Harness.CleanupBucket(context.Background(), bucket)
		return PreparedFixture{}, err
	}
	if err := os.MkdirAll(filepath.Dir(destinationRoot), 0o700); err != nil {
		_ = dropDatabase(context.Background(), adminDSN, cloneName)
		_ = s3Harness.CleanupBucket(context.Background(), bucket)
		return PreparedFixture{}, err
	}
	bundlePath, err := fixture.CopyRuntimeBundle(profile, filepath.Join(templateRuntimeRoot, fixture.RuntimeBundleName), destinationRoot)
	if err != nil {
		_ = dropDatabase(context.Background(), adminDSN, cloneName)
		_ = s3Harness.CleanupBucket(context.Background(), bucket)
		return PreparedFixture{}, err
	}
	return PreparedFixture{
		DatabaseName:      cloneName,
		DSN:               dsn,
		Bucket:            bucket,
		S3Endpoint:        s3Harness.Endpoint,
		S3AccessKey:       s3Harness.AccessKey,
		S3SecretKey:       s3Harness.SecretKey,
		S3Secure:          s3Harness.Secure,
		FixtureProfileID:  profile.FixtureProfileID,
		SnapshotKey:       key,
		BuilderUnitID:     builderID,
		RowID:             rowID,
		PredicateID:       predicateID,
		CloneLeaseID:      leaseID,
		CloneOrdinal:      ordinal,
		RuntimeBundlePath: bundlePath,
		RuntimeBundleRoot: destinationRoot,
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
	if sessionErr != nil || credentialErr != nil || databaseErr != nil || bucketErr != nil || !processClean {
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
		return errors.Join(sessionErr, credentialErr, databaseErr, bucketErr, writeLeaseArtifact(env, profile, artifact))
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
		_ = db.Close()
		return fmt.Errorf("revoke performance fixture sessions: %w", err)
	}
	return db.Close()
}
