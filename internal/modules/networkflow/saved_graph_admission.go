package networkflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

const SavedGraphCutoverAlgorithmID = "network_flow_activity.saved_graph_cutover_v6"

var errSavedGraphCutoverIncompatible = errors.New("saved graph retained state is incompatible with contract major 6")

// ValidateSavedGraphAdmission reads all retained state in the caller's read-only
// snapshot. It neither upgrades old receipts nor treats job expiry as expiry of
// route replay obligations. Page sizes bound memory independently of retention.
func ValidateSavedGraphAdmission(ctx context.Context, reader extensionstore.Querier, definition jobs.Definition) error {
	if definition.JobKind != GraphViewMaterializationJobKind || definition.HandlerName != graphViewWorkerKind || definition.Extension == nil || definition.Extension.OwnerProfileID != "network_flow_activity" {
		return errSavedGraphCutoverIncompatible
	}
	if err := validatePersistedGraphViewFamily(ctx, reader); err != nil {
		return errSavedGraphCutoverIncompatible
	}
	if err := validateSavedGraphSelectedBindings(ctx, reader); err != nil {
		return err
	}
	routes := []string{routeKeyGraphViewsCreate, routeKeyGraphViewsPatch, routeKeyGraphViewsRefresh, routeKeyGraphViewsDelete}
	after := uuid.Nil
	for {
		records, next, err := authn.ReadRouteIdempotencyPage(ctx, reader, routes, after)
		if err != nil {
			return err
		}
		for _, record := range records {
			if err := validateSavedGraphAdmissionReceipt(ctx, reader, definition, record); err != nil {
				return err
			}
		}
		if len(records) < 128 {
			break
		}
		after = next
	}
	after = uuid.Nil
	for {
		ids, err := jobs.ReadRetainedExtensionJobPage(ctx, reader, "network_flow_activity", GraphViewMaterializationJobKind, graphViewWorkerKind, after)
		if err != nil {
			return err
		}
		for _, id := range ids {
			if err := validateSavedGraphRetainedJob(ctx, reader, definition, id); err != nil {
				return err
			}
			after = id
		}
		if len(ids) < 128 {
			break
		}
	}
	after = uuid.Nil
	for {
		ids, err := extensionstore.ReadJobCommitProofPage(ctx, reader, "network_flow_activity", after)
		if err != nil {
			return err
		}
		for _, id := range ids {
			if err := validateSavedGraphRetainedJob(ctx, reader, definition, id); err != nil {
				return err
			}
			after = id
		}
		if len(ids) < 128 {
			break
		}
	}
	return nil
}

func savedGraphReceiptKey(record authn.RouteIdempotencyRecord) authn.RouteIdempotencyKey {
	return authn.RouteIdempotencyKey{RouteKey: record.RouteKey, ScopeKey: record.ScopeKey, ActorUserID: record.ActorUserID, ClientTxnID: record.ClientTxnID}
}

func validateSavedGraphAdmissionReceipt(ctx context.Context, reader extensionstore.Querier, definition jobs.Definition, record authn.RouteIdempotencyRecord) error {
	key := savedGraphReceiptKey(record)
	payload, err := decodeStoredNetworkFlowResponse(record.ResponseJSON)
	if err != nil || len(record.RequestHash) != sha256.Size || validateGraphViewReceipt(key, record.StatusCode, payload, nil) != nil {
		return errSavedGraphCutoverIncompatible
	}
	if record.RouteKey == routeKeyGraphViewsDelete {
		return nil
	}
	graph, _ := graphViewDeclarationFromPublicResource(payload["graph_view"].(map[string]any))
	_, path, _ := strings.Cut(key.ScopeKey, ":")
	comparison := map[string]any{}
	switch record.RouteKey {
	case routeKeyGraphViewsCreate:
		comparison["display_name"] = graph.DisplayName
		comparison["semantic_query"] = payload["graph_view"].(map[string]any)["semantic_query"]
	case routeKeyGraphViewsPatch:
		comparison["display_name"] = graph.DisplayName
		comparison["base_graph_view_version"] = graph.GraphViewVersion - 1
	case routeKeyGraphViewsRefresh:
		comparison["base_graph_view_version"] = graph.GraphViewVersion - 1
	}
	matches := func() bool {
		digest := sha256.Sum256(graphViewMutationBytes(key.RouteKey, path, comparison))
		return bytes.Equal(digest[:], record.RequestHash)
	}
	if !matches() {
		// A same-name rename acknowledges the unchanged version.
		if record.RouteKey != routeKeyGraphViewsPatch {
			return errSavedGraphCutoverIncompatible
		}
		comparison["base_graph_view_version"] = graph.GraphViewVersion
		if !matches() {
			return errSavedGraphCutoverIncompatible
		}
	}
	if record.RouteKey == routeKeyGraphViewsPatch {
		return nil
	}
	retained, err := jobs.ReadRetainedExtensionJob(ctx, reader, *graph.LatestJobID)
	if err != nil {
		return errSavedGraphCutoverIncompatible
	}
	if err := validateSavedGraphJobFacts(ctx, reader, definition, retained, key, record.RequestHash, graph); err != nil {
		return err
	}
	return nil
}

func savedGraphJobIdentity(raw json.RawMessage) (authn.RouteIdempotencyKey, error) {
	var identity map[string]any
	if json.Unmarshal(raw, &identity) != nil || len(identity) != 6 || identity["schema_id"] != "cartulary.route_scoped_idempotency_identity.v1" {
		return authn.RouteIdempotencyKey{}, errSavedGraphCutoverIncompatible
	}
	actor, ok := identity["actor_user_id"].(string)
	if !ok {
		return authn.RouteIdempotencyKey{}, errSavedGraphCutoverIncompatible
	}
	actorID, err := uuid.Parse(actor)
	if err != nil || actorID == uuid.Nil {
		return authn.RouteIdempotencyKey{}, errSavedGraphCutoverIncompatible
	}
	route, ok := identity["route_identity"].(string)
	if !ok {
		return authn.RouteIdempotencyKey{}, errSavedGraphCutoverIncompatible
	}
	txn, ok := identity["client_txn_id"].(string)
	if !ok || txn == "" {
		return authn.RouteIdempotencyKey{}, errSavedGraphCutoverIncompatible
	}
	for _, kind := range []string{routeKeyGraphViewsCreate, routeKeyGraphViewsRefresh} {
		if scope, ok := strings.CutPrefix(route, kind+":"); ok {
			incident, _, found := strings.Cut(scope, ":")
			if !found || identity["scope_kind"] != jobs.ScopeKindIncident || identity["scope_id"] != incident {
				return authn.RouteIdempotencyKey{}, errSavedGraphCutoverIncompatible
			}
			return authn.RouteIdempotencyKey{RouteKey: kind, ScopeKey: scope, ActorUserID: actorID, ClientTxnID: txn}, nil
		}
	}
	return authn.RouteIdempotencyKey{}, errSavedGraphCutoverIncompatible
}

func validateSavedGraphRetainedJob(ctx context.Context, reader extensionstore.Querier, definition jobs.Definition, id uuid.UUID) error {
	retained, err := jobs.ReadRetainedExtensionJob(ctx, reader, id)
	if err != nil {
		return errSavedGraphCutoverIncompatible
	}
	if !savedGraphRegistrationMatches(retained, definition) {
		return errSavedGraphCutoverIncompatible
	}
	identity := retained.IdempotencyIdentity
	if retained.ExpiredAt != nil {
		proof, proofErr := extensionstore.ReadJobCommitProof(ctx, reader, id)
		if retained.Resource.Status != jobs.StatusSucceeded {
			if !errors.Is(proofErr, extensionstore.ErrNotFound) || (retained.Resource.Status != jobs.StatusFailed && retained.Resource.Status != jobs.StatusCanceled) {
				return errSavedGraphCutoverIncompatible
			}
			return nil
		}
		if proofErr != nil {
			return errSavedGraphCutoverIncompatible
		}
		identity = proof.IdempotencyIdentity
	}
	key, err := savedGraphJobIdentity(identity)
	if err != nil {
		return err
	}
	record, err := authn.GetRouteIdempotencyTx(ctx, reader, key)
	if err != nil {
		return errSavedGraphCutoverIncompatible
	}
	payload, err := decodeStoredNetworkFlowResponse(record.ResponseJSON)
	if err != nil || validateGraphViewReceipt(key, record.StatusCode, payload, nil) != nil {
		return errSavedGraphCutoverIncompatible
	}
	graph, err := graphViewDeclarationFromPublicResource(payload["graph_view"].(map[string]any))
	if err != nil || graph.LatestJobID == nil || *graph.LatestJobID != id {
		return errSavedGraphCutoverIncompatible
	}
	return validateSavedGraphJobFacts(ctx, reader, definition, retained, key, record.RequestHash, graph)
}

func savedGraphRegistrationMatches(retained jobs.RetainedExtensionJob, definition jobs.Definition) bool {
	return retained.JobKind == definition.JobKind && retained.HandlerName == definition.HandlerName && retained.ProgressUnitID == definition.ProgressUnitID && retained.OwnerProfileID == definition.Extension.OwnerProfileID && retained.Resource.AuthPolicy == jobs.AuthPolicyIncidentMembership
}

func validateSavedGraphJobFacts(ctx context.Context, reader extensionstore.Querier, definition jobs.Definition, retained jobs.RetainedExtensionJob, key authn.RouteIdempotencyKey, hash []byte, graph graphViewDeclaration) error {
	if !savedGraphRegistrationMatches(retained, definition) || retained.Resource.Scope.IncidentID == nil || *retained.Resource.Scope.IncidentID != graph.IncidentID || retained.Resource.Scope.Kind != jobs.ScopeKindIncident || retained.Resource.SubmittedByUserID != key.ActorUserID.String() {
		return errSavedGraphCutoverIncompatible
	}
	id, err := uuid.Parse(retained.Resource.JobID)
	if err != nil || graph.LatestJobID == nil || id != *graph.LatestJobID {
		return errSavedGraphCutoverIncompatible
	}
	hashText := hex.EncodeToString(hash)
	if retained.ExpiredAt == nil {
		identity, err := savedGraphJobIdentity(retained.IdempotencyIdentity)
		if err != nil || identity != key || retained.RouteKey != key.RouteKey || retained.ScopeKey != key.ScopeKey || retained.RequestSHA256 != hashText {
			return errSavedGraphCutoverIncompatible
		}
		var payload graphViewMaterializationPayload
		var members map[string]any
		if json.Unmarshal(retained.Payload, &members) != nil || len(members) != 5 || json.Unmarshal(retained.Payload, &payload) != nil || !payload.valid() || payload.IncidentID != graph.IncidentID || payload.GraphViewID != graph.GraphViewID || payload.MaterializationGeneration != graph.MaterializationGeneration || payload.SourceSnapshotID != graph.DesiredSourceSnapshotID {
			return errSavedGraphCutoverIncompatible
		}
	}
	proof, proofErr := extensionstore.ReadJobCommitProof(ctx, reader, id)
	if retained.Resource.Status != jobs.StatusSucceeded {
		switch retained.Resource.Status {
		case jobs.StatusQueued, jobs.StatusRunning, jobs.StatusCancelRequested, jobs.StatusFailed, jobs.StatusCanceled:
		default:
			return errSavedGraphCutoverIncompatible
		}
		if !errors.Is(proofErr, extensionstore.ErrNotFound) {
			return errSavedGraphCutoverIncompatible
		}
		return nil
	}
	if proofErr != nil || extensionstore.ValidateJobCommitProofSize(*proof, definition.Extension.MaxProofBytes) != nil || proof.OwnerProfileID != definition.Extension.OwnerProfileID || proof.OperationKind != definition.Extension.OperationKind || proof.FinalCommitID == "" || proof.CommittedAt.IsZero() || proof.NormalizedRequestSHA256 != hashText {
		return errSavedGraphCutoverIncompatible
	}
	identity, err := savedGraphJobIdentity(proof.IdempotencyIdentity)
	if err != nil || identity != key {
		return errSavedGraphCutoverIncompatible
	}
	var summary jobs.ResultSummary
	if json.Unmarshal(proof.TerminalResult, &summary) != nil {
		return errSavedGraphCutoverIncompatible
	}
	_, terminal, refs, digest, err := jobs.CanonicalExtensionTerminalSuccess(definition, &summary)
	if err != nil || digest != proof.TerminalResultSHA256 || !savedGraphJSONEqual(terminal, proof.TerminalResult) || !savedGraphJSONEqual(refs, proof.ResourceRefs) || summary.Code != "network_flow_graph_view_materialized" || len(summary.ResourceRefs) != 1 {
		return errSavedGraphCutoverIncompatible
	}
	ref := summary.ResourceRefs[0]
	if ref.Kind != graphViewResultResourceKind || ref.ID != graph.GraphViewID || ref.Route != graphViewRoute(graph.IncidentID, graph.GraphViewID) {
		return errSavedGraphCutoverIncompatible
	}
	if retained.ExpiredAt == nil {
		raw, _ := json.Marshal(retained.Resource.ResultSummary)
		if !savedGraphJSONEqual(raw, terminal) {
			return errSavedGraphCutoverIncompatible
		}
	}
	return nil
}

func savedGraphJSONEqual(a, b []byte) bool {
	var left, right any
	return json.Unmarshal(a, &left) == nil && json.Unmarshal(b, &right) == nil && bytes.Equal(canonicalJSON(left), canonicalJSON(right))
}

// ValidateRetainedExtensionState implements the existing post-migration owner
// validator over the same read-only admission capability.
func ValidateRetainedExtensionState(ctx context.Context, reader extensionstore.Querier) error {
	return ValidateExtensionState(ctx, savedGraphAdmissionFamilies{reader: reader})
}

type savedGraphAdmissionFamilies struct{ reader extensionstore.Querier }

func (r savedGraphAdmissionFamilies) FamilyCounts(ctx context.Context, ids []string) (map[string]int64, error) {
	result := map[string]int64{}
	for _, id := range ids {
		for _, family := range ExtensionStateFamilyCounters() {
			if family.FamilyID == id {
				count, err := family.Count(ctx, r.reader)
				if err != nil {
					return nil, err
				}
				result[id] = count
			}
		}
	}
	return result, nil
}
func (r savedGraphAdmissionFamilies) ValidateFamilyState(ctx context.Context, id string) error {
	for _, family := range ExtensionStateFamilyCounters() {
		if family.FamilyID == id {
			if family.Validate != nil {
				return family.Validate(ctx, r.reader)
			}
			return nil
		}
	}
	return errSavedGraphCutoverIncompatible
}

func validateSavedGraphSelectedBindings(ctx context.Context, reader extensionstore.Querier) error {
	results, err := postgresresult.NewReader(reader)
	if err != nil {
		return err
	}
	after := ""
	for {
		rows, err := reader.Query(ctx, graphViewDeclarationSelect+` WHERE graph_view_id > $1 AND selected_projection_result_id IS NOT NULL ORDER BY graph_view_id LIMIT 128`, after)
		if err != nil {
			return err
		}
		declarations := make([]graphViewDeclaration, 0, 128)
		for rows.Next() {
			declaration, err := scanGraphViewDeclaration(rows)
			if err != nil {
				rows.Close()
				return err
			}
			declarations = append(declarations, declaration)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		for _, declaration := range declarations {
			binding, err := results.ReadResultEnvelope(ctx, declaration.SelectedResult.ProjectionResultID)
			selected := declaration.SelectedResult
			if err != nil || binding.GraphViewID != declaration.GraphViewID || binding.SourceOwnerID != ProfileID || binding.SourceSnapshotID != selected.SourceSnapshotID || binding.ProjectionSchemaID != selected.ProjectionSchemaID || binding.ProjectionVersion != selected.ProjectionVersion || binding.NormalizedConfigurationSHA256 != selected.NormalizedConfigurationSHA256 || binding.NormalizedSourceSHA256 != selected.NormalizedSourceSHA256 || binding.CanonicalOutputSHA256 != selected.CanonicalOutputSHA256 {
				return errSavedGraphCutoverIncompatible
			}
			after = declaration.GraphViewID
		}
		if len(declarations) < 128 {
			return nil
		}
	}
}
