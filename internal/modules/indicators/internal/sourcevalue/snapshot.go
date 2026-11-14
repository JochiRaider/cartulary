package sourcevalue

import (
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
	shape "github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

func ValidateSnapshotSource(record, source map[string]any) error {
	if err := shape.ValidateSourceObject(record, source,
		shape.SourceFieldGroup{Names: "record_id incident_id", Kind: shape.SourceUUID, Nullable: false},
		shape.SourceFieldGroup{Names: "indicator_type value_kind display_value normalized_value dedupe_key", Kind: shape.SourceString, Nullable: false},
		shape.SourceFieldGroup{Names: "defanged_value hash_algorithm hash_value stix_pattern", Kind: shape.SourceString, Nullable: true},
	); err != nil {
		return err
	}
	text := func(name string) *string {
		if source[name] == nil {
			return nil
		}
		value := source[name].(string)
		return &value
	}
	return identity.ValidateStoredShape(identity.Input{
		IndicatorType: source["indicator_type"].(string), ValueKind: source["value_kind"].(string),
		DisplayValue: source["display_value"].(string), NormalizedValue: text("normalized_value"),
		DefangedValue: text("defanged_value"), HashAlgorithm: text("hash_algorithm"),
		HashValue: text("hash_value"), STIXPattern: text("stix_pattern"),
	}, source["dedupe_key"].(string))
}
