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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
)

const (
	graphViewDeclarationStateActive  = "active"
	graphViewDeclarationStateRetired = "retired"
)

var (
	graphViewIDPattern             = regexp.MustCompile(`^nfgv_[a-f0-9]{32}$`)
	graphProjectionResultIDPattern = regexp.MustCompile(`^gpres_[a-f0-9]{64}$`)
	graphViewSHA256Pattern         = regexp.MustCompile(`^[a-f0-9]{64}$`)

	errGraphViewDeclarationNotFound  = errors.New("network flow graph view declaration not found")
	errGraphViewDeclarationInvalid   = errors.New("network flow graph view declaration invalid")
	errGraphViewDeclarationNotActive = errors.New("network flow graph view declaration not active")
	errGraphViewVersionConflict      = errors.New("network flow graph view version conflict")
	errGraphViewActiveLimit          = errors.New("network flow active graph view limit exceeded")
	errGraphViewRetainedLimit        = errors.New("network flow retained graph view limit exceeded")
	errGraphViewPublicationStale     = errors.New("network flow graph view publication stale")
)

type graphViewDeclarationCounts struct {
	Active   int64
	Retained int64
}

type graphViewVersionConflictError struct {
	Current int64
	Base    int64
}

func (err *graphViewVersionConflictError) Error() string { return errGraphViewVersionConflict.Error() }
func (err *graphViewVersionConflictError) Unwrap() error { return errGraphViewVersionConflict }

type graphViewSelectedResultBinding struct {
	ProjectionResultID            string `json:"projection_result_id"`
	SourceSnapshotID              string `json:"source_snapshot_id"`
	ProjectionSchemaID            string `json:"projection_schema_id"`
	ProjectionVersion             string `json:"projection_version"`
	NormalizedConfigurationSHA256 string `json:"normalized_configuration_sha256"`
	NormalizedSourceSHA256        string `json:"normalized_source_sha256"`
	CanonicalOutputSHA256         string `json:"canonical_output_sha256"`
}

func (selected graphViewSelectedResultBinding) resultBinding(graphViewID string) graphprojection.ResultBindingV2 {
	return graphprojection.ResultBindingV2{
		ProjectionResultID: selected.ProjectionResultID, GraphViewID: graphViewID,
		SourceOwnerID: graphSourceOwnerID, SourceSnapshotID: selected.SourceSnapshotID,
		ProjectionSchemaID: selected.ProjectionSchemaID, ProjectionVersion: selected.ProjectionVersion,
		NormalizedConfigurationSHA256: selected.NormalizedConfigurationSHA256,
		NormalizedSourceSHA256:        selected.NormalizedSourceSHA256, CanonicalOutputSHA256: selected.CanonicalOutputSHA256,
	}
}

type graphViewDeclaration struct {
	GraphViewID               string
	IncidentID                uuid.UUID
	DisplayName               string
	NormalizedDisplayName     string
	DeclarationState          string
	RefreshIntentJSON         json.RawMessage
	SelectedSourceBindingJSON json.RawMessage
	SemanticQueryJSON         json.RawMessage
	SemanticQuerySHA256       string
	DesiredSourceSnapshotID   string
	SelectedResult            *graphViewSelectedResultBinding
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

func (s *store) InsertGraphViewDeclarationTx(ctx context.Context, tx pgx.Tx, declaration graphViewDeclaration) error {
	if s == nil {
		return errGraphViewDeclarationInvalid
	}
	if err := insertGraphViewDeclarationTx(ctx, tx, declaration); err != nil {
		return err
	}
	return s.appendGraphViewResourceIntentTx(ctx, tx, declaration, "created")
}

// Transfer retains an authored declaration without fabricating a new user
// action or materialization job. Publication remains in the caller's transaction.
func insertGraphViewDeclarationTx(ctx context.Context, tx pgx.Tx, declaration graphViewDeclaration) error {
	if tx == nil || !validGraphViewDeclaration(declaration) {
		return errGraphViewDeclarationInvalid
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
    last_failure_code, last_failed_at, refresh_intent_json, selected_source_binding_json
) VALUES (
    $1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10, $11, $12,
    $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
)
`, declaration.GraphViewID, declaration.IncidentID, declaration.DisplayName,
		declaration.NormalizedDisplayName, declaration.DeclarationState, []byte(declaration.SemanticQueryJSON),
		declaration.SemanticQuerySHA256, declaration.DesiredSourceSnapshotID,
		selectedProjectionResultID, selectedSourceSnapshotID, selectedProjectionSchemaID,
		selectedProjectionVersion, selectedConfigurationSHA256, selectedSourceSHA256, selectedOutputSHA256,
		declaration.GraphViewVersion, declaration.MaterializationGeneration, declaration.CreatedByUserID,
		declaration.CreatedAt.UTC(), declaration.UpdatedAt.UTC(), declaration.RetiredAt, declaration.LatestJobID,
		declaration.LastFailureCode, declaration.LastFailedAt, declaration.RefreshIntentJSON, declaration.SelectedSourceBindingJSON)
	if err != nil {
		return fmt.Errorf("insert Network Flow graph view declaration: %w", err)
	}
	return nil
}

func (s *store) GetGraphViewDeclaration(ctx context.Context, incidentID uuid.UUID, graphViewID string) (graphViewDeclaration, error) {
	if s == nil || s.pool == nil || incidentID == uuid.Nil || !graphViewIDPattern.MatchString(graphViewID) {
		return graphViewDeclaration{}, errGraphViewDeclarationInvalid
	}
	value, err := readGraphViewDeclaration(ctx, s.pool, incidentID, graphViewID, false)
	if err != nil {
		return value, err
	}
	return s.resolveGraphAuthor(ctx, s.pool, value)
}

func (s *store) GetGraphViewDeclarationTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, graphViewID string, lock bool) (graphViewDeclaration, error) {
	if s == nil || tx == nil || incidentID == uuid.Nil || !graphViewIDPattern.MatchString(graphViewID) {
		return graphViewDeclaration{}, errGraphViewDeclarationInvalid
	}
	value, err := readGraphViewDeclaration(ctx, tx, incidentID, graphViewID, lock)
	if err != nil {
		return value, err
	}
	return s.resolveGraphAuthor(ctx, tx, value)
}

func (s *store) ListActiveGraphViewDeclarations(ctx context.Context, incidentID uuid.UUID) ([]graphViewDeclaration, error) {
	if s == nil || s.pool == nil || incidentID == uuid.Nil {
		return nil, errGraphViewDeclarationInvalid
	}
	rows, err := s.pool.Query(ctx, graphViewDeclarationSelect+`
 WHERE incident_id = $1
   AND declaration_state = 'active'
 ORDER BY display_name COLLATE "C" ASC, graph_view_id ASC
`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("list Network Flow graph view declarations: %w", err)
	}
	defer rows.Close()
	declarations := make([]graphViewDeclaration, 0)
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
	rows.Close()
	err = resolveHistoricalAuthors(ctx, s, s.pool, incidentID, "network_flow_graph_views", declarations, graphAuthor)
	return declarations, err
}

func (s *store) CountGraphViewDeclarationsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, maximum int64) (graphViewDeclarationCounts, error) {
	if s == nil || tx == nil || incidentID == uuid.Nil || maximum < 1 {
		return graphViewDeclarationCounts{}, errGraphViewDeclarationInvalid
	}
	var counts graphViewDeclarationCounts
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
		return graphViewDeclarationCounts{}, fmt.Errorf("count Network Flow graph view declarations: %w", err)
	}
	return counts, nil
}

func (s *store) SetGraphViewLatestJobTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, graphViewID string, jobID uuid.UUID) (graphViewDeclaration, error) {
	if s == nil || tx == nil || incidentID == uuid.Nil || jobID == uuid.Nil || !graphViewIDPattern.MatchString(graphViewID) {
		return graphViewDeclaration{}, errGraphViewDeclarationInvalid
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
		return graphViewDeclaration{}, errGraphViewDeclarationNotActive
	}
	if err != nil {
		return declaration, err
	}
	return s.resolveGraphAuthor(ctx, tx, declaration)
}

func (s *store) RenameGraphViewDeclarationTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, graphViewID string, baseVersion int64, displayName, normalizedDisplayName string, now time.Time) (graphViewDeclaration, error) {
	declaration, err := s.GetGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, true)
	if err != nil {
		return graphViewDeclaration{}, err
	}
	if declaration.DeclarationState != graphViewDeclarationStateActive {
		return graphViewDeclaration{}, errGraphViewDeclarationNotActive
	}
	if declaration.GraphViewVersion != baseVersion {
		return graphViewDeclaration{}, &graphViewVersionConflictError{Current: declaration.GraphViewVersion, Base: baseVersion}
	}
	if normalized, err := normalizeGraphViewDisplayName(displayName); err != nil || normalized != displayName {
		return graphViewDeclaration{}, errGraphViewDeclarationInvalid
	}
	if declaration.DisplayName == displayName {
		return declaration, nil
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
	changed, err := scanGraphViewDeclaration(row)
	if err == nil {
		err = s.appendGraphViewResourceIntentTx(ctx, tx, changed, "renamed")
	}
	if err != nil {
		return changed, err
	}
	return s.resolveGraphAuthor(ctx, tx, changed)
}

func (s *store) RefreshGraphViewDeclarationTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, graphViewID string, baseVersion int64, desiredSourceSnapshotID string, semanticJSON json.RawMessage, now time.Time) (graphViewDeclaration, error) {
	declaration, err := s.GetGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, true)
	if err != nil {
		return graphViewDeclaration{}, err
	}
	if declaration.DeclarationState != graphViewDeclarationStateActive {
		return graphViewDeclaration{}, errGraphViewDeclarationNotActive
	}
	if declaration.GraphViewVersion != baseVersion {
		return graphViewDeclaration{}, &graphViewVersionConflictError{Current: declaration.GraphViewVersion, Base: baseVersion}
	}
	row := tx.QueryRow(ctx, `
UPDATE network_flow_graph_views
   SET desired_source_snapshot_id = $3,
       semantic_query_json = $5, semantic_query_sha256 = $6,
       graph_view_version = graph_view_version + 1,
       materialization_generation = materialization_generation + 1,
       latest_job_id = NULL,
       last_failure_code = NULL,
       last_failed_at = NULL,
       updated_at = $4
 WHERE incident_id = $1
   AND graph_view_id = $2
RETURNING `+graphViewDeclarationColumns, incidentID, graphViewID, desiredSourceSnapshotID, now.UTC(), semanticJSON, graphViewSemanticQuerySHA256(semanticJSON))
	changed, err := scanGraphViewDeclaration(row)
	if err == nil {
		err = s.appendGraphViewResourceIntentTx(ctx, tx, changed, "refresh_requested")
	}
	if err != nil {
		return changed, err
	}
	return s.resolveGraphAuthor(ctx, tx, changed)
}

func (s *store) RetireGraphViewDeclarationTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, graphViewID string, baseVersion int64, now time.Time) (graphViewDeclaration, error) {
	declaration, err := s.GetGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, true)
	if err != nil {
		return graphViewDeclaration{}, err
	}
	if declaration.DeclarationState != graphViewDeclarationStateActive {
		return graphViewDeclaration{}, errGraphViewDeclarationNotActive
	}
	if declaration.GraphViewVersion != baseVersion {
		return graphViewDeclaration{}, &graphViewVersionConflictError{Current: declaration.GraphViewVersion, Base: baseVersion}
	}
	row := tx.QueryRow(ctx, `
UPDATE network_flow_graph_views
   SET declaration_state = 'retired',
       graph_view_version = graph_view_version + 1,
       materialization_generation = materialization_generation + 1,
       latest_job_id = NULL,
       updated_at = $3,
       retired_at = $3
 WHERE incident_id = $1
   AND graph_view_id = $2
RETURNING `+graphViewDeclarationColumns, incidentID, graphViewID, now.UTC())
	changed, err := scanGraphViewDeclaration(row)
	if err == nil {
		err = s.appendGraphViewResourceIntentTx(ctx, tx, changed, "retired")
	}
	if err != nil {
		return changed, err
	}
	return s.resolveGraphAuthor(ctx, tx, changed)
}

func (s *store) PublishGraphViewResultTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, graphViewID string, generation int64, sourceSnapshotID string, jobID uuid.UUID, selected graphViewSelectedResultBinding, sourceBinding json.RawMessage, now time.Time) (graphViewDeclaration, error) {
	if s == nil || tx == nil || incidentID == uuid.Nil || jobID == uuid.Nil || generation < 1 || !validSelectedGraphViewResult(selected) {
		return graphViewDeclaration{}, errGraphViewDeclarationInvalid
	}
	if _, _, err := decodeSelectedSource(sourceBinding, &selected, s.limits); err != nil {
		return graphViewDeclaration{}, err
	}
	reader, err := postgresresult.NewReader(tx)
	if err != nil {
		return graphViewDeclaration{}, err
	}
	storedBinding, err := reader.LockResultEnvelope(ctx, selected.ProjectionResultID)
	if err != nil {
		return graphViewDeclaration{}, err
	}
	wantBinding := graphprojection.ResultBindingV2{
		ProjectionResultID:            selected.ProjectionResultID,
		GraphViewID:                   graphViewID,
		SourceOwnerID:                 ProfileID,
		SourceSnapshotID:              selected.SourceSnapshotID,
		ProjectionSchemaID:            selected.ProjectionSchemaID,
		ProjectionVersion:             selected.ProjectionVersion,
		NormalizedConfigurationSHA256: selected.NormalizedConfigurationSHA256,
		NormalizedSourceSHA256:        selected.NormalizedSourceSHA256,
		CanonicalOutputSHA256:         selected.CanonicalOutputSHA256,
	}
	if storedBinding != wantBinding {
		return graphViewDeclaration{}, graphprojection.ErrResultV2BindingMismatch
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
       selected_source_binding_json = $14,
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
		selected.NormalizedConfigurationSHA256, selected.NormalizedSourceSHA256, selected.CanonicalOutputSHA256, now.UTC(), sourceBinding)
	declaration, err := scanGraphViewDeclaration(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return graphViewDeclaration{}, errGraphViewPublicationStale
	}
	if err == nil {
		err = s.appendGraphViewResourceIntentTx(ctx, tx, declaration, "materialized")
	}
	if err != nil {
		return declaration, err
	}
	return s.resolveGraphAuthor(ctx, tx, declaration)
}

// LockGraphViewDeclarationsSelectingResultTx checks Network Flow's
// authoritative selected bindings while locking every matching declaration.
// The caller must already hold the immutable result-envelope lock.
func (s *store) LockGraphViewDeclarationsSelectingResultTx(ctx context.Context, tx pgx.Tx, projectionResultID string) (bool, error) {
	if s == nil || tx == nil || !graphProjectionResultIDPattern.MatchString(projectionResultID) {
		return false, errGraphViewDeclarationInvalid
	}
	rows, err := tx.Query(ctx, `
SELECT graph_view_id
  FROM network_flow_graph_views
 WHERE selected_projection_result_id = $1
 ORDER BY graph_view_id
 FOR UPDATE
`, projectionResultID)
	if err != nil {
		return false, fmt.Errorf("lock Network Flow selected graph declarations: %w", err)
	}
	defer rows.Close()
	selected := false
	for rows.Next() {
		var graphViewID string
		if err := rows.Scan(&graphViewID); err != nil {
			return false, fmt.Errorf("scan Network Flow selected graph declaration: %w", err)
		}
		selected = true
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate Network Flow selected graph declarations: %w", err)
	}
	return selected, nil
}

func (s *store) RecordGraphViewMaterializationFailureTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, graphViewID string, generation int64, jobID uuid.UUID, failureCode string, now time.Time) error {
	if s == nil || tx == nil || incidentID == uuid.Nil || jobID == uuid.Nil || generation < 1 || !validGraphViewFailureCode(failureCode) {
		return errGraphViewDeclarationInvalid
	}
	row := tx.QueryRow(ctx, `
UPDATE network_flow_graph_views
   SET last_failure_code = $6,
       last_failed_at = $5,
       updated_at = $5
 WHERE incident_id = $1
   AND graph_view_id = $2
   AND declaration_state = 'active'
   AND materialization_generation = $3
   AND latest_job_id = $4
RETURNING `+graphViewDeclarationColumns, incidentID, graphViewID, generation, jobID, now.UTC(), failureCode)
	graph, err := scanGraphViewDeclaration(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.appendGraphViewResourceIntentTx(ctx, tx, graph, "materialization_failed")
}

func (s *store) InvalidateGraphViewsForTableTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, tableID string, now time.Time) error {
	if s == nil || tx == nil || incidentID == uuid.Nil || tableID == "" {
		return errGraphViewDeclarationInvalid
	}
	rows, err := tx.Query(ctx, `
UPDATE network_flow_graph_views
   SET graph_view_version = graph_view_version + 1,
       materialization_generation = materialization_generation + 1,
       latest_job_id = NULL,
       last_failure_code = 'network_flow_source_table_deleted',
       last_failed_at = $3,
       updated_at = $3
 WHERE incident_id = $1
   AND declaration_state = 'active'
   AND semantic_query_json -> 'selected_table_ids' @> jsonb_build_array($2::text)
RETURNING `+graphViewDeclarationColumns, incidentID, tableID, now.UTC())
	if err != nil {
		return fmt.Errorf("invalidate Network Flow graph views for source table: %w", err)
	}
	var changed []graphViewDeclaration
	for rows.Next() {
		graph, scanErr := scanGraphViewDeclaration(rows)
		if scanErr != nil {
			rows.Close()
			return scanErr
		}
		changed = append(changed, graph)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, graph := range changed {
		if err := s.appendGraphViewResourceIntentTx(ctx, tx, graph, "source_invalidated"); err != nil {
			return err
		}
	}
	return nil
}

func newGraphViewID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate Network Flow graph view id: %w", err)
	}
	return "nfgv_" + hex.EncodeToString(value[:]), nil
}

func graphViewSemanticQuerySHA256(value json.RawMessage) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func validSelectedGraphViewResult(selected graphViewSelectedResultBinding) bool {
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
       last_failure_code, last_failed_at, refresh_intent_json, selected_source_binding_json`

const graphViewDeclarationSelect = `SELECT ` + graphViewDeclarationColumns + `
  FROM network_flow_graph_views`

type graphViewRow interface {
	Scan(...any) error
}

func readGraphViewDeclaration(ctx context.Context, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, incidentID uuid.UUID, graphViewID string, lock bool) (graphViewDeclaration, error) {
	query := graphViewDeclarationSelect + ` WHERE incident_id = $1 AND graph_view_id = $2`
	if lock {
		query += ` FOR UPDATE`
	}
	declaration, err := scanGraphViewDeclaration(db.QueryRow(ctx, query, incidentID, graphViewID))
	if errors.Is(err, pgx.ErrNoRows) {
		return graphViewDeclaration{}, errGraphViewDeclarationNotFound
	}
	return declaration, err
}

func scanGraphViewDeclaration(row graphViewRow) (graphViewDeclaration, error) {
	var declaration graphViewDeclaration
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
		&lastFailureCode, &lastFailedAt, &declaration.RefreshIntentJSON, &declaration.SelectedSourceBindingJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphViewDeclaration{}, err
		}
		return graphViewDeclaration{}, fmt.Errorf("scan Network Flow graph view declaration: %w", err)
	}
	if selectedProjectionResultID.Valid {
		declaration.SelectedResult = &graphViewSelectedResultBinding{
			ProjectionResultID:            selectedProjectionResultID.String,
			SourceSnapshotID:              selectedSourceSnapshotID.String,
			ProjectionSchemaID:            selectedProjectionSchemaID.String,
			ProjectionVersion:             selectedProjectionVersion.String,
			NormalizedConfigurationSHA256: selectedConfigurationSHA256.String,
			NormalizedSourceSHA256:        selectedSourceSHA256.String,
			CanonicalOutputSHA256:         selectedOutputSHA256.String,
		}
	}
	declaration.CreatedAt = declaration.CreatedAt.UTC()
	declaration.UpdatedAt = declaration.UpdatedAt.UTC()
	if retiredAt.Valid {
		value := retiredAt.Time.UTC()
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
		value := lastFailedAt.Time.UTC()
		declaration.LastFailedAt = &value
	}
	return declaration, nil
}

func validGraphViewDeclaration(declaration graphViewDeclaration) bool {
	if _, _, err := decodeRefreshIntent(declaration.RefreshIntentJSON, retainedSemanticLimits()); err != nil {
		return false
	}
	normalized, nameErr := normalizeGraphViewDisplayName(declaration.DisplayName)
	if nameErr != nil || normalized != declaration.DisplayName {
		return false
	}
	if declaration.LatestJobID != nil && *declaration.LatestJobID == uuid.Nil {
		return false
	}
	if !graphViewIDPattern.MatchString(declaration.GraphViewID) || declaration.IncidentID == uuid.Nil ||
		declaration.CreatedByUserID == uuid.Nil || declaration.GraphViewVersion < 1 || declaration.MaterializationGeneration < 1 ||
		declaration.CreatedAt.IsZero() || declaration.UpdatedAt.Before(declaration.CreatedAt) ||
		len(declaration.NormalizedDisplayName) == 0 ||
		strings.ContainsAny(declaration.DisplayName, "\x00\n\r") || strings.ContainsAny(declaration.NormalizedDisplayName, "\x00\n\r") ||
		!validJSONObjectBytes(declaration.RefreshIntentJSON) || !validJSONObjectBytes(declaration.SemanticQueryJSON) || !graphViewSHA256Pattern.MatchString(declaration.SemanticQuerySHA256) ||
		strings.TrimSpace(declaration.DesiredSourceSnapshotID) == "" {
		return false
	}
	if declaration.DeclarationState != graphViewDeclarationStateActive && declaration.DeclarationState != graphViewDeclarationStateRetired {
		return false
	}
	if (declaration.DeclarationState == graphViewDeclarationStateActive) != (declaration.RetiredAt == nil) {
		return false
	}
	if declaration.RetiredAt != nil && declaration.RetiredAt.Before(declaration.CreatedAt) {
		return false
	}
	if (declaration.LastFailureCode == nil) != (declaration.LastFailedAt == nil) {
		return false
	}
	if declaration.LastFailureCode != nil && !validGraphViewFailureCode(*declaration.LastFailureCode) {
		return false
	}
	if declaration.SelectedResult == nil && len(declaration.SelectedSourceBindingJSON) != 0 {
		return false
	}
	if declaration.SelectedResult != nil {
		if _, _, err := decodeSelectedSource(declaration.SelectedSourceBindingJSON, declaration.SelectedResult, retainedSemanticLimits()); err != nil {
			return false
		}
		selected := declaration.SelectedResult
		if !graphProjectionResultIDPattern.MatchString(selected.ProjectionResultID) || strings.TrimSpace(selected.SourceSnapshotID) == "" ||
			selected.ProjectionSchemaID != "graph_projection.v2" || strings.TrimSpace(selected.ProjectionVersion) == "" ||
			!graphViewSHA256Pattern.MatchString(selected.NormalizedConfigurationSHA256) ||
			!graphViewSHA256Pattern.MatchString(selected.NormalizedSourceSHA256) ||
			!graphViewSHA256Pattern.MatchString(selected.CanonicalOutputSHA256) || selected.resultBinding(declaration.GraphViewID).ValidateIdentity() != nil {
			return false
		}
	}
	return true
}

func validJSONObjectBytes(value json.RawMessage) bool {
	trimmed := bytes.TrimSpace(value)
	return len(trimmed) >= 2 && trimmed[0] == '{' && trimmed[len(trimmed)-1] == '}' && json.Valid(trimmed)
}
