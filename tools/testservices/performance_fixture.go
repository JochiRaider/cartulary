package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
	"github.com/JochiRaider/cartulary/internal/testutil/performancefixtureassembly"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

const (
	performanceFixtureVersion             = "cartulary.perf.large_grid.v1"
	performanceFixtureSeed                = 20260405
	performanceFixtureSnapshotSchemaID    = "cartulary.performance_fixture_snapshot.v1"
	performanceFixtureLeaseSchemaID       = "cartulary.performance_fixture_snapshot_lease.v1"
	performanceFixtureRedactionPolicyID   = "cartulary.performance_fixture_redaction.v1"
	performanceFixtureRuntimeStaleAge     = 24 * time.Hour
	performanceFixtureRuntimeJanitorLimit = 32
)

type performanceFixtureBuildArgs struct {
	FixtureProfileID     string
	SnapshotKey          string
	MigrationDigest      string
	SourceContractDigest string
	BuilderUnitID        string
	ArtifactFile         string
}

type performanceFixtureBuildArtifact struct {
	SchemaID                 string                       `json:"schema_id"`
	FixtureProfileID         string                       `json:"fixture_profile_id"`
	FixtureVersion           string                       `json:"fixture_version"`
	Seed                     int                          `json:"seed"`
	SnapshotKeySchemaID      string                       `json:"snapshot_key_schema_id"`
	SnapshotKey              string                       `json:"snapshot_key"`
	MigrationDigest          string                       `json:"migration_digest"`
	SourceContractDigest     string                       `json:"source_contract_digest"`
	BuilderUnitID            string                       `json:"builder_unit_id"`
	BuildOrdinal             int                          `json:"build_ordinal"`
	State                    string                       `json:"state"`
	ContributionReceipts     []fixture.Receipt            `json:"contribution_receipts"`
	SemanticValidationDigest string                       `json:"semantic_validation_digest"`
	Validation               performanceFixtureValidation `json:"validation"`
	FailureCode              string                       `json:"failure_code,omitempty"`
	RedactionPolicyID        string                       `json:"redaction_policy_id"`
	CreatedAt                string                       `json:"created_at"`
}

type performanceFixtureValidation struct {
	ExactCounts              bool `json:"exact_counts"`
	RelationshipDistribution bool `json:"relationship_distribution"`
	DefaultView              bool `json:"default_view"`
	Authorization            bool `json:"authorization"`
	ProjectionReadiness      bool `json:"projection_readiness"`
	NoActiveSessions         bool `json:"no_active_sessions"`
	ConnectionsClosed        bool `json:"connections_closed"`
}

type performanceFixtureLeaseArtifact struct {
	SchemaID              string `json:"schema_id"`
	FixtureProfileID      string `json:"fixture_profile_id"`
	SnapshotKey           string `json:"snapshot_key"`
	BuilderUnitID         string `json:"builder_unit_id"`
	RowID                 string `json:"row_id"`
	PredicateID           string `json:"predicate_id"`
	LeaseID               string `json:"lease_id"`
	CloneOrdinal          int    `json:"clone_ordinal"`
	CreationState         string `json:"creation_state"`
	IsolationResult       string `json:"isolation_result"`
	CredentialCopyCleanup bool   `json:"credential_copy_cleanup"`
	DatabaseCleanup       bool   `json:"database_cleanup"`
	BucketCleanup         bool   `json:"bucket_cleanup"`
	SessionCleanup        bool   `json:"session_cleanup"`
	ProcessCleanup        bool   `json:"process_cleanup"`
	CleanupState          string `json:"cleanup_state"`
	FailureCode           string `json:"failure_code,omitempty"`
	RedactionPolicyID     string `json:"redaction_policy_id"`
	FinalizedAt           string `json:"finalized_at"`
}

func runBuildPerformanceFixture(args []string, env map[string]string) int {
	parsed, err := parsePerformanceFixtureBuildArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if !suiteservices.SuiteActive(env) {
		fmt.Fprintln(os.Stderr, "build-performance-fixture requires an active owned suite")
		return 1
	}
	resultsRoot, err := suiteservices.ResolveResultsRoot(env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve performance fixture result root: %v\n", err)
		return 1
	}
	expectedArtifact := filepath.Join(resultsRoot, suiteservices.ResolveRunID(env), "performance-fixtures", parsed.SnapshotKey, "snapshot-build.json")
	if filepath.Clean(parsed.ArtifactFile) != filepath.Clean(expectedArtifact) {
		fmt.Fprintln(os.Stderr, "performance fixture build artifact must use the current run-scoped canonical path")
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	artifact, err := buildPerformanceFixture(ctx, env, parsed)
	if err != nil {
		failed := failedPerformanceFixtureBuild(parsed, "contribution_invalid")
		if writeErr := writeImmutableJSON(parsed.ArtifactFile, failed); writeErr != nil {
			fmt.Fprintf(os.Stderr, "build performance fixture: %v; retain failure artifact: %v\n", err, writeErr)
		} else {
			fmt.Fprintf(os.Stderr, "build performance fixture: %v\n", err)
		}
		return 1
	}
	if err := writeImmutableJSON(parsed.ArtifactFile, artifact); err != nil {
		suiteID := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.SuiteIDEnv))
		adminDSN := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGAdminDSNEnv))
		_ = dropPerformanceFixtureTemplate(context.Background(), adminDSN, performanceFixtureTemplateName(suiteID, parsed.SnapshotKey))
		_ = removePerformanceFixtureRuntimeRoot(suiteID, parsed.SnapshotKey)
		fmt.Fprintf(os.Stderr, "write performance fixture build artifact: %v\n", err)
		return 1
	}
	return 0
}

func parsePerformanceFixtureBuildArgs(args []string) (performanceFixtureBuildArgs, error) {
	values, err := parseFlagPairs(args, map[string]struct{}{
		"--fixture-profile":        {},
		"--snapshot-key":           {},
		"--migration-digest":       {},
		"--source-contract-digest": {},
		"--builder-unit-id":        {},
		"--artifact-file":          {},
	})
	if err != nil {
		return performanceFixtureBuildArgs{}, err
	}
	result := performanceFixtureBuildArgs{
		FixtureProfileID:     strings.TrimSpace(values["--fixture-profile"]),
		SnapshotKey:          strings.TrimSpace(values["--snapshot-key"]),
		MigrationDigest:      strings.TrimSpace(values["--migration-digest"]),
		SourceContractDigest: strings.TrimSpace(values["--source-contract-digest"]),
		BuilderUnitID:        strings.TrimSpace(values["--builder-unit-id"]),
		ArtifactFile:         strings.TrimSpace(values["--artifact-file"]),
	}
	if result.FixtureProfileID != fixture.LargeGridProfileID ||
		!lowerHexDigest(result.SnapshotKey) || !lowerHexDigest(result.MigrationDigest) ||
		!lowerHexDigest(result.SourceContractDigest) ||
		result.BuilderUnitID != "fixture_snapshot:default:"+result.FixtureProfileID+":"+result.SnapshotKey ||
		!filepath.IsAbs(result.ArtifactFile) {
		return performanceFixtureBuildArgs{}, errors.New("build-performance-fixture received an invalid canonical identity")
	}
	return result, nil
}

func buildPerformanceFixture(ctx context.Context, env map[string]string, args performanceFixtureBuildArgs) (performanceFixtureBuildArtifact, error) {
	adminDSN := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGAdminDSNEnv))
	migratedTemplate := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGTemplateDBEnv))
	if adminDSN == "" || migratedTemplate == "" {
		return performanceFixtureBuildArtifact{}, errors.New("performance fixture builder requires suite postgres administration and migrated template")
	}
	if expected := strings.TrimSpace(env["CARTULARY_FIXTURE_MIGRATION_DIGEST"]); expected != "" && expected != args.MigrationDigest {
		return performanceFixtureBuildArtifact{}, errors.New("performance fixture migration digest diverges from admitted builder")
	}
	if expected := strings.TrimSpace(env["CARTULARY_FIXTURE_SOURCE_CONTRACT_DIGEST"]); expected != "" && expected != args.SourceContractDigest {
		return performanceFixtureBuildArtifact{}, errors.New("performance fixture source contract digest diverges from admitted builder")
	}
	suiteID := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.SuiteIDEnv))
	templateName := performanceFixtureTemplateName(suiteID, args.SnapshotKey)
	runtimeRoot := performanceFixtureRuntimeRoot(suiteID, args.SnapshotKey)
	if err := os.MkdirAll(filepath.Dir(runtimeRoot), 0o700); err != nil {
		return performanceFixtureBuildArtifact{}, fmt.Errorf("create performance fixture suite runtime root: %w", err)
	}
	if err := cloneDatabase(ctx, adminDSN, migratedTemplate, templateName); err != nil {
		return performanceFixtureBuildArtifact{}, err
	}
	if err := markPerformanceFixtureTemplateOwned(ctx, adminDSN, templateName, performanceFixtureTemplateOwner(suiteID, args.SnapshotKey)); err != nil {
		_ = dropPerformanceFixtureTemplate(context.Background(), adminDSN, templateName)
		return performanceFixtureBuildArtifact{}, err
	}
	sealed := false
	defer func() {
		if !sealed {
			_ = dropPerformanceFixtureTemplate(context.Background(), adminDSN, templateName)
			_ = removePerformanceFixtureRuntimeRoot(suiteID, args.SnapshotKey)
		}
	}()
	dsn, err := replaceDatabaseInDSN(adminDSN, templateName)
	if err != nil {
		return performanceFixtureBuildArtifact{}, err
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return performanceFixtureBuildArtifact{}, fmt.Errorf("open performance fixture template: %w", err)
	}
	repoRoot, err := suiteservices.FindRepoRoot()
	if err != nil {
		pool.Close()
		return performanceFixtureBuildArtifact{}, fmt.Errorf("resolve performance fixture repository root: %w", err)
	}
	bootstrapManifest := filepath.Join(repoRoot, "configs", "dev", "bootstrap-admin.json")
	if err := bootstrap.Preflight(ctx, bootstrap.Settings{ManifestPath: bootstrapManifest}, pool); err != nil {
		pool.Close()
		return performanceFixtureBuildArtifact{}, fmt.Errorf("bootstrap performance fixture template: %w", err)
	}
	actor, err := authn.NewStore(pool).GetUserByNormalizedEmail(ctx, "dev-admin@example.test")
	if err != nil {
		pool.Close()
		return performanceFixtureBuildArtifact{}, fmt.Errorf("resolve performance fixture actor: %w", err)
	}
	owners, err := appsupport.NewPerformanceFixtureOwners(pool, conflicttest.NewCodec(args.SnapshotKey[:16]))
	if err != nil {
		pool.Close()
		return performanceFixtureBuildArtifact{}, err
	}
	assembler, err := performancefixtureassembly.NewProduction(pool, actor, owners)
	if err != nil {
		pool.Close()
		return performanceFixtureBuildArtifact{}, err
	}
	bundle, err := fixture.GenerateRuntimeBundle(args.SnapshotKey)
	if err != nil {
		pool.Close()
		return performanceFixtureBuildArtifact{}, err
	}
	if _, err := fixture.WriteRuntimeBundle(runtimeRoot, bundle); err != nil {
		pool.Close()
		return performanceFixtureBuildArtifact{}, err
	}
	state := &fixture.BuildState{SnapshotKey: args.SnapshotKey, Seed: performanceFixtureSeed, RuntimeBundle: bundle}
	result, err := assembler.Assemble(ctx, state)
	if err != nil {
		pool.Close()
		return performanceFixtureBuildArtifact{}, err
	}
	if err := fixture.ValidateReceiptRedaction(result, state); err != nil {
		pool.Close()
		return performanceFixtureBuildArtifact{}, err
	}
	pool.Close()
	if err := requireNoDatabaseConnections(ctx, adminDSN, templateName); err != nil {
		return performanceFixtureBuildArtifact{}, err
	}
	if err := sealPerformanceFixtureTemplate(ctx, adminDSN, templateName, performanceFixtureTemplateOwner(suiteID, args.SnapshotKey)); err != nil {
		return performanceFixtureBuildArtifact{}, err
	}
	sealed = true
	return performanceFixtureBuildArtifact{
		SchemaID:                 performanceFixtureSnapshotSchemaID,
		FixtureProfileID:         args.FixtureProfileID,
		FixtureVersion:           performanceFixtureVersion,
		Seed:                     performanceFixtureSeed,
		SnapshotKeySchemaID:      "cartulary.performance_fixture_snapshot_key.v1",
		SnapshotKey:              args.SnapshotKey,
		MigrationDigest:          args.MigrationDigest,
		SourceContractDigest:     args.SourceContractDigest,
		BuilderUnitID:            args.BuilderUnitID,
		BuildOrdinal:             1,
		State:                    "sealed",
		ContributionReceipts:     result.Receipts,
		SemanticValidationDigest: result.SemanticValidationDigest,
		Validation: performanceFixtureValidation{
			ExactCounts:              true,
			RelationshipDistribution: result.Validation.RelationshipDistribution,
			DefaultView:              result.Validation.DefaultView,
			Authorization:            result.Validation.Authorization,
			ProjectionReadiness:      result.Validation.ProjectionReadiness,
			NoActiveSessions:         result.Validation.NoActiveSessions,
			ConnectionsClosed:        true,
		},
		RedactionPolicyID: performanceFixtureRedactionPolicyID,
		CreatedAt:         time.Now().UTC().Format(time.RFC3339Nano),
	}, nil
}

func failedPerformanceFixtureBuild(args performanceFixtureBuildArgs, code string) performanceFixtureBuildArtifact {
	digest := sha256.Sum256([]byte("cartulary.performance-fixture.failed\x00" + args.SnapshotKey + "\x00" + code))
	return performanceFixtureBuildArtifact{
		SchemaID:                 performanceFixtureSnapshotSchemaID,
		FixtureProfileID:         args.FixtureProfileID,
		FixtureVersion:           performanceFixtureVersion,
		Seed:                     performanceFixtureSeed,
		SnapshotKeySchemaID:      "cartulary.performance_fixture_snapshot_key.v1",
		SnapshotKey:              args.SnapshotKey,
		MigrationDigest:          args.MigrationDigest,
		SourceContractDigest:     args.SourceContractDigest,
		BuilderUnitID:            args.BuilderUnitID,
		BuildOrdinal:             1,
		State:                    "failed",
		ContributionReceipts:     []fixture.Receipt{},
		SemanticValidationDigest: "sha256:" + hex.EncodeToString(digest[:]),
		Validation:               performanceFixtureValidation{},
		FailureCode:              code,
		RedactionPolicyID:        performanceFixtureRedactionPolicyID,
		CreatedAt:                time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func preparePerformanceWebE2EFixture(ctx context.Context, env map[string]string) (webE2EFixture, error) {
	profileID := strings.TrimSpace(env["CARTULARY_FIXTURE_PROFILE_ID"])
	key := strings.TrimSpace(env["CARTULARY_FIXTURE_SNAPSHOT_KEY"])
	builderID := strings.TrimSpace(env["CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID"])
	rowID := strings.TrimSpace(env["CARTULARY_FIXTURE_ROW_ID"])
	predicateID := strings.TrimSpace(env["CARTULARY_FIXTURE_PREDICATE_ID"])
	leaseID := strings.TrimSpace(env["CARTULARY_FIXTURE_CLONE_LEASE_ID"])
	ordinal, err := strconv.Atoi(strings.TrimSpace(env["CARTULARY_FIXTURE_CLONE_ORDINAL"]))
	if profileID != fixture.LargeGridProfileID || !lowerHexDigest(key) ||
		builderID != "fixture_snapshot:default:"+profileID+":"+key || err != nil || ordinal < 1 ||
		!safeCatalogIdentity(rowID) || !strings.HasPrefix(predicateID, "perf.") || leaseID == "" {
		return webE2EFixture{}, errors.New("performance browser clone requires complete admitted profile identity")
	}
	adminDSN := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGAdminDSNEnv))
	suiteID := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.SuiteIDEnv))
	templateName := performanceFixtureTemplateName(suiteID, key)
	if err := validatePerformanceFixtureTemplate(ctx, adminDSN, templateName, performanceFixtureTemplateOwner(suiteID, key)); err != nil {
		return webE2EFixture{}, err
	}
	cloneName, err := newPerformanceFixtureCloneName()
	if err != nil {
		return webE2EFixture{}, err
	}
	if err := cloneDatabase(ctx, adminDSN, templateName, cloneName); err != nil {
		return webE2EFixture{}, err
	}
	dsn, err := replaceDatabaseInDSN(adminDSN, cloneName)
	if err != nil {
		_ = dropDatabase(context.Background(), adminDSN, cloneName)
		return webE2EFixture{}, err
	}
	s3Harness, err := s3test.StartShared(ctx)
	if err != nil {
		_ = dropDatabase(context.Background(), adminDSN, cloneName)
		return webE2EFixture{}, fmt.Errorf("attach suite object-store: %w", err)
	}
	bucket, err := s3Harness.BootstrapBucket(ctx, "web-e2e")
	if err != nil {
		_ = dropDatabase(context.Background(), adminDSN, cloneName)
		return webE2EFixture{}, err
	}
	sourceBundle := filepath.Join(performanceFixtureRuntimeRoot(suiteID, key), fixture.RuntimeBundleName)
	destinationRoot := performanceFixtureCloneRuntimeRoot(suiteID, leaseID)
	if err := os.MkdirAll(filepath.Dir(destinationRoot), 0o700); err != nil {
		_ = dropDatabase(context.Background(), adminDSN, cloneName)
		_ = s3Harness.CleanupBucket(context.Background(), bucket)
		return webE2EFixture{}, err
	}
	bundlePath, err := fixture.CopyRuntimeBundle(sourceBundle, destinationRoot)
	if err != nil {
		_ = dropDatabase(context.Background(), adminDSN, cloneName)
		_ = s3Harness.CleanupBucket(context.Background(), bucket)
		return webE2EFixture{}, err
	}
	return webE2EFixture{
		DatabaseName:      cloneName,
		DSN:               dsn,
		Bucket:            bucket,
		S3Endpoint:        s3Harness.Endpoint,
		S3AccessKey:       s3Harness.AccessKey,
		S3SecretKey:       s3Harness.SecretKey,
		S3Secure:          s3Harness.Secure,
		FixtureProfileID:  profileID,
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

func cleanupPerformanceFixtureLease(ctx context.Context, deps dependencies, env map[string]string, metadata webE2EMetadata) error {
	artifact := performanceFixtureLeaseArtifact{
		SchemaID:          performanceFixtureLeaseSchemaID,
		FixtureProfileID:  metadata.FixtureProfileID,
		SnapshotKey:       metadata.SnapshotKey,
		BuilderUnitID:     metadata.BuilderUnitID,
		RowID:             metadata.RowID,
		PredicateID:       metadata.PredicateID,
		LeaseID:           metadata.CloneLeaseID,
		CloneOrdinal:      metadata.CloneOrdinal,
		CreationState:     "created",
		IsolationResult:   "isolated",
		ProcessCleanup:    env["CARTULARY_FIXTURE_PROCESS_CLEANUP_COMPLETE"] == "1",
		RedactionPolicyID: performanceFixtureRedactionPolicyID,
		FinalizedAt:       time.Now().UTC().Format(time.RFC3339Nano),
	}
	sessionErr := deps.cleanupWebE2ESessions(ctx, env, metadata.DatabaseName)
	artifact.SessionCleanup = sessionErr == nil
	credentialErr := fixture.RemoveRuntimeBundle(metadata.RuntimeBundleRoot)
	artifact.CredentialCopyCleanup = credentialErr == nil
	leakErr := deps.detectWebE2ELeaks(ctx, []webE2EMetadata{metadata}, env)
	var dbErr error
	if leakErr == nil {
		dbErr = deps.cleanupWebE2EDB(ctx, metadata, env)
	} else {
		dbErr = leakErr
	}
	artifact.DatabaseCleanup = dbErr == nil
	bucketErr := deps.cleanupWebE2EBucket(ctx, metadata, env)
	artifact.BucketCleanup = bucketErr == nil
	if sessionErr != nil || credentialErr != nil || dbErr != nil || bucketErr != nil || !artifact.ProcessCleanup {
		artifact.CleanupState = "failed"
		switch {
		case sessionErr != nil:
			artifact.FailureCode = "session_cleanup_failed"
		case credentialErr != nil:
			artifact.FailureCode = "credential_cleanup_failed"
		case dbErr != nil:
			artifact.FailureCode = "database_cleanup_failed"
		case bucketErr != nil:
			artifact.FailureCode = "bucket_cleanup_failed"
		default:
			artifact.FailureCode = "process_cleanup_failed"
		}
		return errors.Join(sessionErr, credentialErr, dbErr, bucketErr, writePerformanceFixtureLeaseArtifact(env, artifact))
	}
	artifact.CleanupState = "complete"
	return writePerformanceFixtureLeaseArtifact(env, artifact)
}

func cleanupPreparedWebE2EFixture(ctx context.Context, deps dependencies, env map[string]string, metadata webE2EMetadata) error {
	if metadata.FixtureProfileID == "" {
		return cleanupWebE2EFixture(ctx, deps, env, metadata)
	}
	cleanupEnv := cloneEnv(env)
	cleanupEnv["CARTULARY_FIXTURE_PROCESS_CLEANUP_COMPLETE"] = "1"
	return cleanupPerformanceFixtureLease(ctx, deps, cleanupEnv, metadata)
}

func writeFailedPerformanceFixtureCloneArtifact(env map[string]string, failureCode string) error {
	profileID := strings.TrimSpace(env["CARTULARY_FIXTURE_PROFILE_ID"])
	if profileID == "" {
		return nil
	}
	ordinal, err := strconv.Atoi(strings.TrimSpace(env["CARTULARY_FIXTURE_CLONE_ORDINAL"]))
	if err != nil {
		return err
	}
	artifact := performanceFixtureLeaseArtifact{
		SchemaID:          performanceFixtureLeaseSchemaID,
		FixtureProfileID:  profileID,
		SnapshotKey:       strings.TrimSpace(env["CARTULARY_FIXTURE_SNAPSHOT_KEY"]),
		BuilderUnitID:     strings.TrimSpace(env["CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID"]),
		RowID:             strings.TrimSpace(env["CARTULARY_FIXTURE_ROW_ID"]),
		PredicateID:       strings.TrimSpace(env["CARTULARY_FIXTURE_PREDICATE_ID"]),
		LeaseID:           strings.TrimSpace(env["CARTULARY_FIXTURE_CLONE_LEASE_ID"]),
		CloneOrdinal:      ordinal,
		CreationState:     "failed",
		IsolationResult:   "not_checked",
		CleanupState:      "failed",
		FailureCode:       failureCode,
		RedactionPolicyID: performanceFixtureRedactionPolicyID,
		FinalizedAt:       time.Now().UTC().Format(time.RFC3339Nano),
	}
	return writePerformanceFixtureLeaseArtifact(env, artifact)
}

func revokePerformanceFixtureSessions(ctx context.Context, env map[string]string, databaseName string) error {
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

func writePerformanceFixtureLeaseArtifact(env map[string]string, artifact performanceFixtureLeaseArtifact) error {
	if err := validatePerformanceFixtureLeaseArtifactIdentity(artifact); err != nil {
		return err
	}
	resultsRoot, err := suiteservices.ResolveResultsRoot(env)
	if err != nil {
		return err
	}
	runID := suiteservices.ResolveRunID(env)
	file := filepath.Join(resultsRoot, runID, "performance-fixtures", artifact.SnapshotKey, "leases", artifact.RowID+".json")
	return writeImmutableJSON(file, artifact)
}

func validatePerformanceFixtureLeaseArtifactIdentity(artifact performanceFixtureLeaseArtifact) error {
	if artifact.SchemaID != performanceFixtureLeaseSchemaID ||
		artifact.FixtureProfileID != fixture.LargeGridProfileID ||
		!lowerHexDigest(artifact.SnapshotKey) ||
		artifact.BuilderUnitID != "fixture_snapshot:default:"+artifact.FixtureProfileID+":"+artifact.SnapshotKey ||
		!safeCatalogIdentity(artifact.RowID) ||
		!safeCatalogIdentity(artifact.PredicateID) || !strings.HasPrefix(artifact.PredicateID, "perf.") ||
		artifact.LeaseID == "" || len(artifact.LeaseID) > 255 || artifact.CloneOrdinal < 1 {
		return errors.New("performance fixture lease artifact identity is invalid")
	}
	return nil
}

func cloneDatabase(ctx context.Context, adminDSN string, source string, target string) error {
	if !safePostgresIdentifier(source) || !safePostgresIdentifier(target) {
		return errors.New("performance fixture database identity is unsafe")
	}
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close()
	_, err = admin.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE "%s" TEMPLATE "%s"`, target, source))
	if err != nil {
		return fmt.Errorf("clone performance fixture database: %w", err)
	}
	return nil
}

func sealPerformanceFixtureTemplate(ctx context.Context, adminDSN string, name string, owner string) error {
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close()
	for _, statement := range []string{
		fmt.Sprintf(`ALTER DATABASE "%s" WITH ALLOW_CONNECTIONS false`, name),
		fmt.Sprintf(`ALTER DATABASE "%s" WITH IS_TEMPLATE true`, name),
		fmt.Sprintf(`COMMENT ON DATABASE "%s" IS '%s'`, name, owner),
	} {
		if _, err := admin.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("seal performance fixture template: %w", err)
		}
	}
	return validatePerformanceFixtureTemplate(ctx, adminDSN, name, owner)
}

func markPerformanceFixtureTemplateOwned(ctx context.Context, adminDSN string, name string, owner string) error {
	if !safePostgresIdentifier(name) || strings.Contains(owner, "'") {
		return errors.New("performance fixture template ownership identity is unsafe")
	}
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close()
	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`COMMENT ON DATABASE "%s" IS '%s'`, name, owner)); err != nil {
		return fmt.Errorf("mark performance fixture template ownership: %w", err)
	}
	return nil
}

func validatePerformanceFixtureTemplate(ctx context.Context, adminDSN string, name string, owner string) error {
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close()
	var isTemplate bool
	var allowConnections bool
	var comment sql.NullString
	if err := admin.QueryRowContext(ctx, `SELECT datistemplate, datallowconn, shobj_description(oid, 'pg_database') FROM pg_database WHERE datname = $1`, name).Scan(&isTemplate, &allowConnections, &comment); err != nil {
		return fmt.Errorf("inspect performance fixture template: %w", err)
	}
	if !isTemplate || allowConnections || !comment.Valid || comment.String != owner {
		return errors.New("performance fixture template is not sealed for this suite and snapshot key")
	}
	return nil
}

func requireNoDatabaseConnections(ctx context.Context, adminDSN string, name string) error {
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close()
	var count int
	if err := admin.QueryRowContext(ctx, `SELECT COUNT(*) FROM pg_stat_activity WHERE datname = $1`, name).Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("performance fixture template has %d unknown open connection(s)", count)
	}
	return nil
}

func dropPerformanceFixtureTemplate(ctx context.Context, adminDSN string, name string) error {
	if adminDSN == "" || !safePostgresIdentifier(name) {
		return errors.New("performance fixture template cleanup identity is incomplete")
	}
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close()
	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`ALTER DATABASE "%s" WITH IS_TEMPLATE false`, name)); err != nil {
		return err
	}
	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, name)); err != nil {
		return err
	}
	return nil
}

func performanceFixtureTemplateName(suiteID string, key string) string {
	return fmt.Sprintf("ct_pfs_%s_%s", suiteservices.ShortHash(suiteID, 8), key[:12])
}

func performanceFixtureTemplateOwner(suiteID string, key string) string {
	return "cartulary.performance-fixture:" + suiteservices.ShortHash(suiteID, 16) + ":" + key
}

func performanceFixtureRuntimeRoot(suiteID string, key string) string {
	return filepath.Join(os.TempDir(), "cartulary-performance-fixture-runtime", suiteservices.ShortHash(suiteID, 16), key)
}

func performanceFixtureCloneRuntimeRoot(suiteID string, leaseID string) string {
	return filepath.Join(
		os.TempDir(),
		"cartulary-performance-fixture-clones",
		suiteservices.ShortHash(suiteID, 16),
		suiteservices.ShortHash(leaseID, 24),
	)
}

func removePerformanceFixtureRuntimeRoot(suiteID string, key string) error {
	return fixture.RemoveRuntimeBundle(performanceFixtureRuntimeRoot(suiteID, key))
}

func cleanupPerformanceFixtureSuite(ctx context.Context, env map[string]string) error {
	adminDSN := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGAdminDSNEnv))
	suiteID := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.SuiteIDEnv))
	if adminDSN == "" || suiteID == "" {
		return nil
	}
	suiteRuntimeRoot := filepath.Join(os.TempDir(), "cartulary-performance-fixture-runtime", suiteservices.ShortHash(suiteID, 16))
	cloneRuntimeRoot := filepath.Join(os.TempDir(), "cartulary-performance-fixture-clones", suiteservices.ShortHash(suiteID, 16))
	foundRuntimeRoot := false
	for _, runtimeRoot := range []string{suiteRuntimeRoot, cloneRuntimeRoot} {
		if _, err := os.Lstat(runtimeRoot); err == nil {
			foundRuntimeRoot = true
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if !foundRuntimeRoot {
		return nil
	}
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	prefix := "cartulary.performance-fixture:" + suiteservices.ShortHash(suiteID, 16) + ":"
	rows, err := admin.QueryContext(ctx, `SELECT datname, shobj_description(oid, 'pg_database') FROM pg_database WHERE shobj_description(oid, 'pg_database') LIKE $1 ORDER BY datname LIMIT 9`, prefix+"%")
	if err != nil {
		_ = admin.Close()
		return err
	}
	type ownedTemplate struct{ name, owner string }
	var owned []ownedTemplate
	for rows.Next() {
		var candidate ownedTemplate
		if err := rows.Scan(&candidate.name, &candidate.owner); err != nil {
			_ = rows.Close()
			_ = admin.Close()
			return err
		}
		owned = append(owned, candidate)
	}
	if err := rows.Close(); err != nil {
		_ = admin.Close()
		return err
	}
	if len(owned) > 8 {
		_ = admin.Close()
		return errors.New("performance fixture suite cleanup exceeded its bounded template scope")
	}
	wantNamePrefix := "ct_pfs_" + suiteservices.ShortHash(suiteID, 8) + "_"
	for _, candidate := range owned {
		key := strings.TrimPrefix(candidate.owner, prefix)
		if !lowerHexDigest(key) || candidate.name != wantNamePrefix+key[:12] {
			_ = admin.Close()
			return errors.New("performance fixture suite cleanup rejected an unowned template")
		}
		if _, err := admin.ExecContext(ctx, fmt.Sprintf(`ALTER DATABASE "%s" WITH IS_TEMPLATE false`, candidate.name)); err != nil {
			_ = admin.Close()
			return err
		}
		if _, err := admin.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE "%s" WITH (FORCE)`, candidate.name)); err != nil {
			_ = admin.Close()
			return err
		}
	}
	if err := admin.Close(); err != nil {
		return err
	}
	return errors.Join(os.RemoveAll(suiteRuntimeRoot), os.RemoveAll(cloneRuntimeRoot))
}

func cleanupStalePerformanceFixtureRuntimeRoots(now time.Time) error {
	return cleanupStalePerformanceFixtureRuntimeRootsUnder(now, []string{
		filepath.Join(os.TempDir(), "cartulary-performance-fixture-runtime"),
		filepath.Join(os.TempDir(), "cartulary-performance-fixture-clones"),
	})
}

func cleanupStalePerformanceFixtureRuntimeRootsUnder(now time.Time, bases []string) error {
	var errs []error
	for _, base := range bases {
		entries, err := os.ReadDir(base)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			errs = append(errs, err)
			continue
		}
		slices.SortFunc(entries, func(left, right os.DirEntry) int {
			return strings.Compare(left.Name(), right.Name())
		})
		admitted := 0
		for _, entry := range entries {
			if admitted >= performanceFixtureRuntimeJanitorLimit {
				break
			}
			if len(entry.Name()) != 16 || !lowerHexString(entry.Name()) {
				continue
			}
			candidate := filepath.Join(base, entry.Name())
			info, err := os.Lstat(candidate)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				continue
			}
			admitted++
			if now.Sub(info.ModTime()) < performanceFixtureRuntimeStaleAge {
				continue
			}
			if err := os.RemoveAll(candidate); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func newPerformanceFixtureCloneName() (string, error) {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", err
	}
	return "ct_" + hex.EncodeToString(data[:]) + "_web_e2e", nil
}

func lowerHexDigest(value string) bool {
	if len(value) != 64 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func lowerHexString(value string) bool {
	if value == "" || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func safePostgresIdentifier(value string) bool {
	if value == "" || len(value) > 63 {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			continue
		}
		return false
	}
	return true
}

func safeCatalogIdentity(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' || char == '.' {
			continue
		}
		return false
	}
	return true
}

func writeImmutableJSON(file string, value any) error {
	if !filepath.IsAbs(file) {
		return errors.New("immutable artifact path must be absolute")
	}
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		return err
	}
	handle, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := handle.Write(append(payload, '\n')); err != nil {
		_ = handle.Close()
		_ = os.Remove(file)
		return err
	}
	if err := handle.Sync(); err != nil {
		_ = handle.Close()
		_ = os.Remove(file)
		return err
	}
	return handle.Close()
}
