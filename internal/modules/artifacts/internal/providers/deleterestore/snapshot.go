package deleterestore

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/sourcecatalog"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/textpolicy"
	shape "github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func ValidateSnapshotSource(record, source map[string]any) error {
	parent := maps.Clone(source)
	for _, name := range []string{"finding", "investigative_query", "forensic_keyword"} {
		if _, present := source[name]; !present {
			return fmt.Errorf("missing artifact snapshot child %s", name)
		}
		delete(parent, name)
	}
	if err := shape.ValidateSourceObject(record, parent,
		shape.SourceFieldGroup{Names: "record_id incident_id", Kind: shape.SourceUUID},
		shape.SourceFieldGroup{Names: "artifact_type", Kind: shape.SourceString},
		shape.SourceFieldGroup{Names: "title body comm_id comm_type audience channel_or_meeting summary privilege_tag handoff_id current_state_summary next_checks status_review_id active_risks_summary lesson_id closure_state", Kind: shape.SourceString, Nullable: true},
		shape.SourceFieldGroup{Names: "timestamp_utc next_report_at acknowledged_at", Kind: shape.SourceTimestamp, Nullable: true},
		shape.SourceFieldGroup{Names: "updated_at created_at", Kind: shape.SourceTimestamp},
		shape.SourceFieldGroup{Names: "outgoing_owner_user_id incoming_owner_user_id review_owner_user_id owner_user_id created_by_user_id", Kind: shape.SourceUUID, Nullable: true},
	); err != nil {
		return err
	}
	typ := parent["artifact_type"].(string)
	active := parent
	childName := ""
	var childFields []shape.SourceFieldGroup
	switch typ {
	case "note", "comm_log", "handoff", "status_review", "lesson":
	case "finding":
		childName = "finding"
		childFields = []shape.SourceFieldGroup{
			{Names: "kind statement state", Kind: shape.SourceString},
			{Names: "owner_user_id", Kind: shape.SourceUUID},
			{Names: "confidence_score", Kind: shape.SourceInteger, Nullable: true},
			{Names: "closed_at", Kind: shape.SourceTimestamp, Nullable: true},
		}
	case "investigative_query":
		childName = "investigative_query"
		childFields = []shape.SourceFieldGroup{
			{Names: "query_id platform purpose query_text", Kind: shape.SourceString},
			{Names: "created_by_user_id", Kind: shape.SourceUUID},
		}
	case "forensic_keyword":
		childName = "forensic_keyword"
		childFields = []shape.SourceFieldGroup{
			{Names: "keyword_id pattern reason match_mode", Kind: shape.SourceString},
			{Names: "case_sensitive", Kind: shape.SourceBoolean},
		}
	default:
		return fmt.Errorf("invalid artifact snapshot type")
	}
	for _, name := range []string{"finding", "investigative_query", "forensic_keyword"} {
		if name != childName {
			if source[name] != nil {
				return fmt.Errorf("unexpected artifact snapshot child %s", name)
			}
			continue
		}
		child, ok := source[name].(map[string]any)
		if !ok || child == nil {
			return fmt.Errorf("missing artifact snapshot child row %s", name)
		}
		fields := append([]shape.SourceFieldGroup{
			{Names: "record_id incident_id", Kind: shape.SourceUUID},
			{Names: "created_at updated_at", Kind: shape.SourceTimestamp},
		}, childFields...)
		if err := shape.ValidateSourceObject(record, child, fields...); err != nil {
			return err
		}
		active = child
	}
	if err := validateRowClocksAndText(parent); err != nil {
		return err
	}
	if err := validateIntrinsicSnapshot(active, typ); err != nil {
		return err
	}
	return validateStoredText(source)
}

func validateStoredText(source map[string]any) error {
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return err
	}
	surface, ok := catalog.SurfaceByArtifactType(source["artifact_type"].(string))
	if !ok {
		return fmt.Errorf("missing historical artifact surface")
	}
	for _, mapping := range catalog.Fields() {
		if mapping.ViewSchemaID != surface.ViewSchemaID || mapping.Kind != sourcecatalog.FieldKindDirect {
			continue
		}
		field, ok := viewschema.LookupField(mapping.ViewSchemaID, mapping.FieldKey)
		if !ok {
			return fmt.Errorf("missing historical artifact field binding")
		}
		valueSource := source
		path := mapping.Storage.SnapshotPath()
		if len(path) == 2 {
			valueSource, _ = source[path[0]].(map[string]any)
		}
		if field.StringContractID == nil || valueSource[mapping.Storage.Column] == nil {
			continue // Required presence was checked by the retained subtype policy.
		}
		value, ok := valueSource[mapping.Storage.Column].(string)
		if !ok {
			return fmt.Errorf("invalid historical artifact text shape")
		}
		normalized, valid := textpolicy.Normalize(value, *field.StringContractID)
		if !valid || normalized == "" || normalized != value {
			return fmt.Errorf("invalid historical artifact text %s", mapping.Storage.Column)
		}
	}
	return nil
}

// Shape admission precedes these checks. They concern the retained subtype,
// without applying create-only minimum signals or consulting current actors.
func validateIntrinsicSnapshot(source map[string]any, artifactType string) error {
	requiredText, requiredValues := "", ""
	switch artifactType {
	case "note":
	case "comm_log":
		requiredText, requiredValues = "comm_id comm_type audience channel_or_meeting summary", "timestamp_utc"
		if !snapshotToken(source["comm_type"], "meeting", "notification", "approval", "briefing", "handoff") {
			return fmt.Errorf("invalid historical communication type")
		}
	case "handoff":
		requiredText, requiredValues = "handoff_id current_state_summary", "timestamp_utc outgoing_owner_user_id incoming_owner_user_id"
	case "status_review":
		requiredText, requiredValues = "status_review_id current_state_summary", "timestamp_utc review_owner_user_id"
	case "lesson":
		requiredText, requiredValues = "lesson_id summary", "timestamp_utc owner_user_id"
		if source["closure_state"] != nil && !snapshotToken(source["closure_state"], "open", "closed") {
			return fmt.Errorf("invalid historical lesson closure state")
		}
	case "finding":
		requiredText = "kind statement state"
		if !snapshotToken(source["kind"], "finding", "hypothesis") || !snapshotToken(source["state"], "open", "closed") {
			return fmt.Errorf("invalid historical finding vocabulary")
		}
		if number := source["confidence_score"]; number != nil {
			value, err := number.(json.Number).Int64()
			if err != nil || value < 0 || value > 100 {
				return fmt.Errorf("invalid historical finding confidence")
			}
		}
		if (source["state"] == "closed") != (source["closed_at"] != nil) {
			return fmt.Errorf("invalid historical finding closure tuple")
		}
	case "investigative_query":
		requiredText, requiredValues = "query_id platform purpose query_text", "created_by_user_id"
	case "forensic_keyword":
		requiredText = "keyword_id pattern reason match_mode"
		if !snapshotToken(source["match_mode"], "literal", "regex") {
			return fmt.Errorf("invalid historical keyword match mode")
		}
	}
	for _, field := range strings.Fields(requiredText) {
		value, ok := source[field].(string)
		if !ok || strings.TrimSpace(value) == "" {
			return fmt.Errorf("missing historical artifact text %s", field)
		}
	}
	for _, field := range strings.Fields(requiredValues) {
		if source[field] == nil {
			return fmt.Errorf("missing historical artifact value %s", field)
		}
	}
	return validateRowClocksAndText(source)
}

func validateRowClocksAndText(source map[string]any) error {
	// PostgreSQL source text cannot contain NUL; accepting it only in retained
	// history would describe a state that the source owner cannot persist.
	for _, value := range source {
		if text, ok := value.(string); ok && strings.ContainsRune(text, 0) {
			return fmt.Errorf("invalid historical artifact text")
		}
	}
	created, _ := time.Parse(time.RFC3339Nano, source["created_at"].(string))
	updated, _ := time.Parse(time.RFC3339Nano, source["updated_at"].(string))
	if updated.Before(created) {
		return fmt.Errorf("historical artifact clocks reversed")
	}
	if raw, ok := source["closed_at"].(string); ok {
		closed, _ := time.Parse(time.RFC3339Nano, raw)
		if closed.Before(created) || closed.After(updated) {
			return fmt.Errorf("historical finding closure outside source lifetime")
		}
	}
	return nil
}
func snapshotToken(raw any, allowed ...string) bool {
	value, ok := raw.(string)
	if !ok {
		return false
	}
	for _, token := range allowed {
		if value == token {
			return true
		}
	}
	return false
}
