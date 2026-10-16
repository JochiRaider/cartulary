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
	source, decodeErr := DecodeRenderSource(canonicalComposition)
	if decodeErr != nil || source.SchemaID != CompositionSchemaID || source.IncidentID != expectedIncidentID.String() || source.CompositionID != tuple.CompositionID.String() || source.CompositionVersion == nil || *source.CompositionVersion != tuple.CompositionVersion || source.CompositionSHA256 == nil || *source.CompositionSHA256 != compositionSHA {
		return ResolvedReleaseTuple{}, &ReleaseTupleError{Field: "composition_sha256", ReasonCode: ReleaseTupleReasonDigestMismatch}
	}
	if source.TemplateID != expectedTemplateID || source.TemplateVersion != expectedTemplateVersion {
		return ResolvedReleaseTuple{}, &ReleaseTupleError{Field: "composition_id", ReasonCode: ReleaseTupleReasonTemplateMismatch}
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
	if err != nil {
		return err
	}
	return ValidateRetainedReleaseBindingTx(ctx, tx, releaseID, releaseScope, &tuple)
}

// ValidateRetainedReleaseBindingTx checks the companion's own retained pin.
// A nil tuple means that this release must have no Composition binding. It
// never inserts a missing binding or replaces a historical version.
func ValidateRetainedReleaseBindingTx(ctx context.Context, tx pgx.Tx, releaseID uuid.UUID, scope string, tuple *ReleaseTuple) error {
	invalid := errors.New("reportcomposition: inconsistent retained release binding")
	var version int64
	if tuple != nil {
		var err error
		version, err = ParseCompositionVersionNumber(tuple.CompositionVersion)
		if err != nil {
			return invalid
		}
	}
	rows, err := tx.Query(ctx, `SELECT composition_id,composition_version,composition_sha256,release_scope FROM report_composition_release_bindings WHERE release_id=$1 ORDER BY composition_id,composition_version LIMIT 2`, releaseID)
	if err != nil {
		return err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
		var id uuid.UUID
		var retainedVersion int64
		var digest, retainedScope string
		if err := rows.Scan(&id, &retainedVersion, &digest, &retainedScope); err != nil {
			return err
		}
		if tuple == nil || count != 1 || id != tuple.CompositionID || retainedVersion != version || digest != tuple.CompositionSHA256 || retainedScope != scope {
			return invalid
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if (tuple != nil) != (count == 1) {
		return invalid
	}
	return nil
}
