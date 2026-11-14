package deleterestore

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/parties/internal/policy"
	shape "github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

func ValidateSnapshotSource(record, source map[string]any) error {
	if err := shape.ValidateSourceObject(record, source,
		shape.SourceFieldGroup{Names: "record_id incident_id", Kind: shape.SourceUUID, Nullable: false},
		shape.SourceFieldGroup{Names: "display_name party_kind", Kind: shape.SourceString, Nullable: false},
		shape.SourceFieldGroup{Names: "organization_name role_title primary_email timezone_name external_ref notes", Kind: shape.SourceString, Nullable: true},
	); err != nil {
		return err
	}
	for _, key := range policy.FieldKeys() {
		field, ok := policy.LookupField(key)
		if !ok {
			return fmt.Errorf("unknown Party snapshot field %q", key)
		}
		raw, present := source[field.SourceColumn]
		if !present {
			return fmt.Errorf("missing Party snapshot field %q", key)
		}
		var text *string
		if raw != nil {
			value, ok := raw.(string)
			if !ok {
				return fmt.Errorf("invalid Party snapshot field %q", key)
			}
			text = &value
		}
		if !policy.IsStoredExact(key, text) {
			return fmt.Errorf("noncanonical Party snapshot field %q", key)
		}
	}
	return nil
}
