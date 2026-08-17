package performancefixturelifecycle

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

type buildStageError struct {
	stage string
	err   error
}

func (e *buildStageError) Error() string { return e.err.Error() }

func (e *buildStageError) Unwrap() error { return e.err }

func withBuildFailureStage(stage string, err error) error {
	return &buildStageError{stage: stage, err: err}
}

type BuildArgs struct {
	FixtureProfileID     string
	SnapshotKey          string
	MigrationDigest      string
	SourceContractDigest string
	BuilderUnitID        string
	ArtifactFile         string
}

type BuildArtifact struct {
	SchemaID                 string                        `json:"schema_id"`
	FixtureProfileID         string                        `json:"fixture_profile_id"`
	FixtureVersion           string                        `json:"fixture_version"`
	Seed                     int                           `json:"seed"`
	SnapshotKeySchemaID      string                        `json:"snapshot_key_schema_id"`
	SnapshotKey              string                        `json:"snapshot_key"`
	MigrationDigest          string                        `json:"migration_digest"`
	SourceContractDigest     string                        `json:"source_contract_digest"`
	BuilderUnitID            string                        `json:"builder_unit_id"`
	BuildOrdinal             int                           `json:"build_ordinal"`
	State                    string                        `json:"state"`
	ContributionReceipts     []BuildReceipt                `json:"contribution_receipts"`
	SemanticValidationDigest string                        `json:"semantic_validation_digest"`
	Validation               BuildValidation               `json:"validation"`
	FailureCode              string                        `json:"failure_code,omitempty"`
	RedactionPolicyID        string                        `json:"redaction_policy_id"`
	CreatedAt                string                        `json:"created_at"`
	ContributionDiagnostics  []BuildContributionDiagnostic `json:"-"`
	SemanticValidationTime   time.Duration                 `json:"-"`
}

type BuildBatchDiagnostic struct {
	Strategy            string `json:"strategy"`
	BatchCount          int    `json:"batch_count"`
	ConfiguredBatchSize int    `json:"configured_batch_size"`
	ItemCount           int    `json:"item_count"`
}

type BuildContributionDiagnostic struct {
	Ordinal        int                   `json:"ordinal"`
	ContributionID string                `json:"contribution_id"`
	OwnerID        string                `json:"owner_id"`
	State          string                `json:"state"`
	DurationMS     int64                 `json:"duration_ms"`
	Batch          *BuildBatchDiagnostic `json:"batch,omitempty"`
}

type BuildDiagnosticArtifact struct {
	SchemaID                     string                        `json:"schema_id"`
	FixtureProfileID             string                        `json:"fixture_profile_id"`
	SnapshotKey                  string                        `json:"snapshot_key"`
	BuilderUnitID                string                        `json:"builder_unit_id"`
	State                        string                        `json:"state"`
	FailureStage                 string                        `json:"failure_stage"`
	ConstructionCount            int                           `json:"construction_count"`
	DurationMS                   int64                         `json:"duration_ms"`
	SemanticValidationDurationMS int64                         `json:"semantic_validation_duration_ms"`
	Contributions                []BuildContributionDiagnostic `json:"contributions"`
	RedactionPolicyID            string                        `json:"redaction_policy_id"`
	CreatedAt                    string                        `json:"created_at"`
}

type ReceiptCount struct {
	CountID string `json:"count_id"`
	Actual  int    `json:"actual"`
}

type BuildReceipt struct {
	ContributionID string         `json:"contribution_id"`
	Version        string         `json:"version"`
	OwnerID        string         `json:"owner_id"`
	Counts         []ReceiptCount `json:"counts"`
}

type ValidationCount struct {
	ExpectationID string `json:"expectation_id"`
	Actual        int    `json:"actual"`
	Passed        bool   `json:"passed"`
}

type ValidationCondition struct {
	ExpectationID string `json:"expectation_id"`
	Actual        bool   `json:"actual"`
	Passed        bool   `json:"passed"`
}

type BuildValidation struct {
	Counts            []ValidationCount     `json:"counts"`
	Conditions        []ValidationCondition `json:"conditions"`
	ConnectionsClosed bool                  `json:"connections_closed"`
}

type leaseArtifact struct {
	SchemaID          string          `json:"schema_id"`
	FixtureProfileID  string          `json:"fixture_profile_id"`
	SnapshotKey       string          `json:"snapshot_key"`
	BuilderUnitID     string          `json:"builder_unit_id"`
	RowID             string          `json:"row_id"`
	PredicateID       string          `json:"predicate_id"`
	LeaseIdentity     string          `json:"lease_identity"`
	CloneOrdinal      int             `json:"clone_ordinal"`
	CreationState     string          `json:"creation_state"`
	IsolationResult   string          `json:"isolation_result"`
	CleanupResults    []CleanupResult `json:"cleanup_results"`
	CleanupState      string          `json:"cleanup_state"`
	FailureCode       string          `json:"failure_code,omitempty"`
	RedactionPolicyID string          `json:"redaction_policy_id"`
	FinalizedAt       string          `json:"finalized_at"`
}

type CleanupResult struct {
	ResourceClass string `json:"resource_class"`
	Outcome       string `json:"outcome"`
}

func ResolveProfile(profileID string) (performancefixtureprofile.Profile, error) {
	profile, ok := performancefixtureprofile.Lookup(strings.TrimSpace(profileID))
	if !ok || profile.Status != "active" {
		return performancefixtureprofile.Profile{}, errors.New("performance fixture profile is unknown or inactive")
	}
	return profile, nil
}

func ValidateBuildArgs(args BuildArgs, profile performancefixtureprofile.Profile) error {
	if args.FixtureProfileID != profile.FixtureProfileID || !lowerHexDigest(args.SnapshotKey) ||
		!lowerHexDigest(args.MigrationDigest) || !lowerHexDigest(args.SourceContractDigest) ||
		args.SourceContractDigest != profile.SourceContractDigest ||
		args.BuilderUnitID != "fixture_snapshot:default:"+profile.FixtureProfileID+":"+args.SnapshotKey ||
		!filepath.IsAbs(args.ArtifactFile) {
		return errors.New("build-performance-fixture received an invalid canonical identity")
	}
	return nil
}

func BuildArtifactPath(env map[string]string, args BuildArgs) (string, error) {
	resultsRoot, err := suiteservices.ResolveResultsRoot(env)
	if err != nil {
		return "", err
	}
	return filepath.Join(resultsRoot, suiteservices.ResolveRunID(env), "performance-fixtures", args.SnapshotKey, "snapshot-build.json"), nil
}

func BuildDiagnosticsPath(env map[string]string, args BuildArgs) (string, error) {
	resultsRoot, err := suiteservices.ResolveResultsRoot(env)
	if err != nil {
		return "", err
	}
	return filepath.Join(resultsRoot, suiteservices.ResolveRunID(env), "performance-fixtures", args.SnapshotKey, "build-diagnostics.json"), nil
}

func SuccessfulBuildDiagnostics(profile performancefixtureprofile.Profile, args BuildArgs, build BuildArtifact, duration time.Duration) BuildDiagnosticArtifact {
	return BuildDiagnosticArtifact{
		SchemaID:                     "cartulary.performance_fixture_build_diagnostics.v2",
		FixtureProfileID:             profile.FixtureProfileID,
		SnapshotKey:                  args.SnapshotKey,
		BuilderUnitID:                args.BuilderUnitID,
		State:                        "sealed",
		FailureStage:                 "none",
		ConstructionCount:            1,
		DurationMS:                   nonNegativeMilliseconds(duration),
		SemanticValidationDurationMS: nonNegativeMilliseconds(build.SemanticValidationTime),
		Contributions:                append([]BuildContributionDiagnostic(nil), build.ContributionDiagnostics...),
		RedactionPolicyID:            profile.RedactionPolicy.PolicyID,
		CreatedAt:                    time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func FailedBuildDiagnostics(profile performancefixtureprofile.Profile, args BuildArgs, duration time.Duration, buildErr error) BuildDiagnosticArtifact {
	stage := "setup"
	semanticValidationDuration := time.Duration(0)
	contributions := []BuildContributionDiagnostic{}
	if diagnostic, ok := fixture.FailureDiagnostics(buildErr); ok {
		stage = diagnostic.Stage
		semanticValidationDuration = diagnostic.SemanticValidationDuration
		contributions = projectBuildDiagnostics(diagnostic.Contributions)
	} else {
		var staged *buildStageError
		if errors.As(buildErr, &staged) {
			stage = staged.stage
		}
	}
	return BuildDiagnosticArtifact{
		SchemaID:                     "cartulary.performance_fixture_build_diagnostics.v2",
		FixtureProfileID:             profile.FixtureProfileID,
		SnapshotKey:                  args.SnapshotKey,
		BuilderUnitID:                args.BuilderUnitID,
		State:                        "failed",
		FailureStage:                 stage,
		ConstructionCount:            1,
		DurationMS:                   nonNegativeMilliseconds(duration),
		SemanticValidationDurationMS: nonNegativeMilliseconds(semanticValidationDuration),
		Contributions:                contributions,
		RedactionPolicyID:            profile.RedactionPolicy.PolicyID,
		CreatedAt:                    time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func nonNegativeMilliseconds(duration time.Duration) int64 {
	if duration < 0 {
		return 0
	}
	return duration.Milliseconds()
}

func FailedBuild(profile performancefixtureprofile.Profile, args BuildArgs, code string) BuildArtifact {
	digest := sha256.Sum256([]byte("cartulary.performance-fixture.failed\x00" + args.SnapshotKey + "\x00" + code))
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
		State:                    "failed",
		ContributionReceipts:     []BuildReceipt{},
		SemanticValidationDigest: "sha256:" + hex.EncodeToString(digest[:]),
		Validation: BuildValidation{
			Counts:     []ValidationCount{},
			Conditions: []ValidationCondition{},
		},
		FailureCode:       code,
		RedactionPolicyID: profile.RedactionPolicy.PolicyID,
		CreatedAt:         time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func writeLeaseArtifact(env map[string]string, profile performancefixtureprofile.Profile, artifact leaseArtifact) error {
	if err := validateLeaseArtifact(profile, artifact); err != nil {
		return err
	}
	resultsRoot, err := suiteservices.ResolveResultsRoot(env)
	if err != nil {
		return err
	}
	file := filepath.Join(resultsRoot, suiteservices.ResolveRunID(env), "performance-fixtures", artifact.SnapshotKey, "leases", artifact.RowID+".json")
	return WriteImmutableJSON(file, artifact)
}

func FailedLease(env map[string]string, failureCode string) error {
	profileID := strings.TrimSpace(env["CARTULARY_FIXTURE_PROFILE_ID"])
	if profileID == "" {
		return nil
	}
	profile, err := ResolveProfile(profileID)
	if err != nil {
		return err
	}
	ordinal, err := strconv.Atoi(strings.TrimSpace(env["CARTULARY_FIXTURE_CLONE_ORDINAL"]))
	if err != nil {
		return err
	}
	leaseIdentity, err := opaqueLeaseIdentity(strings.TrimSpace(env["CARTULARY_FIXTURE_CLONE_LEASE_ID"]))
	if err != nil {
		return err
	}
	return writeLeaseArtifact(env, profile, leaseArtifact{
		SchemaID:          profile.ArtifactPolicy.LeaseSchemaID,
		FixtureProfileID:  profile.FixtureProfileID,
		SnapshotKey:       strings.TrimSpace(env["CARTULARY_FIXTURE_SNAPSHOT_KEY"]),
		BuilderUnitID:     strings.TrimSpace(env["CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID"]),
		RowID:             strings.TrimSpace(env["CARTULARY_FIXTURE_ROW_ID"]),
		PredicateID:       strings.TrimSpace(env["CARTULARY_FIXTURE_PREDICATE_ID"]),
		LeaseIdentity:     leaseIdentity,
		CloneOrdinal:      ordinal,
		CreationState:     "failed",
		IsolationResult:   "not_checked",
		CleanupResults:    cleanupResults("not_acquired", "not_acquired", "not_acquired", "not_acquired", "not_acquired"),
		CleanupState:      "failed",
		FailureCode:       failureCode,
		RedactionPolicyID: profile.RedactionPolicy.PolicyID,
		FinalizedAt:       time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func validateLeaseArtifact(profile performancefixtureprofile.Profile, artifact leaseArtifact) error {
	if artifact.SchemaID != profile.ArtifactPolicy.LeaseSchemaID ||
		artifact.FixtureProfileID != profile.FixtureProfileID || !lowerHexDigest(artifact.SnapshotKey) ||
		artifact.BuilderUnitID != "fixture_snapshot:default:"+profile.FixtureProfileID+":"+artifact.SnapshotKey ||
		!safeCatalogIdentity(artifact.RowID) || !predicateAdmitted(profile, artifact.PredicateID) ||
		!strings.HasPrefix(artifact.LeaseIdentity, "sha256:") || !lowerHexDigest(strings.TrimPrefix(artifact.LeaseIdentity, "sha256:")) || artifact.CloneOrdinal < 1 || len(artifact.CleanupResults) == 0 {
		return errors.New("performance fixture lease artifact identity is invalid")
	}
	return nil
}

func opaqueLeaseIdentity(leaseID string) (string, error) {
	if leaseID == "" || len(leaseID) > 255 {
		return "", errors.New("performance fixture clone lease identity is invalid")
	}
	digest := sha256.Sum256([]byte("cartulary.performance-fixture.lease\x00" + leaseID))
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func cleanupResults(bucket, credentialCopy, database, process, session string) []CleanupResult {
	return []CleanupResult{
		{ResourceClass: "bucket", Outcome: bucket},
		{ResourceClass: "credential_copy", Outcome: credentialCopy},
		{ResourceClass: "database", Outcome: database},
		{ResourceClass: "process", Outcome: process},
		{ResourceClass: "session", Outcome: session},
	}
}

func predicateAdmitted(profile performancefixtureprofile.Profile, predicateID string) bool {
	for _, binding := range profile.VerificationBindings {
		if binding.PredicateID == predicateID {
			return true
		}
	}
	return false
}
