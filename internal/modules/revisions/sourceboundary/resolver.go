package sourceboundary

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const TokenPrefix = "cartulary.source_boundary.v1:"

type ResolveInput struct {
	IncidentID      uuid.UUID
	IncidentVersion int64
}

type Boundary struct {
	Token         string
	CanonicalJSON []byte
}

type Resolver interface {
	ResolveCurrentTx(context.Context, pgx.Tx, ResolveInput) (Boundary, error)
}

type CurrentResolver struct{}

func NewResolver() Resolver {
	return CurrentResolver{}
}

func (CurrentResolver) ResolveCurrentTx(
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
		Token:         TokenPrefix + hex.EncodeToString(sum[:]),
		CanonicalJSON: append([]byte(nil), encoded...),
	}, nil
}

type canonicalState struct {
	IncidentID               string  `json:"incident_id"`
	IncidentVersion          int64   `json:"incident_version"`
	LatestChangeSetID        *string `json:"latest_change_set_id"`
	LatestChangeSetCreatedAt *string `json:"latest_change_set_created_at"`
}
