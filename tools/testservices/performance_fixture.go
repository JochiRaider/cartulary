package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
	appfixture "github.com/JochiRaider/cartulary/internal/testutil/appsupport/performancefixture"
	"github.com/JochiRaider/cartulary/internal/testutil/performancefixturelifecycle"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func runBuildPerformanceFixture(args []string, env map[string]string) int {
	parsed, profile, err := parsePerformanceFixtureBuildArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if !suiteservices.SuiteActive(env) {
		fmt.Fprintln(os.Stderr, "build-performance-fixture requires an active owned suite")
		return 1
	}
	expectedArtifact, err := performancefixturelifecycle.BuildArtifactPath(env, parsed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve performance fixture result root: %v\n", err)
		return 1
	}
	if filepath.Clean(parsed.ArtifactFile) != filepath.Clean(expectedArtifact) {
		fmt.Fprintln(os.Stderr, "performance fixture build artifact must use the current run-scoped canonical path")
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	artifact, err := performancefixturelifecycle.Build(ctx, env, profile, parsed, appfixture.NewProduction)
	if err != nil {
		failed := performancefixturelifecycle.FailedBuild(profile, parsed, "contribution_invalid")
		if writeErr := performancefixturelifecycle.WriteImmutableJSON(parsed.ArtifactFile, failed); writeErr != nil {
			fmt.Fprintf(os.Stderr, "build performance fixture: %v; retain failure artifact: %v\n", err, writeErr)
		} else {
			fmt.Fprintf(os.Stderr, "build performance fixture: %v\n", err)
		}
		return 1
	}
	if err := performancefixturelifecycle.WriteImmutableJSON(parsed.ArtifactFile, artifact); err != nil {
		_ = performancefixturelifecycle.CleanupFailedBuild(context.Background(), env, parsed)
		fmt.Fprintf(os.Stderr, "write performance fixture build artifact: %v\n", err)
		return 1
	}
	return 0
}

func parsePerformanceFixtureBuildArgs(args []string) (performancefixturelifecycle.BuildArgs, performancefixtureprofile.Profile, error) {
	values, err := parseFlagPairs(args, map[string]struct{}{
		"--fixture-profile": {}, "--snapshot-key": {}, "--migration-digest": {},
		"--source-contract-digest": {}, "--builder-unit-id": {}, "--artifact-file": {},
	})
	if err != nil {
		return performancefixturelifecycle.BuildArgs{}, performancefixtureprofile.Profile{}, err
	}
	result := performancefixturelifecycle.BuildArgs{
		FixtureProfileID:     strings.TrimSpace(values["--fixture-profile"]),
		SnapshotKey:          strings.TrimSpace(values["--snapshot-key"]),
		MigrationDigest:      strings.TrimSpace(values["--migration-digest"]),
		SourceContractDigest: strings.TrimSpace(values["--source-contract-digest"]),
		BuilderUnitID:        strings.TrimSpace(values["--builder-unit-id"]),
		ArtifactFile:         strings.TrimSpace(values["--artifact-file"]),
	}
	profile, err := performancefixturelifecycle.ResolveProfile(result.FixtureProfileID)
	if err != nil {
		return performancefixturelifecycle.BuildArgs{}, performancefixtureprofile.Profile{}, err
	}
	if err := performancefixturelifecycle.ValidateBuildArgs(result, profile); err != nil {
		return performancefixturelifecycle.BuildArgs{}, performancefixtureprofile.Profile{}, err
	}
	return result, profile, nil
}

func preparePerformanceWebE2EFixture(ctx context.Context, env map[string]string) (webE2EFixture, error) {
	prepared, err := performancefixturelifecycle.Prepare(ctx, env)
	if err != nil {
		return webE2EFixture{}, err
	}
	return webE2EFixture{
		DatabaseName: prepared.DatabaseName, DSN: prepared.DSN, Bucket: prepared.Bucket,
		S3Endpoint: prepared.S3Endpoint, S3AccessKey: prepared.S3AccessKey,
		S3SecretKey: prepared.S3SecretKey, S3Secure: prepared.S3Secure,
		FixtureProfileID: prepared.FixtureProfileID, SnapshotKey: prepared.SnapshotKey,
		BuilderUnitID: prepared.BuilderUnitID, RowID: prepared.RowID,
		PredicateID: prepared.PredicateID, CloneLeaseID: prepared.CloneLeaseID,
		CloneOrdinal: prepared.CloneOrdinal, RuntimeBundlePath: prepared.RuntimeBundlePath,
		RuntimeBundleRoot: prepared.RuntimeBundleRoot,
	}, nil
}

func cleanupPerformanceFixtureLease(ctx context.Context, deps dependencies, env map[string]string, metadata webE2EMetadata) error {
	lease := lifecycleMetadata(metadata)
	return performancefixturelifecycle.CleanupLease(ctx, env, lease, performancefixturelifecycle.CleanupPorts{
		CleanupSessions: deps.cleanupWebE2ESessions,
		DetectLeaks: func(ctx context.Context, _ performancefixturelifecycle.LeaseMetadata, env map[string]string) error {
			return deps.detectWebE2ELeaks(ctx, []webE2EMetadata{metadata}, env)
		},
		CleanupDatabase: func(ctx context.Context, _ performancefixturelifecycle.LeaseMetadata, env map[string]string) error {
			return deps.cleanupWebE2EDB(ctx, metadata, env)
		},
		CleanupBucket: func(ctx context.Context, _ performancefixturelifecycle.LeaseMetadata, env map[string]string) error {
			return deps.cleanupWebE2EBucket(ctx, metadata, env)
		},
	})
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
	return performancefixturelifecycle.FailedLease(env, failureCode)
}

func revokePerformanceFixtureSessions(ctx context.Context, env map[string]string, databaseName string) error {
	return performancefixturelifecycle.RevokeSessions(ctx, env, databaseName)
}

func cleanupPerformanceFixtureSuite(ctx context.Context, env map[string]string) error {
	return performancefixturelifecycle.CleanupSuite(ctx, env)
}

func lifecycleMetadata(metadata webE2EMetadata) performancefixturelifecycle.LeaseMetadata {
	return performancefixturelifecycle.LeaseMetadata{
		DatabaseName: metadata.DatabaseName, Bucket: metadata.Bucket,
		FixtureProfileID: metadata.FixtureProfileID, SnapshotKey: metadata.SnapshotKey,
		BuilderUnitID: metadata.BuilderUnitID, RowID: metadata.RowID,
		PredicateID: metadata.PredicateID, CloneLeaseID: metadata.CloneLeaseID,
		CloneOrdinal: metadata.CloneOrdinal, RuntimeBundleRoot: metadata.RuntimeBundleRoot,
	}
}
