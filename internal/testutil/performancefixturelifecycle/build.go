package performancefixturelifecycle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

type AssemblerFactory func(postgres.DB, authn.UserRecord, string, performancefixtureprofile.Profile) (*fixture.Assembler, error)

func Build(
	ctx context.Context,
	env map[string]string,
	profile performancefixtureprofile.Profile,
	args BuildArgs,
	newAssembler AssemblerFactory,
) (BuildArtifact, error) {
	if newAssembler == nil {
		return BuildArtifact{}, errors.New("performance fixture assembler factory is required")
	}
	if err := ValidateBuildArgs(args, profile); err != nil {
		return BuildArtifact{}, err
	}
	adminDSN := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGAdminDSNEnv))
	migratedTemplate := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGTemplateDBEnv))
	if adminDSN == "" || migratedTemplate == "" {
		return BuildArtifact{}, errors.New("performance fixture builder requires suite postgres administration and migrated template")
	}
	if expected := strings.TrimSpace(env["CARTULARY_FIXTURE_MIGRATION_DIGEST"]); expected != "" && expected != args.MigrationDigest {
		return BuildArtifact{}, errors.New("performance fixture migration digest diverges from admitted builder")
	}
	if expected := strings.TrimSpace(env["CARTULARY_FIXTURE_SOURCE_CONTRACT_DIGEST"]); expected != "" && expected != args.SourceContractDigest {
		return BuildArtifact{}, errors.New("performance fixture source contract digest diverges from admitted builder")
	}
	suiteID := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.SuiteIDEnv))
	templateName := templateName(suiteID, args.SnapshotKey)
	runtimeRoot, err := templateRuntimeRoot(env, args.SnapshotKey)
	if err != nil {
		return BuildArtifact{}, err
	}
	if err := os.MkdirAll(filepath.Dir(runtimeRoot), 0o700); err != nil {
		return BuildArtifact{}, fmt.Errorf("create performance fixture runtime parent: %w", err)
	}
	if err := cloneDatabase(ctx, adminDSN, migratedTemplate, templateName); err != nil {
		return BuildArtifact{}, err
	}
	if err := markTemplateOwned(ctx, adminDSN, templateName, templateOwner(suiteID, args.SnapshotKey)); err != nil {
		_ = dropTemplate(context.Background(), adminDSN, templateName)
		return BuildArtifact{}, err
	}
	sealed := false
	defer func() {
		if !sealed {
			_ = dropTemplate(context.Background(), adminDSN, templateName)
			_ = removeTemplateRuntime(env, args.SnapshotKey)
		}
	}()
	dsn, err := replaceDatabaseInDSN(adminDSN, templateName)
	if err != nil {
		return BuildArtifact{}, err
	}
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return BuildArtifact{}, fmt.Errorf("parse performance fixture template pool configuration: %w", err)
	}
	ownedConnections := &ownedConnectionPIDs{}
	poolConfig.AfterConnect = func(_ context.Context, connection *pgx.Conn) error {
		ownedConnections.Add(connection.PgConn().PID())
		return nil
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return BuildArtifact{}, fmt.Errorf("open performance fixture template: %w", err)
	}
	closePool := func() { pool.Close() }
	repoRoot, err := suiteservices.FindRepoRoot()
	if err != nil {
		closePool()
		return BuildArtifact{}, fmt.Errorf("resolve performance fixture repository root: %w", err)
	}
	bootstrapManifest := filepath.Join(repoRoot, "configs", "dev", "bootstrap-admin.json")
	if err := bootstrap.Preflight(ctx, bootstrap.Settings{ManifestPath: bootstrapManifest}, pool); err != nil {
		closePool()
		return BuildArtifact{}, fmt.Errorf("bootstrap performance fixture template: %w", err)
	}
	actor, err := authn.NewStore(pool).GetUserByNormalizedEmail(ctx, "dev-admin@example.test")
	if err != nil {
		closePool()
		return BuildArtifact{}, fmt.Errorf("resolve performance fixture actor: %w", err)
	}
	assembler, err := newAssembler(pool, actor, args.SnapshotKey, profile)
	if err != nil {
		closePool()
		return BuildArtifact{}, err
	}
	bundle, err := fixture.GenerateRuntimeBundle(profile, args.SnapshotKey)
	if err != nil {
		closePool()
		return BuildArtifact{}, err
	}
	if _, err := fixture.WriteRuntimeBundle(profile, runtimeRoot, bundle); err != nil {
		closePool()
		return BuildArtifact{}, err
	}
	state := &fixture.BuildState{
		FixtureProfileID: profile.FixtureProfileID,
		SnapshotKey:      args.SnapshotKey,
		Seed:             profile.Seed,
		RuntimeBundle:    bundle,
	}
	result, err := assembler.Assemble(ctx, state)
	if err != nil {
		closePool()
		return BuildArtifact{}, err
	}
	if err := fixture.ValidateReceiptRedaction(result, state); err != nil {
		closePool()
		return BuildArtifact{}, withBuildFailureStage("finalization", err)
	}
	closePool()
	if err := requireNoConnections(ctx, adminDSN, templateName, ownedConnections); err != nil {
		return BuildArtifact{}, withBuildFailureStage("finalization", err)
	}
	if err := sealTemplate(ctx, adminDSN, templateName, templateOwner(suiteID, args.SnapshotKey)); err != nil {
		return BuildArtifact{}, withBuildFailureStage("finalization", err)
	}
	sealed = true
	return BuildArtifact{
		SchemaID:                 profile.ArtifactPolicy.BuildSchemaID,
		FixtureProfileID:         profile.FixtureProfileID,
		FixtureVersion:           profile.FixtureVersion,
		Seed:                     profile.Seed,
		SnapshotKeySchemaID:      profile.ArtifactPolicy.SnapshotKeySchemaID,
		SnapshotKey:              args.SnapshotKey,
		MigrationDigest:          args.MigrationDigest,
		SourceContractDigest:     args.SourceContractDigest,
		BuilderUnitID:            args.BuilderUnitID,
		BuildOrdinal:             1,
		State:                    "sealed",
		ContributionReceipts:     projectBuildReceipts(result.Receipts),
		SemanticValidationDigest: result.SemanticValidationDigest,
		Validation:               projectBuildValidation(profile, result.Validation),
		RedactionPolicyID:        profile.RedactionPolicy.PolicyID,
		CreatedAt:                time.Now().UTC().Format(time.RFC3339Nano),
		ContributionDiagnostics:  projectBuildDiagnostics(result.ContributionDiagnostics),
		SemanticValidationTime:   result.SemanticValidationTime,
	}, nil
}

type ownedConnectionPIDs struct {
	values sync.Map
}

func (pids *ownedConnectionPIDs) Add(pid uint32) {
	if pids != nil {
		pids.values.Store(pid, struct{}{})
	}
}

func (pids *ownedConnectionPIDs) Contains(pid uint32) bool {
	if pids == nil {
		return false
	}
	_, ok := pids.values.Load(pid)
	return ok
}

func projectBuildDiagnostics(diagnostics []fixture.ContributionDiagnostic) []BuildContributionDiagnostic {
	projected := make([]BuildContributionDiagnostic, len(diagnostics))
	for index, diagnostic := range diagnostics {
		projected[index] = BuildContributionDiagnostic{
			Ordinal:        index + 1,
			ContributionID: diagnostic.ContributionID,
			OwnerID:        diagnostic.OwnerID,
			State:          diagnostic.State,
			DurationMS:     nonNegativeMilliseconds(diagnostic.Duration),
		}
		if diagnostic.Batch != nil {
			projected[index].Batch = &BuildBatchDiagnostic{
				Strategy:            diagnostic.Batch.Strategy,
				BatchCount:          diagnostic.Batch.BatchCount,
				ConfiguredBatchSize: diagnostic.Batch.ConfiguredBatchSize,
				ItemCount:           diagnostic.Batch.ItemCount,
			}
		}
	}
	return projected
}

func projectBuildReceipts(receipts []fixture.Receipt) []BuildReceipt {
	projected := make([]BuildReceipt, len(receipts))
	for index, receipt := range receipts {
		countIDs := make([]string, 0, len(receipt.Counts))
		for countID := range receipt.Counts {
			countIDs = append(countIDs, countID)
		}
		sort.Strings(countIDs)
		counts := make([]ReceiptCount, 0, len(countIDs))
		for _, countID := range countIDs {
			counts = append(counts, ReceiptCount{CountID: countID, Actual: receipt.Counts[countID]})
		}
		projected[index] = BuildReceipt{
			ContributionID: receipt.ContributionID,
			Version:        receipt.Version,
			OwnerID:        receipt.OwnerID,
			Counts:         counts,
		}
	}
	return projected
}

func projectBuildValidation(profile performancefixtureprofile.Profile, validation fixture.SemanticValidation) BuildValidation {
	counts := make([]ValidationCount, 0, len(profile.SemanticExpectations.Counts))
	for _, expectation := range profile.SemanticExpectations.Counts {
		actual := validation.Counts[expectation.ExpectationID]
		counts = append(counts, ValidationCount{
			ExpectationID: expectation.ExpectationID,
			Actual:        actual,
			Passed:        actual == expectation.Exact,
		})
	}
	conditions := make([]ValidationCondition, 0, len(profile.SemanticExpectations.Conditions))
	for _, expectation := range profile.SemanticExpectations.Conditions {
		actual := validation.Conditions[expectation.ExpectationID]
		conditions = append(conditions, ValidationCondition{
			ExpectationID: expectation.ExpectationID,
			Actual:        actual,
			Passed:        actual == expectation.Required,
		})
	}
	return BuildValidation{Counts: counts, Conditions: conditions, ConnectionsClosed: true}
}

func CleanupFailedBuild(ctx context.Context, env map[string]string, args BuildArgs) error {
	adminDSN := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.PGAdminDSNEnv))
	suiteID := strings.TrimSpace(suiteservices.LookupEnvValue(env, suiteservices.SuiteIDEnv))
	return errors.Join(
		dropTemplate(ctx, adminDSN, templateName(suiteID, args.SnapshotKey)),
		removeTemplateRuntime(env, args.SnapshotKey),
	)
}
