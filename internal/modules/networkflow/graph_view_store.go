package networkflow

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	GraphViewDeclarationStateActive  = "active"
	GraphViewDeclarationStateRetired = "retired"
)

var (
	graphViewIDPattern             = regexp.MustCompile(`^nfgv_[a-f0-9]{32}$`)
	graphProjectionResultIDPattern = regexp.MustCompile(`^gpres_[a-f0-9]{64}$`)
	graphViewSHA256Pattern         = regexp.MustCompile(`^[a-f0-9]{64}$`)

	ErrGraphViewDeclarationNotFound  = errors.New("network flow graph view declaration not found")
	ErrGraphViewDeclarationInvalid   = errors.New("network flow graph view declaration invalid")
	ErrGraphViewDeclarationNotActive = errors.New("network flow graph view declaration not active")
	ErrGraphViewVersionConflict      = errors.New("network flow graph view version conflict")
	ErrGraphViewDeclarationLimit     = errors.New("network flow graph view declaration limit exceeded")
	ErrGraphViewPublicationStale     = errors.New("network flow graph view publication stale")
)

type GraphViewDeclarationCounts struct {
	Active   int64
	Retained int64
}

type GraphViewVersionConflictError struct {
	Current int64
	Base    int64
}

func (err *GraphViewVersionConflictError) Error() string { return ErrGraphViewVersionConflict.Error() }
func (err *GraphViewVersionConflictError) Unwrap() error { return ErrGraphViewVersionConflict }

type GraphViewSelectedResultBinding struct {
	ProjectionResultID            string
	SourceSnapshotID              string
	ProjectionSchemaID            string
	ProjectionVersion             string
	NormalizedConfigurationSHA256 string
	NormalizedSourceSHA256        string
	CanonicalOutputSHA256         string
}

type GraphViewDeclaration struct {
	GraphViewID               string
	IncidentID                uuid.UUID
	DisplayName               string
	NormalizedDisplayName     string
	DeclarationState          string
	SemanticQueryJSON         json.RawMessage
	SemanticQuerySHA256       string
	DesiredSourceSnapshotID   string
	SelectedResult            *GraphViewSelectedResultBinding
	GraphViewVersion          int64
	MaterializationGeneration int64
	CreatedByUserID           uuid.UUID
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	RetiredAt                 *time.Time
	LatestJobID               *uuid.UUID
	LastFailureCode           *string
	LastFailedAt              *time.Time
}

func (s *Store) InsertGraphViewDeclarationTx(ctx context.Context, tx pgx.Tx, declaration GraphViewDeclaration) error {
	if s == nil || tx == nil || !validGraphViewDeclaration(declaration) {
		return ErrGraphViewDeclarationInvalid
	}
	selected := declaration.SelectedResult
	var selectedProjectionResultID, selectedSourceSnapshotID, selectedProjectionSchemaID any
	var selectedProjectionVersion, selectedConfigurationSHA256, selectedSourceSHA256, selectedOutputSHA256 any
	if selected != nil {
		selectedProjectionResultID = selected.ProjectionResultID
		selectedSourceSnapshotID = selected.SourceSnapshotID
		selectedProjectionSchemaID = selected.ProjectionSchemaID
		selectedProjectionVersion = selected.ProjectionVersion
		selectedConfigurationSHA256 = selected.NormalizedConfigurationSHA256
		selectedSourceSHA256 = selected.NormalizedSourceSHA256
		selectedOutputSHA256 = selected.CanonicalOutputSHA256
	}
	_, err := tx.Exec(ctx, `
INSERT INTO network_flow_graph_views (
    graph_view_id, incident_id, display_name, normalized_display_name,
    declaration_state, semantic_query_json, semantic_query_sha256,
    desired_source_snapshot_id, selected_projection_result_id,
    selected_source_snapshot_id, selected_projection_schema_id,
    selected_projection_version, selected_normalized_configuration_sha256,
    selected_normalized_source_sha256, selected_canonical_output_sha256,
    graph_view_version, materialization_generation, created_by_user_id,
    created_at, updated_at, retired_at, latest_job_id,
    last_failure_code, last_failed_at
) VALUES (
    $1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10, $11, $12,
    $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
)
`, declaration.GraphViewID, declaration.IncidentID, declaration.DisplayName,
		declaration.NormalizedDisplayName, declaration.DeclarationState, []byte(declaration.SemanticQueryJSON),
		declaration.SemanticQuerySHA256, declaration.DesiredSourceSnapshotID,
		selectedProjectionResultID, selectedSourceSnapshotID, selectedProjectionSchemaID,
		selectedProjectionVersion, selectedConfigurationSHA256, selectedSourceSHA256, selectedOutputSHA256,
		declaration.GraphViewVersion, declaration.MaterializationGeneration, declaration.CreatedByUserID,
		declaration.CreatedAt.UTC(), declaration.UpdatedAt.UTC(), declaration.RetiredAt, declaration.LatestJobID,
		declaration.LastFailureCode, declaration.LastFailedAt)
	if err != nil {
		return fmt.Errorf("insert Network Flow graph view declaration: %w", err)
	}
	return nil
}

func (s *Store) GetGraphViewDeclaration(ctx context.Context, incidentID uuid.UUID, graphViewID string) (GraphViewDeclaration, error) {
	if s == nil || s.pool == nil || incidentID == uuid.Nil || !graphViewIDPattern.MatchString(graphViewID) {
		return GraphViewDeclaration{}, ErrGraphViewDeclarationInvalid
	}
	return readGraphViewDeclaration(ctx, s.pool, incidentID, graphViewID, false)
}

func (s *Store) GetGraphViewDeclarationTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, graphViewID string, lock bool) (GraphViewDeclaration, error) {
	if s == nil || tx == nil || incidentID == uuid.Nil || !graphViewIDPattern.MatchString(graphViewID) {
		return GraphViewDeclaration{}, ErrGraphViewDeclarationInvalid
	}
	return readGraphViewDeclaration(ctx, tx, incidentID, graphViewID, lock)
}

func (s *Store) ListActiveGraphViewDeclarations(ctx context.Context, incidentID uuid.UUID) ([]GraphViewDeclaration, error) {
	if s == nil || s.pool == nil || incidentID == uuid.Nil {
		return nil, ErrGraphViewDeclarationInvalid
	}
	rows, err := s.pool.Query(ctx, graphViewDeclarationSelect+`
 WHERE incident_id = $1
   AND declaration_state = 'active'
 ORDER BY normalized_display_name ASC, graph_view_id ASC
`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("list Network Flow graph view declarations: %w", err)
	}
	defer rows.Close()
	declarations := make([]GraphViewDeclaration, 0)
	for rows.Next() {
		declaration, err := scanGraphViewDeclaration(rows)
		if err != nil {
			return nil, err
		}
		declarations = append(declarations, declaration)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list Network Flow graph view declarations: %w", err)
	}
	return declarations, nil
}

func (s *Store) CountGraphViewDeclarationsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, maximum int64) (GraphViewDeclarationCounts, error) {
	if s == nil || tx == nil || incidentID == uuid.Nil || maximum < 1 {
		return GraphViewDeclarationCounts{}, ErrGraphViewDeclarationInvalid
	}
	var counts GraphViewDeclarationCounts
	err := tx.QueryRow(ctx, `
SELECT COUNT(*) FILTER (WHERE declaration_state = 'active'), COUNT(*)
  FROM (
        SELECT declaration_state
          FROM network_flow_graph_views
         WHERE incident_id = $1
         LIMIT $2
       ) bounded_graph_views
`, incidentID, maximum+1).Scan(&counts.Active, &counts.Retained)
	if err != nil {
		return GraphViewDeclarationCounts{}, fmt.Errorf("count Network Flow graph view declarations: %w", err)
	}
	return counts, nil
}

func (s *Store) SetGraphViewLatestJobTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, graphViewID string, jobID uuid.UUID) (GraphViewDeclaration, error) {
	if s == nil || tx == nil || incidentID == uuid.Nil || jobID == uuid.Nil || !graphViewIDPattern.MatchString(graphViewID) {
		return GraphViewDeclaration{}, ErrGraphViewDeclarationInvalid
	}
	row := tx.QueryRow(ctx, `
UPDATE network_flow_graph_views
   SET latest_job_id = $3
 WHERE incident_id = $1
   AND graph_view_id = $2
   AND declaration_state = 'active'
RETURNING `+graphViewDeclarationColumns, incidentID, graphViewID, jobID)
	declaration, err := scanGraphViewDeclaration(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return GraphViewDeclaration{}, ErrGraphViewDeclarationNotActive
	}
	return declaration, err
}

func (s *Store) RenameGraphViewDeclarationTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, graphViewID string, baseVersion int64, displayName, normalizedDisplayName string, now time.Time) (GraphViewDeclaration, error) {
	declaration, err := s.GetGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, true)
	if err != nil {
		return GraphViewDeclaration{}, err
	}
	if declaration.DeclarationState != GraphViewDeclarationStateActive {
		return GraphViewDeclaration{}, ErrGraphViewDeclarationNotActive
	}
	if declaration.GraphViewVersion != baseVersion {
		return GraphViewDeclaration{}, &GraphViewVersionConflictError{Current: declaration.GraphViewVersion, Base: baseVersion}
	}
	row := tx.QueryRow(ctx, `
UPDATE network_flow_graph_views
   SET display_name = $3,
       normalized_display_name = $4,
       graph_view_version = graph_view_version + 1,
       updated_at = $5
 WHERE incident_id = $1
   AND graph_view_id = $2
RETURNING `+graphViewDeclarationColumns, incidentID, graphViewID, displayName, normalizedDisplayName, now.UTC())
	return scanGraphViewDeclaration(row)
}

func (s *Store) RefreshGraphViewDeclarationTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, graphViewID string, baseVersion int64, desiredSourceSnapshotID string, now time.Time) (GraphViewDeclaration, error) {
	declaration, err := s.GetGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, true)
	if err != nil {
		return GraphViewDeclaration{}, err
	}
	if declaration.DeclarationState != GraphViewDeclarationStateActive {
		return GraphViewDeclaration{}, ErrGraphViewDeclarationNotActive
	}
	if declaration.GraphViewVersion != baseVersion {
		return GraphViewDeclaration{}, &GraphViewVersionConflictError{Current: declaration.GraphViewVersion, Base: baseVersion}
	}
	row := tx.QueryRow(ctx, `
UPDATE network_flow_graph_views
   SET desired_source_snapshot_id = $3,
       graph_view_version = graph_view_version + 1,
       materialization_generation = materialization_generation + 1,
       latest_job_id = NULL,
       last_failure_code = NULL,
       last_failed_at = NULL,
       updated_at = $4
 WHERE incident_id = $1
   AND graph_view_id = $2
RETURNING `+graphViewDeclarationColumns, incidentID, graphViewID, desiredSourceSnapshotID, now.UTC())
	return scanGraphViewDeclaration(row)
}

func (s *Store) RetireGraphViewDeclarationTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, graphViewID string, baseVersion int64, now time.Time) (GraphViewDeclaration, error) {
	declaration, err := s.GetGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, true)
	if err != nil {
		return GraphViewDeclaration{}, err
	}
	if declaration.DeclarationState != GraphViewDeclarationStateActive {
		return GraphViewDeclaration{}, ErrGraphViewDeclarationNotActive
	}
	if declaration.GraphViewVersion != baseVersion {
		return GraphViewDeclaration{}, &GraphViewVersionConflictError{Current: declaration.GraphViewVersion, Base: baseVersion}
	}
	row := tx.QueryRow(ctx, `
UPDATE network_flow_graph_views
   SET declaration_state = 'retired',
       graph_view_version = graph_view_version + 1,
       materialization_generation = materialization_generation + 1,
       selected_projection_result_id = NULL,
       selected_source_snapshot_id = NULL,
       selected_projection_schema_id = NULL,
       selected_projection_version = NULL,
       selected_normalized_configuration_sha256 = NULL,
       selected_normalized_source_sha256 = NULL,
       selected_canonical_output_sha256 = NULL,
       latest_job_id = NULL,
       updated_at = $3,
       retired_at = $3
 WHERE incident_id = $1
   AND graph_view_id = $2
RETURNING `+graphViewDeclarationColumns, incidentID, graphViewID, now.UTC())
	return scanGraphViewDeclaration(row)
}

func (s *Store) PublishGraphViewResultTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, graphViewID string, generation int64, sourceSnapshotID string, jobID uuid.UUID, selected GraphViewSelectedResultBinding, now time.Time) (GraphViewDeclaration, error) {
	if s == nil || tx == nil || incidentID == uuid.Nil || jobID == uuid.Nil || generation < 1 || !validSelectedGraphViewResult(selected) {
		return GraphViewDeclaration{}, ErrGraphViewDeclarationInvalid
	}
	row := tx.QueryRow(ctx, `
UPDATE network_flow_graph_views
   SET selected_projection_result_id = $6,
       selected_source_snapshot_id = $7,
       selected_projection_schema_id = $8,
       selected_projection_version = $9,
       selected_normalized_configuration_sha256 = $10,
       selected_normalized_source_sha256 = $11,
       selected_canonical_output_sha256 = $12,
       last_failure_code = NULL,
       last_failed_at = NULL,
       updated_at = $13
 WHERE incident_id = $1
   AND graph_view_id = $2
   AND declaration_state = 'active'
   AND materialization_generation = $3
   AND desired_source_snapshot_id = $4
   AND latest_job_id = $5
RETURNING `+graphViewDeclarationColumns, incidentID, graphViewID, generation, sourceSnapshotID, jobID,
		selected.ProjectionResultID, selected.SourceSnapshotID, selected.ProjectionSchemaID, selected.ProjectionVersion,
		selected.NormalizedConfigurationSHA256, selected.NormalizedSourceSHA256, selected.CanonicalOutputSHA256, now.UTC())
	declaration, err := scanGraphViewDeclaration(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return GraphViewDeclaration{}, ErrGraphViewPublicationStale
	}
	return declaration, err
}

func (s *Store) RecordGraphViewMaterializationFailureTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, graphViewID string, generation int64, jobID uuid.UUID, failureCode string, now time.Time) error {
	if s == nil || tx == nil || incidentID == uuid.Nil || jobID == uuid.Nil || generation < 1 || !strings.HasPrefix(failureCode, "network_flow_") {
		return ErrGraphViewDeclarationInvalid
	}
	_, err := tx.Exec(ctx, `
UPDATE network_flow_graph_views
   SET last_failure_code = $6,
       last_failed_at = $5,
       updated_at = $5
 WHERE incident_id = $1
   AND graph_view_id = $2
   AND declaration_state = 'active'
   AND materialization_generation = $3
   AND latest_job_id = $4
`, incidentID, graphViewID, generation, jobID, now.UTC(), failureCode)
	return err
}

func (s *Store) InvalidateGraphViewsForTableTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, tableID string, now time.Time) error {
	if s == nil || tx == nil || incidentID == uuid.Nil || tableID == "" {
		return ErrGraphViewDeclarationInvalid
	}
	_, err := tx.Exec(ctx, `
UPDATE network_flow_graph_views
   SET graph_view_version = graph_view_version + 1,
       materialization_generation = materialization_generation + 1,
       selected_projection_result_id = NULL,
       selected_source_snapshot_id = NULL,
       selected_projection_schema_id = NULL,
       selected_projection_version = NULL,
       selected_normalized_configuration_sha256 = NULL,
       selected_normalized_source_sha256 = NULL,
       selected_canonical_output_sha256 = NULL,
       latest_job_id = NULL,
       last_failure_code = 'network_flow_source_table_deleted',
       last_failed_at = $3,
       updated_at = $3
 WHERE incident_id = $1
   AND declaration_state = 'active'
   AND semantic_query_json -> 'selected_table_ids' @> jsonb_build_array($2::text)
`, incidentID, tableID, now.UTC())
	if err != nil {
		return fmt.Errorf("invalidate Network Flow graph views for source table: %w", err)
	}
	return nil
}

func NewGraphViewID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate Network Flow graph view id: %w", err)
	}
	return "nfgv_" + hex.EncodeToString(value[:]), nil
}

func GraphViewSemanticQuerySHA256(value json.RawMessage) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func validSelectedGraphViewResult(selected GraphViewSelectedResultBinding) bool {
	return graphProjectionResultIDPattern.MatchString(selected.ProjectionResultID) && strings.TrimSpace(selected.SourceSnapshotID) != "" &&
		selected.ProjectionSchemaID == "graph_projection.v2" && strings.TrimSpace(selected.ProjectionVersion) != "" &&
		graphViewSHA256Pattern.MatchString(selected.NormalizedConfigurationSHA256) && graphViewSHA256Pattern.MatchString(selected.NormalizedSourceSHA256) &&
		graphViewSHA256Pattern.MatchString(selected.CanonicalOutputSHA256)
}

const graphViewDeclarationColumns = `graph_view_id, incident_id, display_name, normalized_display_name,
       declaration_state, semantic_query_json, semantic_query_sha256,
       desired_source_snapshot_id, selected_projection_result_id,
       selected_source_snapshot_id, selected_projection_schema_id,
       selected_projection_version, selected_normalized_configuration_sha256,
       selected_normalized_source_sha256, selected_canonical_output_sha256,
       graph_view_version, materialization_generation, created_by_user_id,
       created_at, updated_at, retired_at, latest_job_id,
       last_failure_code, last_failed_at`

const graphViewDeclarationSelect = `SELECT ` + graphViewDeclarationColumns + `
  FROM network_flow_graph_views`

type graphViewRow interface {
	Scan(...any) error
}

func readGraphViewDeclaration(ctx context.Context, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, incidentID uuid.UUID, graphViewID string, lock bool) (GraphViewDeclaration, error) {
	query := graphViewDeclarationSelect + ` WHERE incident_id = $1 AND graph_view_id = $2`
	if lock {
		query += ` FOR UPDATE`
	}
	declaration, err := scanGraphViewDeclaration(db.QueryRow(ctx, query, incidentID, graphViewID))
	if errors.Is(err, pgx.ErrNoRows) {
		return GraphViewDeclaration{}, ErrGraphViewDeclarationNotFound
	}
	return declaration, err
}

func scanGraphViewDeclaration(row graphViewRow) (GraphViewDeclaration, error) {
	var declaration GraphViewDeclaration
	var selectedProjectionResultID, selectedSourceSnapshotID, selectedProjectionSchemaID pgtype.Text
	var selectedProjectionVersion, selectedConfigurationSHA256, selectedSourceSHA256, selectedOutputSHA256 pgtype.Text
	var retiredAt, lastFailedAt pgtype.Timestamptz
	var latestJobID pgtype.UUID
	var lastFailureCode pgtype.Text
	err := row.Scan(
		&declaration.GraphViewID, &declaration.IncidentID, &declaration.DisplayName,
		&declaration.NormalizedDisplayName, &declaration.DeclarationState, &declaration.SemanticQueryJSON,
		&declaration.SemanticQuerySHA256, &declaration.DesiredSourceSnapshotID,
		&selectedProjectionResultID, &selectedSourceSnapshotID, &selectedProjectionSchemaID,
		&selectedProjectionVersion, &selectedConfigurationSHA256, &selectedSourceSHA256, &selectedOutputSHA256,
		&declaration.GraphViewVersion, &declaration.MaterializationGeneration, &declaration.CreatedByUserID,
		&declaration.CreatedAt, &declaration.UpdatedAt, &retiredAt, &latestJobID,
		&lastFailureCode, &lastFailedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GraphViewDeclaration{}, err
		}
		return GraphViewDeclaration{}, fmt.Errorf("scan Network Flow graph view declaration: %w", err)
	}
	if selectedProjectionResultID.Valid {
		declaration.SelectedResult = &GraphViewSelectedResultBinding{
			ProjectionResultID:            selectedProjectionResultID.String,
			SourceSnapshotID:              selectedSourceSnapshotID.String,
			ProjectionSchemaID:            selectedProjectionSchemaID.String,
			ProjectionVersion:             selectedProjectionVersion.String,
			NormalizedConfigurationSHA256: selectedConfigurationSHA256.String,
			NormalizedSourceSHA256:        selectedSourceSHA256.String,
			CanonicalOutputSHA256:         selectedOutputSHA256.String,
		}
	}
	if retiredAt.Valid {
		value := retiredAt.Time
		declaration.RetiredAt = &value
	}
	if latestJobID.Valid {
		value := uuid.UUID(latestJobID.Bytes)
		declaration.LatestJobID = &value
	}
	if lastFailureCode.Valid {
		value := lastFailureCode.String
		declaration.LastFailureCode = &value
	}
	if lastFailedAt.Valid {
		value := lastFailedAt.Time
		declaration.LastFailedAt = &value
	}
	return declaration, nil
}

func validGraphViewDeclaration(declaration GraphViewDeclaration) bool {
	if !graphViewIDPattern.MatchString(declaration.GraphViewID) || declaration.IncidentID == uuid.Nil ||
		declaration.CreatedByUserID == uuid.Nil || declaration.GraphViewVersion < 1 || declaration.MaterializationGeneration < 1 ||
		declaration.CreatedAt.IsZero() || declaration.UpdatedAt.Before(declaration.CreatedAt) ||
		utf8.RuneCountInString(declaration.DisplayName) < 1 || utf8.RuneCountInString(declaration.DisplayName) > 64 ||
		utf8.RuneCountInString(declaration.NormalizedDisplayName) < 1 || utf8.RuneCountInString(declaration.NormalizedDisplayName) > 64 ||
		strings.ContainsAny(declaration.DisplayName, "\x00\n\r") || strings.ContainsAny(declaration.NormalizedDisplayName, "\x00\n\r") ||
		!validJSONObjectBytes(declaration.SemanticQueryJSON) || !graphViewSHA256Pattern.MatchString(declaration.SemanticQuerySHA256) ||
		strings.TrimSpace(declaration.DesiredSourceSnapshotID) == "" {
		return false
	}
	if declaration.DeclarationState != GraphViewDeclarationStateActive && declaration.DeclarationState != GraphViewDeclarationStateRetired {
		return false
	}
	if (declaration.DeclarationState == GraphViewDeclarationStateActive) != (declaration.RetiredAt == nil) {
		return false
	}
	if declaration.RetiredAt != nil && declaration.RetiredAt.Before(declaration.CreatedAt) {
		return false
	}
	if (declaration.LastFailureCode == nil) != (declaration.LastFailedAt == nil) {
		return false
	}
	if declaration.LastFailureCode != nil && !strings.HasPrefix(*declaration.LastFailureCode, "network_flow_") {
		return false
	}
	if declaration.SelectedResult != nil {
		selected := declaration.SelectedResult
		if !graphProjectionResultIDPattern.MatchString(selected.ProjectionResultID) || strings.TrimSpace(selected.SourceSnapshotID) == "" ||
			selected.ProjectionSchemaID != "graph_projection.v2" || strings.TrimSpace(selected.ProjectionVersion) == "" ||
			!graphViewSHA256Pattern.MatchString(selected.NormalizedConfigurationSHA256) ||
			!graphViewSHA256Pattern.MatchString(selected.NormalizedSourceSHA256) ||
			!graphViewSHA256Pattern.MatchString(selected.CanonicalOutputSHA256) {
			return false
		}
	}
	return true
}

func validJSONObjectBytes(value json.RawMessage) bool {
	trimmed := bytes.TrimSpace(value)
	return len(trimmed) >= 2 && trimmed[0] == '{' && trimmed[len(trimmed)-1] == '}' && json.Valid(trimmed)
}
