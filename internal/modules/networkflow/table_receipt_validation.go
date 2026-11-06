package networkflow

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/google/uuid"
)

// A receipt is an immutable public result. Admit its complete current shape
// and scope before replay; a matching schema token does not establish validity.
func validateTableReceipt(key authn.RouteIdempotencyKey, status int, payload map[string]any) error {
	invalid := errSavedGraphCutoverIncompatible
	if status != 200 || len(payload) != 2 || payload["schema_id"] != "cartulary.network_flow_table_mutation_result.v3" {
		return invalid
	}
	if key.RouteKey != routeKeyTablesPatch && key.RouteKey != routeKeyTablesDelete {
		return invalid
	}
	raw, err := json.Marshal(payload["table"])
	if err != nil {
		return invalid
	}
	var table struct {
		ID             string    `json:"network_flow_table_id"`
		Incident       uuid.UUID `json:"incident_id"`
		Name           string    `json:"display_name"`
		Version        int64     `json:"table_version"`
		Status         string    `json:"table_status"`
		Source         uuid.UUID `json:"committed_source_id"`
		Interpretation string    `json:"source_interpretation_sha256"`
		RawCellLimit   int64     `json:"source_raw_cell_scalar_limit"`
		SourceSHA      string    `json:"source_content_sha256"`
		Filename       string    `json:"source_filename_display"`
		FilenameDigest string    `json:"source_filename_digest"`
		FilenameKey    string    `json:"source_filename_digest_key_id"`
		Fingerprint    string    `json:"mapping_fingerprint"`
		Profile        string    `json:"source_profile_id"`
		Counters       string    `json:"counter_semantics"`
		Parser         string    `json:"parser_profile_id"`
		Accepted       int64     `json:"row_count_accepted"`
		Rejected       int64     `json:"row_count_rejected"`
		Truncated      bool      `json:"diagnostics_truncated"`
		Actor          uuid.UUID `json:"created_by_user_id"`
		Created        string    `json:"created_at"`
		Updated        string    `json:"updated_at"`
		Deleted        *string   `json:"deleted_at"`
	}
	const members = "network_flow_table_id incident_id display_name table_version table_status committed_source_id source_interpretation_sha256 source_raw_cell_scalar_limit source_content_sha256 source_filename_display source_filename_digest source_filename_digest_key_id mapping_fingerprint source_profile_id counter_semantics parser_profile_id row_count_accepted row_count_rejected diagnostics_truncated created_by_user_id created_at updated_at deleted_at"
	if _, err := decodeClosedImportObject(raw, members, "", "deleted_at", &table); err != nil {
		return invalid
	}
	name, err := normalizeExplicitDisplayName(table.Name)
	profile, ok := mappingSourceProfile(table.Profile)
	if err != nil || name != table.Name || !ok || table.Parser != profile.ParserProfileID || table.Counters != counterSemanticsForProfile(table.Profile) || !linkTableIDPattern.MatchString(table.ID) || table.Incident == uuid.Nil || table.Source == uuid.Nil || !validLowerSHA256(table.Interpretation) || table.RawCellLimit < 1 || table.RawCellLimit > 16384 || table.Actor == uuid.Nil || table.Version < 1 || table.Accepted < 0 || table.Rejected < 0 || !validLowerSHA256(table.SourceSHA) || !validLowerSHA256(table.FilenameDigest) || !validLowerSHA256(table.Fingerprint) || !safeKeyIDPattern.MatchString(table.FilenameKey) || utf8.RuneCountInString(table.Filename) > 256 {
		return invalid
	}
	if key.ScopeKey != table.Incident.String()+":"+table.ID {
		return invalid
	}
	parse := func(value string) (time.Time, error) {
		t, e := time.Parse(time.RFC3339Nano, value)
		if e == nil && (!strings.HasSuffix(value, "Z") || timestamp(t) != value) {
			return time.Time{}, invalid
		}
		return t, e
	}
	created, e := parse(table.Created)
	if e != nil {
		return invalid
	}
	updated, e := parse(table.Updated)
	if e != nil || updated.Before(created) {
		return invalid
	}
	switch table.Status {
	case "active":
		if table.Deleted != nil || key.RouteKey == routeKeyTablesDelete {
			return invalid
		}
	case "soft_deleted":
		if table.Deleted == nil {
			return invalid
		}
		deleted, e := parse(*table.Deleted)
		if e != nil || deleted.Before(created) || deleted.After(updated) {
			return invalid
		}
	default:
		return invalid
	}
	return nil
}
