package sourcecontract

import (
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

type StorageMapping struct {
	Table  string
	Column string
}

// WritableDirectStorageMappings is the exact reviewed identifier allowlist for
// artifact scalar mutation. Callers receive a copy.
func WritableDirectStorageMappings() map[string]StorageMapping {
	return map[string]StorageMapping{
		"note.title":                          {Table: "artifacts", Column: "title"},
		"note.body":                           {Table: "artifacts", Column: "body"},
		"comm_log.timestamp_utc":              {Table: "artifacts", Column: "timestamp_utc"},
		"comm_log.comm_type":                  {Table: "artifacts", Column: "comm_type"},
		"comm_log.audience":                   {Table: "artifacts", Column: "audience"},
		"comm_log.channel_or_meeting":         {Table: "artifacts", Column: "channel_or_meeting"},
		"comm_log.summary":                    {Table: "artifacts", Column: "summary"},
		"comm_log.next_report_at":             {Table: "artifacts", Column: "next_report_at"},
		"comm_log.privilege_tag":              {Table: "artifacts", Column: "privilege_tag"},
		"handoff.timestamp_utc":               {Table: "artifacts", Column: "timestamp_utc"},
		"handoff.outgoing_owner_user_id":      {Table: "artifacts", Column: "outgoing_owner_user_id"},
		"handoff.incoming_owner_user_id":      {Table: "artifacts", Column: "incoming_owner_user_id"},
		"handoff.current_state_summary":       {Table: "artifacts", Column: "current_state_summary"},
		"handoff.next_checks":                 {Table: "artifacts", Column: "next_checks"},
		"handoff.acknowledged_at":             {Table: "artifacts", Column: "acknowledged_at"},
		"status_review.timestamp_utc":         {Table: "artifacts", Column: "timestamp_utc"},
		"status_review.review_owner_user_id":  {Table: "artifacts", Column: "review_owner_user_id"},
		"status_review.current_state_summary": {Table: "artifacts", Column: "current_state_summary"},
		"status_review.active_risks_summary":  {Table: "artifacts", Column: "active_risks_summary"},
		"status_review.next_report_at":        {Table: "artifacts", Column: "next_report_at"},
		"lesson.timestamp_utc":                {Table: "artifacts", Column: "timestamp_utc"},
		"lesson.summary":                      {Table: "artifacts", Column: "summary"},
		"lesson.owner_user_id":                {Table: "artifacts", Column: "owner_user_id"},
		"lesson.closure_state":                {Table: "artifacts", Column: "closure_state"},
		"finding.statement":                   {Table: "artifact_findings", Column: "statement"},
		"finding.kind":                        {Table: "artifact_findings", Column: "kind"},
		"finding.state":                       {Table: "artifact_findings", Column: "state"},
		"finding.owner_user_id":               {Table: "artifact_findings", Column: "owner_user_id"},
		"finding.confidence_score":            {Table: "artifact_findings", Column: "confidence_score"},
		"investigative_query.platform":        {Table: "artifact_investigative_queries", Column: "platform"},
		"investigative_query.purpose":         {Table: "artifact_investigative_queries", Column: "purpose"},
		"investigative_query.query_text":      {Table: "artifact_investigative_queries", Column: "query_text"},
		"forensic_keyword.pattern":            {Table: "artifact_forensic_keywords", Column: "pattern"},
		"forensic_keyword.reason":             {Table: "artifact_forensic_keywords", Column: "reason"},
		"forensic_keyword.match_mode":         {Table: "artifact_forensic_keywords", Column: "match_mode"},
		"forensic_keyword.case_sensitive":     {Table: "artifact_forensic_keywords", Column: "case_sensitive"},
	}
}

// ConflictFieldSourceKeys declares the source-owned scalar projection required
// to reconstruct optimistic-concurrency facts from canonical snapshots.
func ConflictFieldSourceKeys() map[string]string {
	result := make(map[string]string)
	for fieldKey, storage := range WritableDirectStorageMappings() {
		result[fieldKey] = storage.Column
	}
	return result
}

// ExtractRollbackSource accepts only the source member of a canonical retained
// snapshot envelope.
func ExtractRollbackSource(value map[string]any) (map[string]any, bool) {
	if source, ok := objectMap(value, "source"); ok {
		return source, len(source) > 0
	}
	return nil, false
}

// ValidRollbackSource rejects invalid typed values and subtype invariants
// before a rollback adapter executes source SQL.
func ValidRollbackSource(source map[string]any) bool {
	for _, key := range []string{
		"title", "body", "comm_id", "comm_type", "audience", "channel_or_meeting", "summary",
		"privilege_tag", "handoff_id", "current_state_summary", "next_checks", "status_review_id",
		"active_risks_summary", "lesson_id", "closure_state", "kind", "statement", "state",
		"query_id", "platform", "purpose", "query_text", "keyword_id", "pattern", "reason", "match_mode",
	} {
		if raw, present := source[key]; present && raw != nil {
			if _, valid := raw.(string); !valid {
				return false
			}
		}
	}
	for _, key := range []string{
		"comm_id", "comm_type", "audience", "channel_or_meeting", "summary", "handoff_id",
		"current_state_summary", "status_review_id", "lesson_id", "kind", "statement", "state",
		"query_id", "platform", "purpose", "query_text", "keyword_id", "pattern", "reason", "match_mode",
	} {
		if raw, present := source[key]; present && !nonEmptyText(raw) {
			return false
		}
	}
	if raw, present := source["comm_type"]; present && !oneOfText(raw, "meeting", "notification", "approval", "briefing", "handoff") {
		return false
	}
	if raw, present := source["closure_state"]; present && !oneOfText(raw, "open", "closed") {
		return false
	}
	if raw, present := source["kind"]; present && !oneOfText(raw, "finding", "hypothesis") {
		return false
	}
	if raw, present := source["state"]; present && !oneOfText(raw, "open", "closed") {
		return false
	}
	if raw, present := source["match_mode"]; present && !oneOfText(raw, "literal", "regex") {
		return false
	}
	if raw, present := source["confidence_score"]; present && raw != nil {
		score, valid := integerValue(raw)
		if !valid || score < 0 || score > 100 {
			return false
		}
	}
	for key, kind := range map[string]string{
		"timestamp_utc": "time", "next_report_at": "time", "acknowledged_at": "time", "closed_at": "time",
		"outgoing_owner_user_id": "uuid", "incoming_owner_user_id": "uuid",
		"review_owner_user_id": "uuid", "owner_user_id": "uuid", "case_sensitive": "bool",
	} {
		if raw, present := source[key]; present && raw != nil && !validKind(raw, kind) {
			return false
		}
	}
	for _, key := range []string{"timestamp_utc", "incoming_owner_user_id", "review_owner_user_id", "owner_user_id", "case_sensitive"} {
		if raw, present := source[key]; present && raw == nil {
			return false
		}
	}
	return true
}

func validKind(value any, kind string) bool {
	switch kind {
	case "time":
		switch typed := value.(type) {
		case time.Time:
			return true
		case string:
			_, err := time.Parse(time.RFC3339Nano, typed)
			return err == nil
		}
	case "uuid":
		text, ok := value.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return false
		}
		_, err := uuid.Parse(text)
		return err == nil
	case "bool":
		_, ok := value.(bool)
		return ok
	}
	return false
}

func objectMap(value map[string]any, key string) (map[string]any, bool) {
	raw, ok := value[key]
	if !ok || raw == nil {
		return nil, false
	}
	typed, ok := raw.(map[string]any)
	return typed, ok
}

func integerValue(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), math.Trunc(typed) == typed
	default:
		return 0, false
	}
}

func oneOfText(value any, allowed ...string) bool {
	text, valid := value.(string)
	if !valid {
		return false
	}
	for _, candidate := range allowed {
		if text == candidate {
			return true
		}
	}
	return false
}

func nonEmptyText(value any) bool {
	text, valid := value.(string)
	return valid && strings.TrimSpace(text) != ""
}
