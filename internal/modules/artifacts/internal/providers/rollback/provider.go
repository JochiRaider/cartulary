package rollback

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type Provider struct{}

var _ rollbackcontract.RowSourceProvider = Provider{}

func NewProvider() Provider { return Provider{} }

func (Provider) ValidateRollbackValue(value map[string]any) error {
	source, ok := extractRollbackSource(value)
	if !ok || !validRollbackSource(source) {
		return rollbackcontract.ErrTargetNotReversible
	}
	return nil
}

func (Provider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok := extractRollbackSource(request.RetainedValue)
	if !ok || !validRollbackSource(source) {
		return rollbackcontract.ErrTargetNotReversible
	}
	var artifactType string
	if err := tx.QueryRow(ctx, `SELECT artifact_type FROM artifacts WHERE record_id = $1`, request.RecordID).Scan(&artifactType); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rollbackcontract.ErrTargetNotFound
		}
		return err
	}
	if retainedType, present := source["artifact_type"]; present {
		text, valid := retainedType.(string)
		if !valid || text != artifactType {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	switch artifactType {
	case "note":
		return restoreNoteTx(ctx, tx, request, source)
	case "comm_log":
		return restoreCommLogTx(ctx, tx, request, source)
	case "handoff":
		return restoreHandoffTx(ctx, tx, request, source)
	case "status_review":
		return restoreStatusReviewTx(ctx, tx, request, source)
	case "lesson":
		return restoreLessonTx(ctx, tx, request, source)
	case "finding":
		return restoreFindingTx(ctx, tx, request, source)
	case "investigative_query":
		return restoreInvestigativeQueryTx(ctx, tx, request, source)
	case "forensic_keyword":
		return restoreForensicKeywordTx(ctx, tx, request, source)
	default:
		return rollbackcontract.ErrTargetNotReversible
	}
}

func restoreNoteTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest, source map[string]any) error {
	titleP, title := rawPair(source, "title")
	bodyP, body := rawPair(source, "body")
	_, err := tx.Exec(ctx, `
UPDATE artifacts
   SET title = CASE WHEN $2 THEN $3::text ELSE title END,
       body = CASE WHEN $4 THEN $5::text ELSE body END,
       updated_at = $6
 WHERE record_id = $1 AND artifact_type = 'note'
`, request.RecordID, titleP, title, bodyP, body, request.Now.UTC())
	return err
}

func restoreCommLogTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest, source map[string]any) error {
	if err := checkImmutableTextTx(ctx, tx, `SELECT comm_id FROM artifacts WHERE record_id = $1`, request.RecordID, source, "comm_id"); err != nil {
		return err
	}
	values, err := typedPairs(source, []typedField{
		{"timestamp_utc", kindTime}, {"comm_type", kindText}, {"audience", kindText},
		{"channel_or_meeting", kindText}, {"summary", kindText}, {"next_report_at", kindTime},
		{"privilege_tag", kindText},
	})
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	_, err = tx.Exec(ctx, `
UPDATE artifacts
   SET timestamp_utc = CASE WHEN $2 THEN $3::timestamptz ELSE timestamp_utc END,
       comm_type = CASE WHEN $4 THEN $5::text ELSE comm_type END,
       audience = CASE WHEN $6 THEN $7::text ELSE audience END,
       channel_or_meeting = CASE WHEN $8 THEN $9::text ELSE channel_or_meeting END,
       summary = CASE WHEN $10 THEN $11::text ELSE summary END,
       next_report_at = CASE WHEN $12 THEN $13::timestamptz ELSE next_report_at END,
       privilege_tag = CASE WHEN $14 THEN $15::text ELSE privilege_tag END,
       updated_at = $16
 WHERE record_id = $1 AND artifact_type = 'comm_log'
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	return err
}

func restoreHandoffTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest, source map[string]any) error {
	if err := checkImmutableTextTx(ctx, tx, `SELECT handoff_id FROM artifacts WHERE record_id = $1`, request.RecordID, source, "handoff_id"); err != nil {
		return err
	}
	values, err := typedPairs(source, []typedField{
		{"timestamp_utc", kindTime}, {"outgoing_owner_user_id", kindUUID}, {"incoming_owner_user_id", kindUUID},
		{"current_state_summary", kindText}, {"next_checks", kindText}, {"acknowledged_at", kindTime},
	})
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	_, err = tx.Exec(ctx, `
UPDATE artifacts
   SET timestamp_utc = CASE WHEN $2 THEN $3::timestamptz ELSE timestamp_utc END,
       outgoing_owner_user_id = CASE WHEN $4 THEN $5::uuid ELSE outgoing_owner_user_id END,
       incoming_owner_user_id = CASE WHEN $6 THEN $7::uuid ELSE incoming_owner_user_id END,
       current_state_summary = CASE WHEN $8 THEN $9::text ELSE current_state_summary END,
       next_checks = CASE WHEN $10 THEN $11::text ELSE next_checks END,
       acknowledged_at = CASE WHEN $12 THEN $13::timestamptz ELSE acknowledged_at END,
       updated_at = $14
 WHERE record_id = $1 AND artifact_type = 'handoff'
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	return err
}

func restoreStatusReviewTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest, source map[string]any) error {
	if err := checkImmutableTextTx(ctx, tx, `SELECT status_review_id FROM artifacts WHERE record_id = $1`, request.RecordID, source, "status_review_id"); err != nil {
		return err
	}
	values, err := typedPairs(source, []typedField{
		{"timestamp_utc", kindTime}, {"review_owner_user_id", kindUUID}, {"current_state_summary", kindText},
		{"active_risks_summary", kindText}, {"next_report_at", kindTime},
	})
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	_, err = tx.Exec(ctx, `
UPDATE artifacts
   SET timestamp_utc = CASE WHEN $2 THEN $3::timestamptz ELSE timestamp_utc END,
       review_owner_user_id = CASE WHEN $4 THEN $5::uuid ELSE review_owner_user_id END,
       current_state_summary = CASE WHEN $6 THEN $7::text ELSE current_state_summary END,
       active_risks_summary = CASE WHEN $8 THEN $9::text ELSE active_risks_summary END,
       next_report_at = CASE WHEN $10 THEN $11::timestamptz ELSE next_report_at END,
       updated_at = $12
 WHERE record_id = $1 AND artifact_type = 'status_review'
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	return err
}

func restoreLessonTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest, source map[string]any) error {
	if err := checkImmutableTextTx(ctx, tx, `SELECT lesson_id FROM artifacts WHERE record_id = $1`, request.RecordID, source, "lesson_id"); err != nil {
		return err
	}
	values, err := typedPairs(source, []typedField{
		{"timestamp_utc", kindTime}, {"summary", kindText}, {"owner_user_id", kindUUID}, {"closure_state", kindText},
	})
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	_, err = tx.Exec(ctx, `
UPDATE artifacts
   SET timestamp_utc = CASE WHEN $2 THEN $3::timestamptz ELSE timestamp_utc END,
       summary = CASE WHEN $4 THEN $5::text ELSE summary END,
       owner_user_id = CASE WHEN $6 THEN $7::uuid ELSE owner_user_id END,
       closure_state = CASE WHEN $8 THEN $9::text ELSE closure_state END,
       updated_at = $10
 WHERE record_id = $1 AND artifact_type = 'lesson'
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	return err
}

func restoreFindingTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest, source map[string]any) error {
	values, err := typedPairs(source, []typedField{
		{"kind", kindText}, {"statement", kindText}, {"state", kindText}, {"confidence_score", kindInteger},
		{"owner_user_id", kindUUID}, {"closed_at", kindTime},
	})
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	tag, err := tx.Exec(ctx, `
UPDATE artifact_findings
   SET kind = CASE WHEN $2 THEN $3::text ELSE kind END,
       statement = CASE WHEN $4 THEN $5::text ELSE statement END,
       state = CASE WHEN $6 THEN $7::text ELSE state END,
       confidence_score = CASE WHEN $8 THEN $9::integer ELSE confidence_score END,
       owner_user_id = CASE WHEN $10 THEN $11::uuid ELSE owner_user_id END,
       closed_at = CASE WHEN $12 THEN $13::timestamptz ELSE closed_at END,
       updated_at = $14
 WHERE record_id = $1
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return rollbackcontract.ErrTargetNotFound
	}
	_, err = tx.Exec(ctx, `UPDATE artifacts SET updated_at = $2 WHERE record_id = $1 AND artifact_type = 'finding'`, request.RecordID, request.Now.UTC())
	return err
}

func restoreInvestigativeQueryTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest, source map[string]any) error {
	if err := checkImmutableTextTx(ctx, tx, `SELECT query_id FROM artifact_investigative_queries WHERE record_id = $1`, request.RecordID, source, "query_id"); err != nil {
		return err
	}
	values, err := typedPairs(source, []typedField{{"platform", kindText}, {"purpose", kindText}, {"query_text", kindText}})
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	tag, err := tx.Exec(ctx, `
UPDATE artifact_investigative_queries
   SET platform = CASE WHEN $2 THEN $3::text ELSE platform END,
       purpose = CASE WHEN $4 THEN $5::text ELSE purpose END,
       query_text = CASE WHEN $6 THEN $7::text ELSE query_text END,
       updated_at = $8
 WHERE record_id = $1
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return rollbackcontract.ErrTargetNotFound
	}
	_, err = tx.Exec(ctx, `UPDATE artifacts SET updated_at = $2 WHERE record_id = $1 AND artifact_type = 'investigative_query'`, request.RecordID, request.Now.UTC())
	return err
}

func restoreForensicKeywordTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest, source map[string]any) error {
	if err := checkImmutableTextTx(ctx, tx, `SELECT keyword_id FROM artifact_forensic_keywords WHERE record_id = $1`, request.RecordID, source, "keyword_id"); err != nil {
		return err
	}
	values, err := typedPairs(source, []typedField{
		{"pattern", kindText}, {"reason", kindText}, {"match_mode", kindText}, {"case_sensitive", kindBool},
	})
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	tag, err := tx.Exec(ctx, `
UPDATE artifact_forensic_keywords
   SET pattern = CASE WHEN $2 THEN $3::text ELSE pattern END,
       reason = CASE WHEN $4 THEN $5::text ELSE reason END,
       match_mode = CASE WHEN $6 THEN $7::text ELSE match_mode END,
       case_sensitive = CASE WHEN $8 THEN $9::boolean ELSE case_sensitive END,
       updated_at = $10
 WHERE record_id = $1
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return rollbackcontract.ErrTargetNotFound
	}
	_, err = tx.Exec(ctx, `UPDATE artifacts SET updated_at = $2 WHERE record_id = $1 AND artifact_type = 'forensic_keyword'`, request.RecordID, request.Now.UTC())
	return err
}

type valueKind int

const (
	kindText valueKind = iota
	kindUUID
	kindTime
	kindInteger
	kindBool
)

type typedField struct {
	key  string
	kind valueKind
}

func typedPairs(source map[string]any, fields []typedField) ([]any, error) {
	values := make([]any, 0, len(fields)*2)
	for _, field := range fields {
		raw, present := source[field.key]
		var err error
		if present && raw != nil {
			switch field.kind {
			case kindText:
				if _, valid := raw.(string); !valid {
					err = errors.New("invalid text")
				}
			case kindUUID:
				text, valid := raw.(string)
				if !valid || strings.TrimSpace(text) == "" {
					err = errors.New("invalid uuid")
				} else {
					parsed, parseErr := uuid.Parse(text)
					raw, err = parsed, parseErr
				}
			case kindTime:
				switch typed := raw.(type) {
				case time.Time:
					raw = typed.UTC()
				case string:
					var parsed time.Time
					parsed, err = time.Parse(time.RFC3339Nano, typed)
					raw = parsed.UTC()
				default:
					err = errors.New("invalid timestamp")
				}
			case kindInteger:
				var valid bool
				raw, valid = integerValue(raw)
				if !valid {
					err = errors.New("invalid integer")
				}
			case kindBool:
				if _, valid := raw.(bool); !valid {
					err = errors.New("invalid bool")
				}
			}
		}
		if err != nil {
			return nil, err
		}
		values = append(values, present, raw)
	}
	return values, nil
}

func rawPair(source map[string]any, key string) (bool, any) {
	value, present := source[key]
	return present, value
}

func checkImmutableTextTx(ctx context.Context, tx pgx.Tx, query string, recordID uuid.UUID, source map[string]any, key string) error {
	retained, present := source[key]
	if !present {
		return nil
	}
	retainedText, valid := retained.(string)
	if !valid || strings.TrimSpace(retainedText) == "" {
		return rollbackcontract.ErrTargetNotReversible
	}
	var current string
	if err := tx.QueryRow(ctx, query, recordID).Scan(&current); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rollbackcontract.ErrTargetNotFound
		}
		return err
	}
	if current != retainedText {
		return rollbackcontract.ErrTargetNotReversible
	}
	return nil
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

func extractRollbackSource(value map[string]any) (map[string]any, bool) {
	source, ok := objectMap(value, "source")
	return source, ok && len(source) > 0
}

func validRollbackSource(source map[string]any) bool {
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
		if raw, present := source[key]; present && raw != nil && !validRollbackKind(raw, kind) {
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

func validRollbackKind(value any, kind string) bool {
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
