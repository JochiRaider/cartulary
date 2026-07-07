package reportcomposition

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	ReleaseTupleReasonCompositionNotFound       = "composition_not_found"
	ReleaseTupleReasonVersionNotFound           = "composition_version_not_found"
	ReleaseTupleReasonTemplateMismatch          = "composition_template_mismatch"
	ReleaseTupleReasonDigestMismatch            = "composition_digest_mismatch"
	ReleaseTupleReasonInvalidCompositionVersion = "invalid_value"
)

type ReleaseTuple struct {
	CompositionID      uuid.UUID
	CompositionVersion string
	CompositionSHA256  string
}

type ResolvedReleaseTuple struct {
	CompositionID        uuid.UUID
	CompositionVersion   string
	VersionNumber        int64
	CompositionSHA256    string
	TemplateID           string
	TemplateVersion      string
	CanonicalComposition []byte
}

type ReleaseTupleError struct {
	Field      string
	ReasonCode string
}

func (e *ReleaseTupleError) Error() string {
	if e == nil || e.ReasonCode == "" {
		return "reportcomposition: invalid release tuple"
	}
	return "reportcomposition: invalid release tuple: " + e.ReasonCode
}

func ParseCompositionVersionNumber(value string) (int64, error) {
	if !compositionVersionPattern.MatchString(value) {
		return 0, fmt.Errorf("invalid composition version %q", value)
	}
	parsed, err := strconv.ParseInt(strings.TrimPrefix(value, "v"), 10, 64)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("invalid composition version %q", value)
	}
	return parsed, nil
}

func ResolveReleaseTupleTx(ctx context.Context, tx pgx.Tx, expectedIncidentID uuid.UUID, expectedTemplateID string, expectedTemplateVersion string, tuple ReleaseTuple) (ResolvedReleaseTuple, error) {
	versionNumber, err := ParseCompositionVersionNumber(tuple.CompositionVersion)
	if err != nil {
		return ResolvedReleaseTuple{}, &ReleaseTupleError{Field: "composition_version", ReasonCode: ReleaseTupleReasonInvalidCompositionVersion}
	}

	row := tx.QueryRow(ctx, `
SELECT template_id, template_version
  FROM report_compositions
 WHERE composition_id = $1
   AND incident_id = $2
`, tuple.CompositionID, expectedIncidentID)
	var templateID string
	var templateVersion string
	if err := row.Scan(&templateID, &templateVersion); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ResolvedReleaseTuple{}, &ReleaseTupleError{Field: "composition_id", ReasonCode: ReleaseTupleReasonCompositionNotFound}
		}
		return ResolvedReleaseTuple{}, err
	}
	if templateID != expectedTemplateID || templateVersion != expectedTemplateVersion {
		return ResolvedReleaseTuple{}, &ReleaseTupleError{Field: "composition_id", ReasonCode: ReleaseTupleReasonTemplateMismatch}
	}

	row = tx.QueryRow(ctx, `
SELECT composition_sha256, canonical_composition_bytes
  FROM report_composition_versions
 WHERE composition_id = $1
   AND composition_version = $2
`, tuple.CompositionID, versionNumber)
	var compositionSHA string
	var canonicalComposition []byte
	if err := row.Scan(&compositionSHA, &canonicalComposition); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ResolvedReleaseTuple{}, &ReleaseTupleError{Field: "composition_version", ReasonCode: ReleaseTupleReasonVersionNotFound}
		}
		return ResolvedReleaseTuple{}, err
	}
	if compositionSHA != tuple.CompositionSHA256 {
		return ResolvedReleaseTuple{}, &ReleaseTupleError{Field: "composition_sha256", ReasonCode: ReleaseTupleReasonDigestMismatch}
	}

	return ResolvedReleaseTuple{
		CompositionID:        tuple.CompositionID,
		CompositionVersion:   tuple.CompositionVersion,
		VersionNumber:        versionNumber,
		CompositionSHA256:    compositionSHA,
		TemplateID:           templateID,
		TemplateVersion:      templateVersion,
		CanonicalComposition: append([]byte(nil), canonicalComposition...),
	}, nil
}

func BindReleaseTupleTx(ctx context.Context, tx pgx.Tx, releaseID uuid.UUID, tuple ReleaseTuple, releaseScope string, now time.Time) error {
	versionNumber, err := ParseCompositionVersionNumber(tuple.CompositionVersion)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO report_composition_release_bindings (
    composition_id, composition_version, composition_sha256, release_id, release_scope, created_at
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (composition_id, composition_version, release_id) DO NOTHING
`, tuple.CompositionID, versionNumber, tuple.CompositionSHA256, releaseID, releaseScope, now.UTC())
	return err
}
