package networkflow

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

// Retained outcomes are admitted through the same closed current contract as
// new outcomes. A version token alone is insufficient evidence of compatibility.
func validateRetainedImportResult(raw json.RawMessage) error {
	var result struct {
		SchemaID    string          `json:"schema_id"`
		SessionID   uuid.UUID       `json:"import_session_id"`
		UnitID      uuid.UUID       `json:"import_unit_id"`
		SourceSHA   string          `json:"source_content_sha256"`
		Profile     string          `json:"source_profile_id"`
		Parser      string          `json:"parser_profile_id"`
		Fingerprint string          `json:"mapping_fingerprint"`
		TableRef    json.RawMessage `json:"table_ref"`
	}
	if _, err := decodeClosedImportObject(raw, "schema_id import_session_id import_unit_id source_content_sha256 source_profile_id parser_profile_id mapping_fingerprint table_ref", "", "", &result); err != nil {
		return errSavedGraphCutoverIncompatible
	}
	profile, ok := mappingSourceProfile(result.Profile)
	if result.SchemaID != importApplyResultID || !ok || result.Parser != profile.ParserProfileID || result.SessionID == uuid.Nil || result.UnitID == uuid.Nil || !validLowerSHA256(result.SourceSHA) || !validLowerSHA256(result.Fingerprint) {
		return errSavedGraphCutoverIncompatible
	}
	var ref struct {
		Kind        string `json:"kind"`
		ID          string `json:"id"`
		Route       string `json:"route"`
		DisplayName string `json:"display_name"`
		Accepted    int64  `json:"row_count_accepted"`
		Rejected    int64  `json:"row_count_rejected"`
		Truncated   bool   `json:"diagnostics_truncated"`
		Version     int64  `json:"table_version"`
	}
	if _, err := decodeClosedImportObject(result.TableRef, "kind id route display_name row_count_accepted row_count_rejected diagnostics_truncated table_version", "", "", &ref); err != nil {
		return errSavedGraphCutoverIncompatible
	}
	name, err := normalizeExplicitDisplayName(ref.DisplayName)
	segments := strings.Split(ref.Route, "/")
	if err != nil || name != ref.DisplayName || ref.Kind != targetKindNetworkFlowTable || !linkTableIDPattern.MatchString(ref.ID) || ref.Accepted < 0 || ref.Rejected < 0 || ref.Version < 1 || len(segments) != 8 || segments[1] != "api" || segments[2] != "v1" || segments[3] != "incidents" || !linkUUIDPattern.MatchString(segments[4]) || segments[5] != "network-flow" || segments[6] != "tables" || segments[7] != ref.ID {
		return errSavedGraphCutoverIncompatible
	}
	return nil
}
