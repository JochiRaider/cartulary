package textpolicy

import (
	"fmt"
	"maps"
	"slices"

	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

// ValidateStoredSource checks the exact persisted representation of bound text.
// Callers own the complete row shape, lifecycle and reference validation. Stored
// history and portable input are never normalized into acceptance here.
func ValidateStoredSource(recordType string, source map[string]any) error {
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return err
	}
	surface, ok := catalog.SurfaceByRecordType(recordType)
	if !ok {
		return fmt.Errorf("unknown TasksDecisions source type")
	}
	fields := catalog.Fields()
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		mapping := fields[key]
		if mapping.ViewSchemaID != surface.ViewSchemaID || mapping.Kind != sourcecatalog.FieldKindDirect {
			continue
		}
		field, ok := viewschema.LookupField(mapping.ViewSchemaID, mapping.FieldKey)
		if !ok {
			return fmt.Errorf("missing TasksDecisions source field binding")
		}
		if field.StringContractID == nil {
			continue
		}
		raw, present := source[mapping.Storage.Column]
		if present && raw == nil && field.Clearable {
			continue
		}
		value, ok := raw.(string)
		if !present || !ok {
			return fmt.Errorf("invalid stored text shape for %s", mapping.FieldKey)
		}
		normalized, valid := Normalize(value, *field.StringContractID)
		if !valid || normalized == "" || normalized != value {
			return fmt.Errorf("invalid stored text for %s", mapping.FieldKey)
		}
	}
	return nil
}
