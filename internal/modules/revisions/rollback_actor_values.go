package revisions

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Runtime images are separate from retained history and from the plan's source
// values. Source providers still validate all values and compare complete rows.
func (s *commandStore) runtimeHistoryValuesTx(ctx context.Context, tx pgx.Tx, incidentID, changeSetID uuid.UUID, values ...map[string]any) ([]map[string]any, error) {
	actors := map[string]struct{}{}
	for _, value := range values {
		if !walkSnapshotActors(value, func(id string) bool {
			actors[id] = struct{}{}
			return true
		}) {
			return nil, errors.New("revisions: invalid historical actor")
		}
	}
	ids := make([]string, 0, len(actors))
	for id := range actors {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	mapping, err := s.importedAttributions.ResolveImportedHistoryActorsTx(ctx, tx, incidentID, changeSetID, ids)
	if err != nil {
		return nil, err
	}
	if mapping == nil {
		return values, nil
	}
	for _, id := range ids {
		if _, valid := canonicalRevisionsUUID(mapping[id]); !valid {
			return nil, errors.New("revisions: incomplete imported history actor mapping")
		}
	}
	result := make([]map[string]any, len(values))
	for index, value := range values {
		if value != nil {
			result[index] = runtimeHistoryActorValue(value, mapping).(map[string]any)
		}
	}
	return result, nil
}

func runtimeHistoryActorValue(value any, mapping map[string]string) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, member := range typed {
			if id, ok := member.(string); ok && strings.HasSuffix(key, "_user_id") {
				result[key] = mapping[id]
			} else {
				result[key] = runtimeHistoryActorValue(member, mapping)
			}
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, item := range typed {
			result[index] = runtimeHistoryActorValue(item, mapping)
		}
		return result
	default:
		return value
	}
}
