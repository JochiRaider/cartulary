package performancefixture

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
)

const SnapshotKeySchemaID = "cartulary.performance_fixture_snapshot_key.v2"

var lowercaseDigestPattern = regexp.MustCompile("^[a-f0-9]{64}$")
var fixtureProfilePattern = regexp.MustCompile("^[a-z][a-z0-9_]*_v[1-9][0-9]*$")
var fixtureVersionPattern = regexp.MustCompile(`^cartulary\.[a-z0-9_.]+\.v[1-9][0-9]*$`)

type SnapshotKeyInput struct {
	SchemaID             string `json:"schema_id"`
	FixtureProfileID     string `json:"fixture_profile_id"`
	MigrationDigest      string `json:"migration_digest"`
	SourceContractDigest string `json:"source_contract_digest"`
	FixtureVersion       string `json:"fixture_version"`
	Seed                 int    `json:"seed"`
}

func snapshotKeyInput(profile performancefixtureprofile.Profile, migrationDigest string) (SnapshotKeyInput, error) {
	if profile.Status != "active" {
		return SnapshotKeyInput{}, errors.New("performance fixture snapshot profile is not active")
	}
	if !fixtureProfilePattern.MatchString(profile.FixtureProfileID) {
		return SnapshotKeyInput{}, errors.New("performance fixture profile identity is not canonical")
	}
	if !fixtureVersionPattern.MatchString(profile.FixtureVersion) {
		return SnapshotKeyInput{}, errors.New("performance fixture version is not canonical")
	}
	if profile.Seed < 0 {
		return SnapshotKeyInput{}, errors.New("performance fixture seed is not canonical")
	}
	input := SnapshotKeyInput{
		SchemaID:             profile.ArtifactPolicy.SnapshotKeySchemaID,
		FixtureProfileID:     profile.FixtureProfileID,
		MigrationDigest:      migrationDigest,
		SourceContractDigest: profile.SourceContractDigest,
		FixtureVersion:       profile.FixtureVersion,
		Seed:                 profile.Seed,
	}
	if input.SchemaID != SnapshotKeySchemaID {
		return SnapshotKeyInput{}, errors.New("performance fixture snapshot key schema is unsupported")
	}
	return input, nil
}

func validateSnapshotKeyInput(input SnapshotKeyInput) error {
	if !lowercaseDigestPattern.MatchString(input.MigrationDigest) {
		return errors.New("performance fixture migration digest is not canonical")
	}
	if !lowercaseDigestPattern.MatchString(input.SourceContractDigest) {
		return errors.New("performance fixture source contract digest is not canonical")
	}
	if !fixtureVersionPattern.MatchString(input.FixtureVersion) {
		return errors.New("performance fixture version is not canonical")
	}
	if input.Seed < 0 {
		return errors.New("performance fixture seed is not canonical")
	}
	if input.SchemaID != SnapshotKeySchemaID {
		return errors.New("performance fixture snapshot key schema is unsupported")
	}
	if !fixtureProfilePattern.MatchString(input.FixtureProfileID) {
		return errors.New("performance fixture profile identity is not canonical")
	}
	return nil
}

func CanonicalSnapshotKeyInput(profile performancefixtureprofile.Profile, migrationDigest string) ([]byte, error) {
	input, err := snapshotKeyInput(profile, migrationDigest)
	if err != nil {
		return nil, err
	}
	if err := validateSnapshotKeyInput(input); err != nil {
		return nil, err
	}
	envelope := map[string]any{
		"fixture_profile_id":     input.FixtureProfileID,
		"fixture_version":        input.FixtureVersion,
		"migration_digest":       input.MigrationDigest,
		"schema_id":              input.SchemaID,
		"seed":                   input.Seed,
		"source_contract_digest": input.SourceContractDigest,
	}
	return json.Marshal(envelope)
}

func SnapshotKey(profile performancefixtureprofile.Profile, migrationDigest string) (string, error) {
	canonical, err := CanonicalSnapshotKeyInput(profile, migrationDigest)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}
