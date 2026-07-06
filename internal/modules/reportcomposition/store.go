package reportcomposition

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var (
	ErrNotFound             = errors.New("reportcomposition: not found")
	ErrDraftVersionConflict = errors.New("reportcomposition: draft version conflict")
	ErrResourceRetired      = errors.New("reportcomposition: resource retired")
	ErrVersionBound         = errors.New("reportcomposition: version bound")
)

type Store struct {
	db postgres.DB
}

func NewStore(db postgres.DB) *Store {
	return &Store{db: db}
}

func (s *Store) ListResources(ctx context.Context, incidentID uuid.UUID) ([]map[string]any, error) {
	rows, err := s.db.Query(ctx, `
SELECT composition_id, incident_id, created_by_user_id, client_txn_id, template_id, template_version,
       draft_version, authored_against_snapshot_id, deck_ops, diagram_decls, authored_texts,
       latest_composition_version, retired_at, created_at, updated_at
  FROM report_compositions
 WHERE incident_id = $1
 ORDER BY (retired_at IS NOT NULL) ASC, template_id ASC, template_version ASC, composition_id::text ASC
`, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		record, err := scanResource(rows)
		if err != nil {
			return nil, err
		}
		bound, err := s.releaseBoundVersions(ctx, record.CompositionID)
		if err != nil {
			return nil, err
		}
		out = append(out, resourceView(record, bound))
	}
	return out, rows.Err()
}

func (s *Store) GetResource(ctx context.Context, incidentID uuid.UUID, compositionID uuid.UUID) (ResourceRecord, map[string]any, error) {
	record, err := s.getResource(ctx, incidentID, compositionID)
	if err != nil {
		return ResourceRecord{}, nil, err
	}
	bound, err := s.releaseBoundVersions(ctx, compositionID)
	if err != nil {
		return ResourceRecord{}, nil, err
	}
	return record, resourceView(record, bound), nil
}

func (s *Store) GetVersion(ctx context.Context, incidentID uuid.UUID, compositionID uuid.UUID, compositionVersion int64) (VersionRecord, map[string]any, error) {
	if _, err := s.getResource(ctx, incidentID, compositionID); err != nil {
		return VersionRecord{}, nil, err
	}
	version, err := s.getVersion(ctx, compositionID, compositionVersion)
	if err != nil {
		return VersionRecord{}, nil, err
	}
	return version, versionView(version), nil
}

func (s *Store) CreateDraft(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, request CreateDraftRequest, now time.Time) (MutationResult, error) {
	requestHash := hashRequest(request.Normalized)
	key := authn.RouteIdempotencyKey{
		RouteKey:    "report_compositions.create",
		ActorUserID: actorUserID,
		ScopeKey:    incidentID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if replay, ok, err := replayMutation(ctx, tx, key, requestHash); ok || err != nil {
		return replay, err
	}
	row := tx.QueryRow(ctx, `
INSERT INTO report_compositions (
    incident_id, created_by_user_id, client_txn_id, template_id, template_version,
    draft_version, authored_against_snapshot_id, deck_ops, diagram_decls, authored_texts, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, 1, $6, $7, $8, $9, $10, $10)
RETURNING composition_id, incident_id, created_by_user_id, client_txn_id, template_id, template_version,
          draft_version, authored_against_snapshot_id, deck_ops, diagram_decls, authored_texts,
          latest_composition_version, retired_at, created_at, updated_at
`, incidentID, actorUserID, request.ClientTxnID, request.TemplateID, request.TemplateVersion, request.AuthoredAgainstSnapshotID, []byte(request.DeckOps), []byte(request.DiagramDecls), []byte(request.AuthoredTexts), now.UTC())
	record, err := scanResource(row)
	if err != nil {
		return MutationResult{}, err
	}
	payload := resourceView(record, nil)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusCreated, payload); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, err
	}
	return MutationResult{Payload: payload, StatusCode: http.StatusCreated, IncidentID: incidentID}, nil
}

func (s *Store) UpdateDraft(ctx context.Context, incidentID uuid.UUID, compositionID uuid.UUID, actorUserID uuid.UUID, request UpdateDraftRequest, now time.Time) (MutationResult, error) {
	requestHash := hashRequest(request.Normalized)
	key := authn.RouteIdempotencyKey{
		RouteKey:    "report_compositions.update",
		ActorUserID: actorUserID,
		ScopeKey:    compositionID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if replay, ok, err := replayMutation(ctx, tx, key, requestHash); ok || err != nil {
		return replay, err
	}
	record, err := getResourceForUpdateTx(ctx, tx, incidentID, compositionID)
	if err != nil {
		return MutationResult{}, err
	}
	if record.RetiredAt != nil {
		return MutationResult{}, ErrResourceRetired
	}
	if record.DraftVersion != request.BaseDraftVersion {
		return MutationResult{}, ErrDraftVersionConflict
	}
	next := record
	if request.AuthoredAgainstPresent {
		next.AuthoredAgainstSnapshotID = request.AuthoredAgainstSnapshotID
	}
	if request.DeckOps != nil {
		next.DeckOps = cloneRaw(*request.DeckOps)
	}
	if request.DiagramDecls != nil {
		next.DiagramDecls = cloneRaw(*request.DiagramDecls)
	}
	if request.AuthoredTexts != nil {
		next.AuthoredTexts = cloneRaw(*request.AuthoredTexts)
	}
	if summary := validateDraft(next.DeckOps, next.DiagramDecls, next.AuthoredTexts, nil); !summaryValid(summary) {
		return MutationResult{}, &validationStoreError{Status: http.StatusBadRequest, Code: "invalid_request", Summary: summary}
	}
	row := tx.QueryRow(ctx, `
UPDATE report_compositions
   SET draft_version = draft_version + 1,
       authored_against_snapshot_id = $3,
       deck_ops = $4,
       diagram_decls = $5,
       authored_texts = $6,
       updated_at = $7
 WHERE incident_id = $1
   AND composition_id = $2
RETURNING composition_id, incident_id, created_by_user_id, client_txn_id, template_id, template_version,
          draft_version, authored_against_snapshot_id, deck_ops, diagram_decls, authored_texts,
          latest_composition_version, retired_at, created_at, updated_at
`, incidentID, compositionID, next.AuthoredAgainstSnapshotID, []byte(next.DeckOps), []byte(next.DiagramDecls), []byte(next.AuthoredTexts), now.UTC())
	updated, err := scanResource(row)
	if err != nil {
		return MutationResult{}, err
	}
	bound, err := releaseBoundVersionsTx(ctx, tx, compositionID)
	if err != nil {
		return MutationResult{}, err
	}
	payload := resourceView(updated, bound)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusOK, payload); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, err
	}
	return MutationResult{Payload: payload, StatusCode: http.StatusOK, IncidentID: incidentID}, nil
}

func (s *Store) RetireResource(ctx context.Context, incidentID uuid.UUID, compositionID uuid.UUID, actorUserID uuid.UUID, request DraftVersionRequest, now time.Time) (MutationResult, error) {
	requestHash := hashRequest(request.Normalized)
	key := authn.RouteIdempotencyKey{
		RouteKey:    "report_compositions.retire",
		ActorUserID: actorUserID,
		ScopeKey:    compositionID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if replay, ok, err := replayMutation(ctx, tx, key, requestHash); ok || err != nil {
		return replay, err
	}
	record, err := getResourceForUpdateTx(ctx, tx, incidentID, compositionID)
	if err != nil {
		return MutationResult{}, err
	}
	if record.RetiredAt != nil {
		return MutationResult{}, ErrResourceRetired
	}
	if record.DraftVersion != request.BaseDraftVersion {
		return MutationResult{}, ErrDraftVersionConflict
	}
	bound, err := releaseBoundVersionsTx(ctx, tx, compositionID)
	if err != nil {
		return MutationResult{}, err
	}
	if len(bound) > 0 {
		return MutationResult{}, ErrVersionBound
	}
	row := tx.QueryRow(ctx, `
UPDATE report_compositions
   SET retired_at = $3,
       updated_at = $3
 WHERE incident_id = $1
   AND composition_id = $2
RETURNING composition_id, incident_id, created_by_user_id, client_txn_id, template_id, template_version,
          draft_version, authored_against_snapshot_id, deck_ops, diagram_decls, authored_texts,
          latest_composition_version, retired_at, created_at, updated_at
`, incidentID, compositionID, now.UTC())
	retired, err := scanResource(row)
	if err != nil {
		return MutationResult{}, err
	}
	payload := resourceView(retired, nil)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusOK, payload); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, err
	}
	return MutationResult{Payload: payload, StatusCode: http.StatusOK, IncidentID: incidentID}, nil
}

func (s *Store) FreezeVersion(ctx context.Context, incidentID uuid.UUID, compositionID uuid.UUID, actorUserID uuid.UUID, request DraftVersionRequest, now time.Time) (MutationResult, error) {
	requestHash := hashRequest(request.Normalized)
	key := authn.RouteIdempotencyKey{
		RouteKey:    "report_compositions.freeze",
		ActorUserID: actorUserID,
		ScopeKey:    compositionID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if replay, ok, err := replayMutation(ctx, tx, key, requestHash); ok || err != nil {
		return replay, err
	}
	record, err := getResourceForUpdateTx(ctx, tx, incidentID, compositionID)
	if err != nil {
		return MutationResult{}, err
	}
	if record.RetiredAt != nil {
		return MutationResult{}, ErrResourceRetired
	}
	if record.DraftVersion != request.BaseDraftVersion {
		return MutationResult{}, ErrDraftVersionConflict
	}
	versionNumber := int64(1)
	if record.LatestCompositionVersion != nil {
		versionNumber = *record.LatestCompositionVersion + 1
	}
	canonicalBytes, digest, err := canonicalComposition(record, versionNumber)
	if err != nil {
		return MutationResult{}, err
	}
	row := tx.QueryRow(ctx, `
INSERT INTO report_composition_versions (
    composition_id, composition_version, composition_sha256, canonical_composition, created_by_user_id, created_at
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING composition_id, composition_version, composition_sha256, canonical_composition, created_by_user_id, created_at,
          EXISTS (
              SELECT 1
                FROM report_composition_release_bindings b
               WHERE b.composition_id = report_composition_versions.composition_id
                 AND b.composition_version = report_composition_versions.composition_version
                 AND b.composition_sha256 = report_composition_versions.composition_sha256
          ) AS release_bound
`, compositionID, versionNumber, digest, []byte(canonicalBytes), actorUserID, now.UTC())
	version, err := scanVersion(row)
	if err != nil {
		return MutationResult{}, err
	}
	if _, err := tx.Exec(ctx, `
UPDATE report_compositions
   SET latest_composition_version = $3,
       updated_at = $4
 WHERE incident_id = $1
   AND composition_id = $2
`, incidentID, compositionID, versionNumber, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	payload := versionView(version)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusCreated, payload); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, err
	}
	return MutationResult{Payload: payload, StatusCode: http.StatusCreated, IncidentID: incidentID}, nil
}

func (s *Store) CreatePreviewAttempt(ctx context.Context, incidentID uuid.UUID, compositionID uuid.UUID, actorUserID uuid.UUID, request PreviewRequest, now time.Time) (PreviewResult, error) {
	requestHash := hashRequest(request.Normalized)
	key := authn.RouteIdempotencyKey{
		RouteKey:    "report_compositions.preview",
		ActorUserID: actorUserID,
		ScopeKey:    compositionID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return PreviewResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if mutation, ok, err := replayMutation(ctx, tx, key, requestHash); ok || err != nil {
		return PreviewResult{Payload: mutation.Payload, StatusCode: mutation.StatusCode, IncidentID: mutation.IncidentID, Replayed: mutation.Replayed}, err
	}
	resource, err := getResourceForUpdateTx(ctx, tx, incidentID, compositionID)
	if err != nil {
		return PreviewResult{}, err
	}
	if resource.RetiredAt != nil {
		return PreviewResult{}, ErrResourceRetired
	}
	if resource.TemplateID != request.TemplateID || resource.TemplateVersion != request.TemplateVersion {
		return PreviewResult{}, &validationStoreError{Status: http.StatusBadRequest, Code: "invalid_request", ValidationCode: "composition_template_mismatch"}
	}
	var version *VersionRecord
	if request.SourceKind == SourceKindVersion {
		selected, err := getVersionTx(ctx, tx, compositionID, *request.CompositionVersion)
		if err != nil {
			return PreviewResult{}, err
		}
		version = &selected
	}
	sourceJSON, sourceSHA, draftVersion, compositionSHA, err := previewSource(resource, version, request)
	if err != nil {
		return PreviewResult{}, err
	}
	row := tx.QueryRow(ctx, `
INSERT INTO report_composition_preview_attempts (
    incident_id, composition_id, source_kind, draft_version, composition_version,
    preview_source_sha256, composition_sha256, preview_source_json, snapshot_id, derivation_version,
    template_id, template_version, redaction_profile_id, redaction_profile_version, redaction_profile_sha256,
    render_environment_profile_id, output_kind, output_options, recipient_partition_refs, graph_projection_refs,
    created_by_user_id, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
RETURNING preview_attempt_id
`, incidentID, compositionID, request.SourceKind, draftVersion, request.CompositionVersion, sourceSHA, compositionSHA, []byte(sourceJSON), request.SnapshotID, request.DerivationVersion, request.TemplateID, request.TemplateVersion, request.RedactionProfileID, request.RedactionProfileVersion, request.RedactionProfileSHA256, request.RenderEnvironmentProfile, request.OutputKind, []byte(request.OutputOptions), []byte(request.RecipientPartitionRefs), []byte(request.GraphProjectionRefs), actorUserID, now.UTC())
	var previewAttemptID uuid.UUID
	if err := row.Scan(&previewAttemptID); err != nil {
		return PreviewResult{}, err
	}
	payload := previewView(previewAttemptID, resource, version, request, sourceSHA, draftVersion, compositionSHA)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusAccepted, payload); err != nil {
		return PreviewResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return PreviewResult{}, err
	}
	return PreviewResult{Payload: payload, StatusCode: http.StatusAccepted, IncidentID: incidentID}, nil
}

func (s *Store) getResource(ctx context.Context, incidentID uuid.UUID, compositionID uuid.UUID) (ResourceRecord, error) {
	row := s.db.QueryRow(ctx, `
SELECT composition_id, incident_id, created_by_user_id, client_txn_id, template_id, template_version,
       draft_version, authored_against_snapshot_id, deck_ops, diagram_decls, authored_texts,
       latest_composition_version, retired_at, created_at, updated_at
  FROM report_compositions
 WHERE incident_id = $1
   AND composition_id = $2
`, incidentID, compositionID)
	return scanResource(row)
}

func getResourceForUpdateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, compositionID uuid.UUID) (ResourceRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT composition_id, incident_id, created_by_user_id, client_txn_id, template_id, template_version,
       draft_version, authored_against_snapshot_id, deck_ops, diagram_decls, authored_texts,
       latest_composition_version, retired_at, created_at, updated_at
  FROM report_compositions
 WHERE incident_id = $1
   AND composition_id = $2
 FOR UPDATE
`, incidentID, compositionID)
	return scanResource(row)
}

func (s *Store) getVersion(ctx context.Context, compositionID uuid.UUID, compositionVersion int64) (VersionRecord, error) {
	row := s.db.QueryRow(ctx, versionSelectSQL()+`
 WHERE v.composition_id = $1
   AND v.composition_version = $2
`, compositionID, compositionVersion)
	return scanVersion(row)
}

func getVersionTx(ctx context.Context, tx pgx.Tx, compositionID uuid.UUID, compositionVersion int64) (VersionRecord, error) {
	row := tx.QueryRow(ctx, versionSelectSQL()+`
 WHERE v.composition_id = $1
   AND v.composition_version = $2
`, compositionID, compositionVersion)
	return scanVersion(row)
}

func versionSelectSQL() string {
	return `
SELECT v.composition_id, v.composition_version, v.composition_sha256, v.canonical_composition,
       v.created_by_user_id, v.created_at,
       EXISTS (
           SELECT 1
             FROM report_composition_release_bindings b
            WHERE b.composition_id = v.composition_id
              AND b.composition_version = v.composition_version
              AND b.composition_sha256 = v.composition_sha256
       ) AS release_bound
  FROM report_composition_versions v`
}

func (s *Store) releaseBoundVersions(ctx context.Context, compositionID uuid.UUID) ([]int64, error) {
	rows, err := s.db.Query(ctx, `
SELECT DISTINCT composition_version
  FROM report_composition_release_bindings
 WHERE composition_id = $1
 ORDER BY composition_version ASC
`, compositionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanInt64Rows(rows)
}

func releaseBoundVersionsTx(ctx context.Context, tx pgx.Tx, compositionID uuid.UUID) ([]int64, error) {
	rows, err := tx.Query(ctx, `
SELECT DISTINCT composition_version
  FROM report_composition_release_bindings
 WHERE composition_id = $1
 ORDER BY composition_version ASC
`, compositionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanInt64Rows(rows)
}

func replayMutation(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey, requestHash []byte) (MutationResult, bool, error) {
	existing, err := lookupRouteIdempotencyTx(ctx, tx, key)
	if errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, false, nil
	}
	if err != nil {
		return MutationResult{}, false, err
	}
	if !bytes.Equal(existing.RequestHash, requestHash) {
		return MutationResult{}, true, authn.ErrClientTxnConflict
	}
	var payload map[string]any
	if err := json.Unmarshal(existing.ResponseJSON, &payload); err != nil {
		return MutationResult{}, true, err
	}
	return MutationResult{Payload: payload, StatusCode: existing.StatusCode, Replayed: true}, true, nil
}

func lookupRouteIdempotencyTx(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT route_key, scope_key, client_txn_id, actor_user_id, request_hash, status_code, response_json
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id = $2
   AND scope_key = $3
   AND client_txn_id = $4
`, key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID)
	var record authn.RouteIdempotencyRecord
	if err := row.Scan(&record.RouteKey, &record.ScopeKey, &record.ClientTxnID, &record.ActorUserID, &record.RequestHash, &record.StatusCode, &record.ResponseJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authn.RouteIdempotencyRecord{}, authn.ErrNotFound
		}
		return authn.RouteIdempotencyRecord{}, err
	}
	return record, nil
}

type validationStoreError struct {
	Status         int
	Code           string
	ValidationCode string
	Summary        map[string]any
}

func (e *validationStoreError) Error() string {
	if e.ValidationCode != "" {
		return e.ValidationCode
	}
	return "composition validation failed"
}

type scanner interface {
	Scan(dest ...any) error
}

func scanResource(row scanner) (ResourceRecord, error) {
	var record ResourceRecord
	var authored pgtype.Text
	var latest pgtype.Int8
	var retired pgtype.Timestamptz
	if err := row.Scan(
		&record.CompositionID,
		&record.IncidentID,
		&record.CreatedByUserID,
		&record.ClientTxnID,
		&record.TemplateID,
		&record.TemplateVersion,
		&record.DraftVersion,
		&authored,
		&record.DeckOps,
		&record.DiagramDecls,
		&record.AuthoredTexts,
		&latest,
		&retired,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ResourceRecord{}, ErrNotFound
		}
		return ResourceRecord{}, err
	}
	if authored.Valid {
		value := authored.String
		record.AuthoredAgainstSnapshotID = &value
	}
	if latest.Valid {
		value := latest.Int64
		record.LatestCompositionVersion = &value
	}
	if retired.Valid {
		value := retired.Time.UTC()
		record.RetiredAt = &value
	}
	record.CreatedAt = record.CreatedAt.UTC()
	record.UpdatedAt = record.UpdatedAt.UTC()
	return record, nil
}

func scanVersion(row scanner) (VersionRecord, error) {
	var version VersionRecord
	if err := row.Scan(
		&version.CompositionID,
		&version.CompositionVersion,
		&version.CompositionSHA256,
		&version.CanonicalComposition,
		&version.CreatedByUserID,
		&version.CreatedAt,
		&version.ReleaseBound,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return VersionRecord{}, ErrNotFound
		}
		return VersionRecord{}, err
	}
	version.CreatedAt = version.CreatedAt.UTC()
	return version, nil
}

func scanInt64Rows(rows pgx.Rows) ([]int64, error) {
	out := []int64{}
	for rows.Next() {
		var value int64
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, rows.Err()
}

func hashRequest(data []byte) []byte {
	sum := sha256.Sum256(data)
	out := make([]byte, len(sum))
	copy(out, sum[:])
	return out
}

func resourceView(record ResourceRecord, bound []int64) map[string]any {
	releaseBoundVersions := make([]string, 0, len(bound))
	for _, version := range bound {
		releaseBoundVersions = append(releaseBoundVersions, formatCompositionVersion(version))
	}
	return map[string]any{
		"composition_id":               record.CompositionID.String(),
		"incident_id":                  record.IncidentID.String(),
		"template_id":                  record.TemplateID,
		"template_version":             record.TemplateVersion,
		"draft_version":                record.DraftVersion,
		"authored_against_snapshot_id": optionalStringForJSON(record.AuthoredAgainstSnapshotID),
		"deck_ops":                     rawJSONValue(record.DeckOps),
		"diagram_decls":                rawJSONValue(record.DiagramDecls),
		"authored_texts":               rawJSONValue(record.AuthoredTexts),
		"latest_composition_version":   optionalVersionForJSON(record.LatestCompositionVersion),
		"release_bound_versions":       releaseBoundVersions,
		"retired_at":                   optionalTimeForJSON(record.RetiredAt),
	}
}

func versionView(version VersionRecord) map[string]any {
	return map[string]any{
		"composition_id":        version.CompositionID.String(),
		"composition_version":   formatCompositionVersion(version.CompositionVersion),
		"composition_sha256":    version.CompositionSHA256,
		"canonical_composition": rawJSONValue(version.CanonicalComposition),
		"created_at":            version.CreatedAt.UTC(),
		"release_bound":         version.ReleaseBound,
	}
}

func previewView(previewAttemptID uuid.UUID, resource ResourceRecord, version *VersionRecord, request PreviewRequest, previewSHA string, draftVersion *int64, compositionSHA *string) map[string]any {
	var compositionVersion any
	if version != nil {
		compositionVersion = formatCompositionVersion(version.CompositionVersion)
	}
	return map[string]any{
		"preview_attempt_id":            previewAttemptID.String(),
		"render_attempt_id":             nil,
		"incident_id":                   resource.IncidentID.String(),
		"composition_id":                resource.CompositionID.String(),
		"source_kind":                   request.SourceKind,
		"draft_version":                 draftVersion,
		"composition_version":           compositionVersion,
		"preview_source_sha256":         previewSHA,
		"composition_sha256":            optionalStringForJSON(compositionSHA),
		"snapshot_id":                   request.SnapshotID,
		"derivation_version":            request.DerivationVersion,
		"template_id":                   request.TemplateID,
		"template_version":              request.TemplateVersion,
		"redaction_profile_id":          request.RedactionProfileID,
		"redaction_profile_version":     request.RedactionProfileVersion,
		"redaction_profile_sha256":      request.RedactionProfileSHA256,
		"render_environment_profile_id": request.RenderEnvironmentProfile,
		"output_kind":                   request.OutputKind,
		"output_options":                rawJSONValue(request.OutputOptions),
		"recipient_partition_refs":      rawJSONValue(request.RecipientPartitionRefs),
		"graph_projection_refs":         rawJSONValue(request.GraphProjectionRefs),
		"release_scope":                 ReleaseScopeInternalDraft,
	}
}

func optionalVersionForJSON(value *int64) any {
	if value == nil {
		return nil
	}
	return formatCompositionVersion(*value)
}

func optionalTimeForJSON(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC()
}
