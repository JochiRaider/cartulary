package revisions

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"slices"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// A path belongs to a capture, not to its actor's current account or attribution.
// Paths are appended only when inserting retained history into a new deployment.
type revisionActorImportContext struct {
	ImportID uuid.UUID `json:"import_id"`
	ActorID  uuid.UUID `json:"actor_user_id"`
}

// Reserve the complete canonical row's capacity before admitting immutable
// source bytes. Incoming JSON may be shorter than its canonical export encoding.
func revisionChangeSetExportFits(row map[string]any, input sourceport.ImportContext) bool {
	path, ok := row["actor_import_contexts"].([]any)
	if !ok {
		return false
	}
	exported := maps.Clone(row)
	exported["actor_import_contexts"] = append(slices.Clone(path), map[string]any{
		"import_id": input.OperationID, "actor_user_id": input.ActorUserID.String(),
	})
	encoded, err := incidentportability.CanonicalJSONString(exported)
	return err == nil && len(encoded)-1 <= incidentportability.MaxNDJSONLogicalLineBytes
}

func parseRevisionActorContexts(value any, actors sourceport.ActorCatalog) ([]revisionActorImportContext, string) {
	items, ok := value.([]any)
	if !ok {
		return nil, revisionsReferencesInvariant
	}
	result := make([]revisionActorImportContext, 0, len(items))
	// Shape/reference defects across the entire path precede actor defects.
	for _, item := range items {
		row, ok := item.(map[string]any)
		if !ok || !exactRevisionsMembers(row, []string{"import_id", "actor_user_id"}) {
			return nil, revisionsReferencesInvariant
		}
		id, ok := canonicalRevisionsUUID(row["import_id"])
		if !ok {
			return nil, revisionsReferencesInvariant
		}
		result = append(result, revisionActorImportContext{ImportID: id})
	}
	for index, item := range items {
		actor, ok := canonicalRevisionsUUID(item.(map[string]any)["actor_user_id"])
		if !ok {
			return nil, revisionsActorsInvariant
		}
		if _, ok := actors.Lookup(actor.String()); !ok {
			return nil, revisionsActorsInvariant
		}
		result[index].ActorID = actor
	}
	return result, ""
}

func uniqueRevisionActorContexts(path []revisionActorImportContext, current uuid.UUID) bool {
	seen := make(map[uuid.UUID]bool, len(path))
	for _, step := range path {
		if step.ImportID == current || seen[step.ImportID] {
			return false
		}
		seen[step.ImportID] = true
	}
	return true
}

func revisionActorContextSuffix(path, suffix []revisionActorImportContext) bool {
	return len(path) >= len(suffix) && slices.Equal(path[len(path)-len(suffix):], suffix)
}

func (p preparedRevisionsImport) validateActorContextPaths(ctx context.Context) error {
	var longest []revisionActorImportContext
	if err := p.walkChangeSets(ctx, func(row portableChangeSet) error {
		if len(row.ActorImportContexts) > len(longest) {
			longest = row.ActorImportContexts
		}
		return nil
	}); err != nil {
		return err
	}
	return p.walkChangeSets(ctx, func(row portableChangeSet) error {
		if !revisionActorContextSuffix(longest, row.ActorImportContexts) {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		return nil
	})
}

func (p preparedRevisionsImport) importedActorContexts(path []revisionActorImportContext) []revisionActorImportContext {
	return append(slices.Clone(path), p.actorContext)
}

func decodeRevisionActorContexts(raw []byte) ([]revisionActorImportContext, error) {
	var result []revisionActorImportContext
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	value, err := decodeDatabaseJSON(raw)
	if err != nil {
		return nil, err
	}
	if result == nil || !sameCanonicalJSON(value, revisionActorContextValues(result)) || !uniqueRevisionActorContexts(result, uuid.Nil) {
		return nil, errors.New("revisions: invalid stored actor import contexts")
	}
	for _, step := range result {
		if step.ActorID == uuid.Nil {
			return nil, errors.New("revisions: invalid stored context actor")
		}
	}
	return result, nil
}

func revisionActorContextValues(path []revisionActorImportContext) []any {
	result := make([]any, 0, len(path))
	for _, step := range path {
		result = append(result, map[string]any{"import_id": step.ImportID.String(), "actor_user_id": step.ActorID.String()})
	}
	return result
}

func revisionSnapshotsContinue(previous, before map[string]any, priorPath, nextPath []revisionActorImportContext) bool {
	if !revisionActorContextSuffix(priorPath, nextPath) {
		return false
	}
	if len(priorPath) == len(nextPath) {
		return sameCanonicalJSON(previous, before)
	}
	actor := priorPath[len(priorPath)-len(nextPath)-1].ActorID.String()
	mapping := map[string]string{}
	if !walkSnapshotActors(previous, func(id string) bool {
		mapping[id] = actor
		return true
	}) {
		return false
	}
	return sameCanonicalJSON(runtimeHistoryActorValue(previous, mapping), before)
}
