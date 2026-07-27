package reference_data

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

const (
	IncidentBundleReferenceExactShapeInvariant = "reference_pack_refs.exact_shape"
	IncidentBundleReferenceIdentityInvariant   = "reference_pack_refs.identity_exact"
)

type IncidentBundleReferenceValidationError struct {
	InvariantID string
}

func (e *IncidentBundleReferenceValidationError) Error() string {
	return "incident bundle reference-pack references violate " + e.InvariantID
}

func IncidentBundleReferenceInvariant(err error) (string, bool) {
	var validationErr *IncidentBundleReferenceValidationError
	if !errors.As(err, &validationErr) {
		return "", false
	}
	return validationErr.InvariantID, true
}

// ValidateIncidentBundleReferences is the Reference Pack owner's closed,
// non-mutating validator for the Incident Bundle reference catalog.
func ValidateIncidentBundleReferences(payload []byte) error {
	var rows []map[string]any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&rows); err != nil {
		return &IncidentBundleReferenceValidationError{InvariantID: IncidentBundleReferenceExactShapeInvariant}
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return &IncidentBundleReferenceValidationError{InvariantID: IncidentBundleReferenceExactShapeInvariant}
	}

	required := map[string]struct{}{
		"pack_key": {}, "pack_version": {}, "manifest_sha256": {},
		"payload_sha256": {}, "pack_contract_version": {}, "content_profile_id": {},
		"content_profile_version": {}, "distribution_kind": {}, "verification_method": {},
		"source_profile_id": {}, "source_profile_sha256": {},
	}
	seen := map[string]struct{}{}
	for _, row := range rows {
		if len(row) != len(required) {
			return &IncidentBundleReferenceValidationError{InvariantID: IncidentBundleReferenceExactShapeInvariant}
		}
		for field := range required {
			value, ok := row[field].(string)
			if !ok || strings.TrimSpace(value) == "" {
				return &IncidentBundleReferenceValidationError{InvariantID: IncidentBundleReferenceExactShapeInvariant}
			}
		}
		identity := row["pack_key"].(string) + "\x00" + row["pack_version"].(string)
		if _, duplicate := seen[identity]; duplicate {
			return &IncidentBundleReferenceValidationError{InvariantID: IncidentBundleReferenceIdentityInvariant}
		}
		seen[identity] = struct{}{}
	}
	return nil
}
