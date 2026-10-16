package sourceboundary

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const tokenPrefix = "cartulary.source_boundary.v1:"

type ResolveInput struct {
	IncidentID      uuid.UUID
	IncidentVersion int64
}

type Boundary struct {
	Token         string
	CanonicalJSON []byte
}

type Resolver struct{}

func NewResolver() Resolver {
	return Resolver{}
}

func (Resolver) ResolveCurrentTx(
	ctx context.Context,
	tx pgx.Tx,
	input ResolveInput,
) (Boundary, error) {
	if tx == nil {
		return Boundary{}, errors.New("revisions source boundary transaction is required")
	}
	if err := validateInput(input); err != nil {
		return Boundary{}, err
	}

	var latestID uuid.UUID
	var latestCreatedAt time.Time
	err := tx.QueryRow(ctx, `
SELECT change_set_id, created_at
  FROM change_sets
 WHERE incident_id = $1
 ORDER BY created_at DESC, change_set_id DESC
 LIMIT 1
`, input.IncidentID).Scan(&latestID, &latestCreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return buildBoundary(input, nil, nil)
	}
	if err != nil {
		return Boundary{}, fmt.Errorf("resolve current Revisions source boundary: %w", err)
	}
	return buildBoundary(input, &latestID, &latestCreatedAt)
}

func validateInput(input ResolveInput) error {
	if input.IncidentID == uuid.Nil {
		return errors.New("revisions source boundary incident ID is required")
	}
	if input.IncidentVersion < 1 {
		return errors.New("revisions source boundary incident version must be positive")
	}
	return nil
}

func buildBoundary(
	input ResolveInput,
	latestID *uuid.UUID,
	latestCreatedAt *time.Time,
) (Boundary, error) {
	if err := validateInput(input); err != nil {
		return Boundary{}, err
	}
	state := canonicalState{
		IncidentID:      input.IncidentID.String(),
		IncidentVersion: input.IncidentVersion,
	}
	if latestID != nil || latestCreatedAt != nil {
		if latestID == nil || latestCreatedAt == nil || *latestID == uuid.Nil {
			return Boundary{}, errors.New("revisions source boundary latest change set is incomplete")
		}
		canonicalID := latestID.String()
		canonicalTimestamp := latestCreatedAt.UTC().Format(time.RFC3339Nano)
		state.LatestChangeSetID = &canonicalID
		state.LatestChangeSetCreatedAt = &canonicalTimestamp
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		return Boundary{}, fmt.Errorf("encode Revisions source boundary: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return Boundary{
		Token:         tokenPrefix + hex.EncodeToString(sum[:]),
		CanonicalJSON: append([]byte(nil), encoded...),
	}, nil
}

type canonicalState struct {
	IncidentID               string  `json:"incident_id"`
	IncidentVersion          int64   `json:"incident_version"`
	LatestChangeSetID        *string `json:"latest_change_set_id"`
	LatestChangeSetCreatedAt *string `json:"latest_change_set_created_at"`
}

// ValidateRetained verifies a stored boundary without resolving current state.
// JSONB formatting is allowed; missing, unknown or changed values are not.
// The token continues to hash the owner's ordered bytes, never a JSONB or JCS
// serialization. Keep this codec beside the producer rather than in consumers.
func ValidateRetained(boundary Boundary, incidentID uuid.UUID) error {
	invalid := errors.New("revisions: invalid retained source boundary")
	value, err := canonicaljson.DecodeStrict(boundary.CanonicalJSON)
	if err != nil {
		return invalid
	}
	var state canonicalState
	if json.Unmarshal(boundary.CanonicalJSON, &state) != nil || incidentID == uuid.Nil || state.IncidentID != incidentID.String() {
		return invalid
	}
	var latestID *uuid.UUID
	var latestAt *time.Time
	if state.LatestChangeSetID != nil || state.LatestChangeSetCreatedAt != nil {
		if state.LatestChangeSetID == nil || state.LatestChangeSetCreatedAt == nil {
			return invalid
		}
		id, err := uuid.Parse(*state.LatestChangeSetID)
		if err != nil {
			return invalid
		}
		at, err := time.Parse(time.RFC3339Nano, *state.LatestChangeSetCreatedAt)
		if err != nil || at.IsZero() {
			return invalid
		}
		latestID, latestAt = &id, &at
	}
	expected, err := buildBoundary(ResolveInput{IncidentID: incidentID, IncidentVersion: state.IncidentVersion}, latestID, latestAt)
	if err != nil || expected.Token != boundary.Token {
		return invalid
	}
	expectedValue, err := canonicaljson.DecodeStrict(expected.CanonicalJSON)
	if err != nil {
		return invalid
	}
	actualJSON, err := json.Marshal(value)
	if err != nil {
		return invalid
	}
	expectedJSON, err := json.Marshal(expectedValue)
	if err != nil || !bytes.Equal(actualJSON, expectedJSON) {
		return invalid
	}
	return nil
}
