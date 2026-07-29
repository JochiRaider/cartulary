package imports

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

func (s *Service) applyUnit(ctx context.Context, actor authn.UserRecord, start ApplyStartResult, unit ApplyUnitData) ([]jobs.ResourceRef, error) {
	target, ok := lookupApprovedImportTarget(unit.ApprovedMapping)
	if !ok || !target.importable(s.extensionProfileClaimed) {
		if unit.ApprovedMapping.targetKindOrDefault() == ImportTargetKindViewSchema {
			return nil, importApplyBlockedError("target_view_schema_not_importable")
		}
		return nil, importApplyBlockedError("target_kind_not_importable")
	}
	if unit.ApprovedMapping.targetKindOrDefault() != ImportTargetKindViewSchema {
		if !target.ownerApplyFacadeAvailable() {
			return nil, importApplyBlockedError("owner_apply_contract_unavailable")
		}
		return s.applyExtensionOwnerUnit(ctx, actor, start, unit, target)
	}
	if !target.ownerCreateFacadeAvailable() {
		return nil, importApplyBlockedError("owner_create_contract_unavailable")
	}
	return nil, s.applyGenericOwnerUnit(ctx, actor, start, unit, target)
}

func importApplyResourceRefs(sessionID uuid.UUID, extensionRefs []jobs.ResourceRef) []jobs.ResourceRef {
	refs := make([]jobs.ResourceRef, 0, 1+len(extensionRefs))
	refs = append(refs, jobs.ResourceRef{
		Kind:  "import_session",
		ID:    sessionID.String(),
		Route: "/api/v1/import-sessions/" + sessionID.String(),
	})
	if len(extensionRefs) == 0 {
		return refs
	}
	sorted := append([]jobs.ResourceRef(nil), extensionRefs...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Kind != sorted[j].Kind {
			return sorted[i].Kind < sorted[j].Kind
		}
		if sorted[i].Route != sorted[j].Route {
			return sorted[i].Route < sorted[j].Route
		}
		return sorted[i].ID < sorted[j].ID
	})
	return append(refs, sorted...)
}

func sourceRowCellsByOrdinal(sourceRow map[string]any) map[int]map[string]any {
	cellsByOrdinal := map[int]map[string]any{}
	if cells, ok := sourceRow["cells"].([]any); ok {
		for _, rawCell := range cells {
			cell, ok := rawCell.(map[string]any)
			if !ok {
				continue
			}
			ordinal, ok := intFromAny(cell["source_column_ordinal"])
			if ok {
				cellsByOrdinal[ordinal] = cell
			}
		}
		return cellsByOrdinal
	}
	if cells, ok := sourceRow["cells"].([]map[string]any); ok {
		for _, cell := range cells {
			ordinal, ok := intFromAny(cell["source_column_ordinal"])
			if ok {
				cellsByOrdinal[ordinal] = cell
			}
		}
	}
	return cellsByOrdinal
}

func transformImportValue(value string, column SourceColumnMapping) (string, error) {
	result := value
	if column.TransformID == nil {
		return result, nil
	}
	switch *column.TransformID {
	case "trim_v1":
		return strings.TrimSpace(result), nil
	case "collapse_whitespace_v1":
		return strings.Join(strings.Fields(result), " "), nil
	case "lowercase_v1":
		return strings.ToLower(result), nil
	case "split_delimited_v1":
		delimiter, _ := column.TransformOptions["delimiter"].(string)
		trimItems, _ := column.TransformOptions["trim_items"].(bool)
		dropEmpty, _ := column.TransformOptions["drop_empty_items"].(bool)
		parts := strings.Split(result, delimiter)
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimItems {
				part = strings.TrimSpace(part)
			}
			if dropEmpty && part == "" {
				continue
			}
			out = append(out, part)
		}
		return strings.Join(out, delimiter), nil
	default:
		return "", fmt.Errorf("unsupported transform %q", *column.TransformID)
	}
}
