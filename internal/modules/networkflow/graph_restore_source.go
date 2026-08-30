package networkflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	graphRestoreSourceRegistrationID = "network_flow_activity.graph_views.v1"
	graphRestoreAuthoritativeFamily  = "network_flow_activity.graph_views"
	graphRestoreEnumeratorBindingID  = "network_flow_activity.graph_view_restore_enumerator_v2"
)

// NewGraphRestoreSourceRegistration constructs Network Flow's source-owned
// Recovery contribution. Graph sees only deterministic v2 semantic inputs and
// the exact immutable result bindings selected by authoritative declarations.
func NewGraphRestoreSourceRegistration(db postgres.DB) (graphrestore.RestoreSourceRegistration, error) {
	if db == nil {
		return graphrestore.RestoreSourceRegistration{}, fmt.Errorf("network flow graph restore source requires PostgreSQL")
	}
	limits := DefaultEffectiveLimits()
	store := NewStore(db, limits)
	return graphrestore.RestoreSourceRegistration{
		Entry: graphrestore.RestoreSourceRegistryEntry{
			SourceRegistrationID:       graphRestoreSourceRegistrationID,
			SourceOwnerID:              graphSourceOwnerID,
			AuthoritativeFamilyID:      graphRestoreAuthoritativeFamily,
			EnumeratorBindingID:        graphRestoreEnumeratorBindingID,
			SemanticQuerySchemaIDs:     []string{schemaGraphSemanticQueryV2},
			ProjectionInputContractID:  graphprojection.ProjectionSchemaIDV2,
			ProjectionResultContractID: "graph_projection_result.v2",
			Status:                     "active",
		},
		Enumerate: func(ctx context.Context, _ graphrestore.RestoreSourceState, _ time.Time) ([]graphrestore.RestoreCandidate, error) {
			rows, err := db.Query(ctx, graphViewDeclarationSelect+`
 WHERE declaration_state = 'active'
   AND selected_projection_result_id IS NOT NULL
 ORDER BY graph_view_id ASC
`)
			if err != nil {
				return nil, fmt.Errorf("enumerate Network Flow saved graphs for restore: %w", err)
			}
			declarations := make([]GraphViewDeclaration, 0)
			for rows.Next() {
				declaration, scanErr := scanGraphViewDeclaration(rows)
				if scanErr != nil {
					rows.Close()
					return nil, scanErr
				}
				declarations = append(declarations, declaration)
			}
			if err := rows.Err(); err != nil {
				rows.Close()
				return nil, fmt.Errorf("iterate Network Flow saved graphs for restore: %w", err)
			}
			rows.Close()

			composer := &Service{store: store, graphProjection: newGraphProjectionAdapter(), now: func() time.Time { return time.Now().UTC() }}
			candidates := make([]graphrestore.RestoreCandidate, 0, len(declarations))
			for _, declaration := range declarations {
				if declaration.SelectedResult == nil {
					return nil, fmt.Errorf("active selected Network Flow saved graph has no result binding")
				}
				semantic, apiErr := decodeGraphSemanticRequest(declaration.SemanticQueryJSON, limits)
				if apiErr != nil {
					return nil, fmt.Errorf("decode Network Flow saved graph %s for restore", declaration.GraphViewID)
				}
				composition, apiErr := composer.composeGraphSourceFromSemantic(ctx, declaration.IncidentID, semantic)
				if apiErr != nil {
					return nil, fmt.Errorf("compose Network Flow saved graph %s for restore", declaration.GraphViewID)
				}
				sourceSnapshotID := graphSourceSnapshotDigest(declaration.IncidentID, composition.SourceTables, composition.Digest)
				if sourceSnapshotID != declaration.SelectedResult.SourceSnapshotID || sourceSnapshotID != declaration.DesiredSourceSnapshotID {
					return nil, fmt.Errorf("network flow saved graph %s selected source is stale", declaration.GraphViewID)
				}
				selected := declaration.SelectedResult
				candidates = append(candidates, graphrestore.RestoreCandidate{
					CandidateID:           declaration.GraphViewID,
					GraphViewID:           declaration.GraphViewID,
					SemanticQuerySchemaID: semantic.SchemaID,
					SemanticInput:         canonicalJSON(networkFlowProjectionInput(sourceSnapshotID, composition)),
					ExpectedBinding: graphprojection.ResultBindingV2{
						ProjectionResultID: selected.ProjectionResultID, GraphViewID: declaration.GraphViewID,
						SourceOwnerID: graphSourceOwnerID, SourceSnapshotID: selected.SourceSnapshotID,
						ProjectionSchemaID: selected.ProjectionSchemaID, ProjectionVersion: selected.ProjectionVersion,
						NormalizedConfigurationSHA256: selected.NormalizedConfigurationSHA256,
						NormalizedSourceSHA256:        selected.NormalizedSourceSHA256,
						CanonicalOutputSHA256:         selected.CanonicalOutputSHA256,
					},
				})
			}
			return candidates, nil
		},
	}, nil
}

// ReconcileGraphRestoreJobsTx makes every restored nonterminal saved-graph
// materialization job reclaimable after the source runtime was quiesced for
// backup. Payload ownership and validation remain in Network Flow.
func ReconcileGraphRestoreJobsTx(ctx context.Context, tx pgx.Tx) (int, error) {
	if ctx == nil || tx == nil {
		return 0, fmt.Errorf("network flow graph restore job reconciliation requires a transaction")
	}
	type restoredGraphJob struct {
		jobID      uuid.UUID
		incidentID uuid.UUID
		payload    []byte
	}
	rows, err := tx.Query(ctx, `
SELECT job_id, incident_id, handler_payload_json
  FROM jobs
 WHERE status IN ('queued', 'running', 'cancel_requested')
   AND job_kind = $1
   AND extension_owner_profile_id = $2
 ORDER BY job_id
`, GraphViewMaterializationJobKind, ProfileID)
	if err != nil {
		return 0, fmt.Errorf("enumerate restored Network Flow graph jobs: %w", err)
	}
	restoredJobs := make([]restoredGraphJob, 0)
	for rows.Next() {
		var restored restoredGraphJob
		if err := rows.Scan(&restored.jobID, &restored.incidentID, &restored.payload); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan restored Network Flow graph job: %w", err)
		}
		restoredJobs = append(restoredJobs, restored)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, fmt.Errorf("iterate restored Network Flow graph jobs: %w", err)
	}
	rows.Close()
	for _, restored := range restoredJobs {
		var payload graphViewMaterializationPayload
		decoder := json.NewDecoder(bytes.NewReader(restored.payload))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil || !payload.valid() || payload.IncidentID != restored.incidentID {
			return 0, fmt.Errorf("restored Network Flow graph job %s has an invalid payload", restored.jobID)
		}
		if err := jobs.ReconcileRestoredNonterminalTx(ctx, tx, restored.jobID, GraphViewMaterializationJobKind); err != nil {
			return 0, err
		}
	}
	return len(restoredJobs), nil
}
