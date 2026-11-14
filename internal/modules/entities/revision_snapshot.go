package entities

import (
	"fmt"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/deleterestore"
)

func validateHostSnapshotSource(record, source map[string]any) error {
	if err := deleterestore.ValidateHostSnapshotSource(record, source); err != nil {
		return err
	}
	return validateEntitySnapshotSemantics(record, source, "host_state",
		[]string{"hostname", "aad_device_id", "fqdn"},
		[]string{"location", "os_platform", "business_owner", "criticality", "containment_status"})
}

func validateIdentitySnapshotSource(record, source map[string]any) error {
	if err := deleterestore.ValidateIdentitySnapshotSource(record, source); err != nil {
		return err
	}
	return validateEntitySnapshotSemantics(record, source, "identity_state",
		[]string{"upn", "email", "sam_account_name", "aad_object_id", "sid"},
		[]string{"privilege_level", "mfa_state", "reset_status"})
}

// These checks concern the retained row and its own historical envelope. They
// neither resolve historical actors against a fabricated catalog nor consult
// current active entities to reinterpret a previous merge or identifier value.
func validateEntitySnapshotSemantics(record, source map[string]any, stateField string, identifiers, annotations []string) error {
	if _, ok := entityPortableText(source["display_name"], false); !ok {
		return fmt.Errorf("invalid entity snapshot display name")
	}
	for _, name := range identifiers {
		if _, ok := entityNullableText(source[name], false); !ok {
			return fmt.Errorf("invalid entity snapshot identifier %q", name)
		}
	}
	for _, name := range annotations {
		if _, ok := entityNullableText(source[name], true); !ok {
			return fmt.Errorf("invalid entity snapshot annotation %q", name)
		}
	}
	if !entityOriginAllowed(source["entity_origin"].(string)) {
		return fmt.Errorf("invalid entity snapshot origin")
	}
	state := source[stateField].(string)
	if !entityStateAllowed(state) || (state == "merged") != (source["merged_into_record_id"] != nil) ||
		source["merged_into_record_id"] == source["record_id"] {
		return fmt.Errorf("invalid entity snapshot merge tuple")
	}
	version, ok := entityCanonicalInteger(source["row_version"])
	envelopeVersion, envelopeOK := entityCanonicalInteger(record["row_version"])
	if !ok || version < 1 || !envelopeOK || version != envelopeVersion {
		return fmt.Errorf("entity snapshot version disagrees with envelope")
	}
	var clocks [2]time.Time
	for index, name := range []string{"created_at", "updated_at"} {
		sourceText, sourceOK := source[name].(string)
		envelopeText, recordOK := record[name].(string)
		sourceTime, sourceErr := time.Parse(time.RFC3339Nano, sourceText)
		envelopeTime, recordErr := time.Parse(time.RFC3339Nano, envelopeText)
		if !sourceOK || !recordOK || sourceErr != nil || recordErr != nil || !sourceTime.Equal(envelopeTime) {
			return fmt.Errorf("entity snapshot %s disagrees with envelope", name)
		}
		clocks[index] = sourceTime
	}
	if clocks[1].Before(clocks[0]) {
		return fmt.Errorf("entity snapshot clocks are reversed")
	}
	for _, name := range []string{"created_by_user_id", "updated_by_user_id"} {
		if source[name] != record[name] {
			return fmt.Errorf("entity snapshot %s disagrees with envelope", name)
		}
	}
	return nil
}
