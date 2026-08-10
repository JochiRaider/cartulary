package artifacts

import "github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/sourcecatalog"

func lookupArtifactSourceSurface(viewSchemaID string) (sourcecatalog.Surface, bool) {
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return sourcecatalog.Surface{}, false
	}
	return catalog.SurfaceByViewID(viewSchemaID)
}

func lookupArtifactSourceField(fieldKey string) (sourcecatalog.Field, bool) {
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return sourcecatalog.Field{}, false
	}
	return catalog.Field(fieldKey)
}

func validateArtifactDirectValue(policy sourcecatalog.Field, value FieldValue) error {
	count := 0
	if value.Text != nil {
		count++
	}
	if value.Timestamp != nil {
		count++
	}
	if value.UUID != nil {
		count++
	}
	if value.Number != nil {
		count++
	}
	if value.Bool != nil {
		count++
	}
	if count == 0 {
		if policy.View.Clearable {
			return nil
		}
		return &ValidationError{Field: policy.FieldKey, ReasonCode: "field_not_nullable"}
	}
	if count != 1 {
		return &ValidationError{Field: policy.FieldKey, ReasonCode: "invalid_value"}
	}
	validKind := false
	switch {
	case policy.View.ReferenceContractID != "":
		validKind = value.UUID != nil
	case policy.View.ReadKind == "timestamp":
		validKind = value.Timestamp != nil
	case policy.View.ReadKind == "number":
		validKind = value.Number != nil
	case policy.View.ReadKind == "boolean":
		validKind = value.Bool != nil
	case policy.View.ReadKind == "text" || policy.View.ReadKind == "enum":
		validKind = value.Text != nil
	}
	if !validKind {
		return &ValidationError{Field: policy.FieldKey, ReasonCode: "invalid_value"}
	}
	return nil
}
