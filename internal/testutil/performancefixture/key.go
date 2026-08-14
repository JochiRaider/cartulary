package performancefixture

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
)

const (
	SnapshotKeySchemaID = "cartulary.performance_fixture_snapshot_key.v1"
	LargeGridVersion    = "cartulary.perf.large_grid.v1"
	LargeGridSeed       = 20260405
)

var lowercaseDigestPattern = regexp.MustCompile("^[a-f0-9]{64}$")

type SnapshotKeyInput struct {
	SchemaID             string `json:"schema_id"`
	MigrationDigest      string `json:"migration_digest"`
	SourceContractDigest string `json:"source_contract_digest"`
	FixtureVersion       string `json:"fixture_version"`
	Seed                 int    `json:"seed"`
}

func validateSnapshotKeyInput(input SnapshotKeyInput) error {
	if input.SchemaID != SnapshotKeySchemaID {
		return errors.New("performance fixture snapshot key schema is unsupported")
	}
	if !lowercaseDigestPattern.MatchString(input.MigrationDigest) {
		return errors.New("performance fixture migration digest is not canonical")
	}
	if !lowercaseDigestPattern.MatchString(input.SourceContractDigest) {
		return errors.New("performance fixture source contract digest is not canonical")
	}
	if input.FixtureVersion != LargeGridVersion {
		return errors.New("performance fixture version is unsupported")
	}
	if input.Seed != LargeGridSeed {
		return errors.New("performance fixture seed is unsupported")
	}
	return nil
}

func CanonicalSnapshotKeyInput(input SnapshotKeyInput) ([]byte, error) {
	if err := validateSnapshotKeyInput(input); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"fixture_version":        input.FixtureVersion,
		"migration_digest":       input.MigrationDigest,
		"schema_id":              input.SchemaID,
		"seed":                   input.Seed,
		"source_contract_digest": input.SourceContractDigest,
	})
}

func SnapshotKey(input SnapshotKeyInput) (string, error) {
	canonical, err := CanonicalSnapshotKeyInput(input)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}
