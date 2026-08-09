package queryengine

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
)

const (
	artifactNotesViewSchemaID                = "cartulary.view.notes.v1"
	artifactCommLogViewSchemaID              = "cartulary.view.comm_log.v1"
	artifactHandoffViewSchemaID              = "cartulary.view.handoff.v1"
	artifactStatusReviewViewSchemaID         = "cartulary.view.status_review.v1"
	artifactLessonViewSchemaID               = "cartulary.view.lesson.v1"
	artifactFindingsViewSchemaID             = "cartulary.view.findings.v1"
	artifactInvestigativeQueriesViewSchemaID = "cartulary.view.investigative_queries.v1"
	artifactForensicKeywordsViewSchemaID     = "cartulary.view.forensic_keywords.v1"
)

type ArtifactReader struct{}

func NewArtifactReader() *ArtifactReader { return &ArtifactReader{} }

func (*ArtifactReader) CollectDerivedFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]artifactprojection.DerivedFact, error) {
	rows, err := tx.Query(ctx, `
SELECT a.record_id, a.artifact_type, a.finding_kind, to_jsonb(a) - 'incident_id'
  FROM artifact_grid_projection a
  JOIN records r
    ON r.incident_id = a.incident_id
   AND r.record_id = a.record_id
   AND r.deleted_at IS NULL
 WHERE a.incident_id = $1
 ORDER BY a.record_id
`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("collect Artifact projection facts: %w", err)
	}
	defer rows.Close()

	facts := make([]artifactprojection.DerivedFact, 0)
	for rows.Next() {
		var (
			fact        artifactprojection.DerivedFact
			findingKind pgtype.Text
			raw         []byte
		)
		if err := rows.Scan(&fact.RecordID, &fact.ArtifactType, &findingKind, &raw); err != nil {
			return nil, fmt.Errorf("scan Artifact projection fact: %w", err)
		}
		if findingKind.Valid {
			value := findingKind.String
			fact.FindingKind = &value
		}
		fact.Value = map[string]any{}
		if err := json.Unmarshal(raw, &fact.Value); err != nil {
			return nil, fmt.Errorf("decode Artifact projection fact: %w", err)
		}
		facts = append(facts, fact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate Artifact projection facts: %w", err)
	}
	return facts, nil
}

func ArtifactPlans() []Surface {
	return []Surface{
		artifactSurface(artifactFindingsViewSchemaID, "finding", []Field{
			{Key: "finding.statement", Expr: "p.finding_statement", Kind: FieldKindText},
			{Key: "finding.kind", Expr: "p.finding_kind", SortExpr: enumSortExpr("p.finding_kind", "finding", "hypothesis"), Kind: FieldKindText},
			{Key: "finding.state", Expr: "p.finding_state", SortExpr: enumSortExpr("p.finding_state", "open", "closed"), Kind: FieldKindText},
			{Key: "finding.owner_user_id", Expr: "p.finding_owner_user_id", Kind: FieldKindText},
			{Key: "finding.confidence_score", Expr: "p.finding_confidence_score", Kind: FieldKindNumber},
			{Key: "finding.closed_at", Expr: "p.finding_closed_at", Kind: FieldKindTimestamp},
			{Key: "finding.updated_at", Expr: "p.finding_updated_at", Kind: FieldKindTimestamp},
			{Key: "finding.supporting_refs", Expr: recordRefCollectionExprFor("p", "finding.supporting_refs", "supported_by"), Kind: FieldKindCollection},
			{Key: "finding.contradictory_refs", Expr: recordRefCollectionExprFor("p", "finding.contradictory_refs", "references_record"), Kind: FieldKindCollection},
			{Key: "finding.confidence_band", Expr: "p.finding_confidence_band", SortExpr: enumSortExpr("p.finding_confidence_band", "unset", "low", "medium", "high"), Kind: FieldKindText},
		}),
		artifactSurface(artifactForensicKeywordsViewSchemaID, "forensic_keyword", []Field{
			{Key: "forensic_keyword.pattern", Expr: "p.forensic_keyword_pattern", Kind: FieldKindText},
			{Key: "forensic_keyword.reason", Expr: "p.forensic_keyword_reason", Kind: FieldKindText},
			{Key: "forensic_keyword.match_mode", Expr: "p.forensic_keyword_match_mode", SortExpr: enumSortExpr("p.forensic_keyword_match_mode", "literal", "regex"), Kind: FieldKindText},
			{Key: "forensic_keyword.case_sensitive", Expr: "p.forensic_keyword_case_sensitive", Kind: FieldKindBool},
			{Key: "forensic_keyword.created_at", Expr: "p.forensic_keyword_created_at", Kind: FieldKindTimestamp},
			{Key: "forensic_keyword.keyword_id", Expr: "p.forensic_keyword_keyword_id", Kind: FieldKindText},
			{Key: "forensic_keyword.created_day", Expr: "p.forensic_keyword_created_day", Kind: FieldKindDate},
		}),
		artifactSurface(artifactInvestigativeQueriesViewSchemaID, "investigative_query", []Field{
			{Key: "investigative_query.platform", Expr: "p.investigative_query_platform", Kind: FieldKindText},
			{Key: "investigative_query.purpose", Expr: "p.investigative_query_purpose", Kind: FieldKindText},
			{Key: "investigative_query.query_text", Expr: "p.investigative_query_query_text", Kind: FieldKindText},
			{Key: "investigative_query.created_by_user_id", Expr: "p.investigative_query_created_by_user_id", Kind: FieldKindText},
			{Key: "investigative_query.created_at", Expr: "p.investigative_query_created_at", Kind: FieldKindTimestamp},
			{Key: "investigative_query.query_id", Expr: "p.investigative_query_query_id", Kind: FieldKindText},
			{Key: "investigative_query.created_day", Expr: "p.investigative_query_created_day", Kind: FieldKindDate},
		}),
		artifactSurface(artifactNotesViewSchemaID, "note", []Field{
			{Key: "note.title", Expr: "p.title", Kind: FieldKindText},
			{Key: "note.body", Expr: "p.body", Kind: FieldKindText},
			{Key: "note.tags", Expr: tagCollectionExprFor("p"), Kind: FieldKindCollection},
			{Key: "note.linked_record_count", Expr: "p.linked_record_count", Kind: FieldKindNumber},
			{Key: "note.updated_at", Expr: "p.updated_at", Kind: FieldKindTimestamp},
			{Key: "note.created_by_user_id", Expr: "p.created_by_user_id", Kind: FieldKindText},
		}),
		artifactSurface(artifactCommLogViewSchemaID, "comm_log", []Field{
			{Key: "comm_log.timestamp_utc", Expr: "p.timestamp_utc", Kind: FieldKindTimestamp},
			{Key: "comm_log.comm_type", Expr: "p.comm_type", Kind: FieldKindText},
			{Key: "comm_log.audience", Expr: "p.audience", Kind: FieldKindText},
			{Key: "comm_log.channel_or_meeting", Expr: "p.channel_or_meeting", Kind: FieldKindText},
			{Key: "comm_log.summary", Expr: "p.summary", Kind: FieldKindText},
			{Key: "comm_log.next_report_at", Expr: "p.next_report_at", Kind: FieldKindTimestamp},
			{Key: "comm_log.privilege_tag", Expr: "p.privilege_tag", Kind: FieldKindText},
			{Key: "comm_log.decision_ids", Expr: recordRefCollectionExpr("comm_log.decision_ids"), Kind: FieldKindCollection},
			{Key: "comm_log.action_task_ids", Expr: recordRefCollectionExpr("comm_log.action_task_ids"), Kind: FieldKindCollection},
			{Key: "comm_log.audience_party_ids", Expr: partyRefCollectionExpr("comm_log.audience_party_ids"), Kind: FieldKindCollection},
			{Key: "comm_log.attendee_party_ids", Expr: partyRefCollectionExpr("comm_log.attendee_party_ids"), Kind: FieldKindCollection},
			{Key: "comm_log.comm_id", Expr: "p.comm_id", Kind: FieldKindText},
			{Key: "comm_log.timestamp_day", Expr: "p.timestamp_day", Kind: FieldKindDate},
			{Key: "comm_log.next_report_day", Expr: "p.next_report_day", Kind: FieldKindDate},
			{Key: "comm_log.updated_at", Expr: "p.updated_at", Kind: FieldKindTimestamp},
		}),
		artifactSurface(artifactHandoffViewSchemaID, "handoff", []Field{
			{Key: "handoff.timestamp_utc", Expr: "p.timestamp_utc", Kind: FieldKindTimestamp},
			{Key: "handoff.outgoing_owner_user_id", Expr: "p.outgoing_owner_user_id", Kind: FieldKindText},
			{Key: "handoff.incoming_owner_user_id", Expr: "p.incoming_owner_user_id", Kind: FieldKindText},
			{Key: "handoff.current_state_summary", Expr: "p.current_state_summary", Kind: FieldKindText},
			{Key: "handoff.open_task_ids", Expr: recordRefCollectionExpr("handoff.open_task_ids"), Kind: FieldKindCollection},
			{Key: "handoff.open_decision_ids", Expr: recordRefCollectionExpr("handoff.open_decision_ids"), Kind: FieldKindCollection},
			{Key: "handoff.open_risk_refs", Expr: riskRefCollectionExpr(), Kind: FieldKindCollection},
			{Key: "handoff.next_checks", Expr: "p.next_checks", Kind: FieldKindText},
			{Key: "handoff.acknowledged_at", Expr: "p.acknowledged_at", Kind: FieldKindTimestamp},
			{Key: "handoff.handoff_id", Expr: "p.handoff_id", Kind: FieldKindText},
			{Key: "handoff.timestamp_day", Expr: "p.timestamp_day", Kind: FieldKindDate},
			{Key: "handoff.ack_state", Expr: "p.ack_state", SortExpr: enumSortExpr("p.ack_state", "pending", "acknowledged"), Kind: FieldKindText},
			{Key: "handoff.updated_at", Expr: "p.updated_at", Kind: FieldKindTimestamp},
		}),
		artifactSurface(artifactStatusReviewViewSchemaID, "status_review", []Field{
			{Key: "status_review.timestamp_utc", Expr: "p.timestamp_utc", Kind: FieldKindTimestamp},
			{Key: "status_review.review_owner_user_id", Expr: "p.review_owner_user_id", Kind: FieldKindText},
			{Key: "status_review.current_state_summary", Expr: "p.current_state_summary", Kind: FieldKindText},
			{Key: "status_review.blocked_task_ids", Expr: recordRefCollectionExpr("status_review.blocked_task_ids"), Kind: FieldKindCollection},
			{Key: "status_review.pending_evidence_ids", Expr: recordRefCollectionExpr("status_review.pending_evidence_ids"), Kind: FieldKindCollection},
			{Key: "status_review.open_decision_ids", Expr: recordRefCollectionExpr("status_review.open_decision_ids"), Kind: FieldKindCollection},
			{Key: "status_review.active_risks_summary", Expr: "p.active_risks_summary", Kind: FieldKindText},
			{Key: "status_review.next_report_at", Expr: "p.next_report_at", Kind: FieldKindTimestamp},
			{Key: "status_review.status_review_id", Expr: "p.status_review_id", Kind: FieldKindText},
			{Key: "status_review.timestamp_day", Expr: "p.timestamp_day", Kind: FieldKindDate},
			{Key: "status_review.next_report_day", Expr: "p.next_report_day", Kind: FieldKindDate},
			{Key: "status_review.updated_at", Expr: "p.updated_at", Kind: FieldKindTimestamp},
		}),
		artifactSurface(artifactLessonViewSchemaID, "lesson", []Field{
			{Key: "lesson.timestamp_utc", Expr: "p.timestamp_utc", Kind: FieldKindTimestamp},
			{Key: "lesson.summary", Expr: "p.summary", Kind: FieldKindText},
			{Key: "lesson.owner_user_id", Expr: "p.owner_user_id", Kind: FieldKindText},
			{Key: "lesson.closure_state", Expr: "p.closure_state", Kind: FieldKindText},
			{Key: "lesson.follow_up_task_ids", Expr: recordRefCollectionExpr("lesson.follow_up_task_ids"), Kind: FieldKindCollection},
			{Key: "lesson.evidence_refs", Expr: recordRefCollectionExpr("lesson.evidence_refs"), Kind: FieldKindCollection},
			{Key: "lesson.lesson_id", Expr: "p.lesson_id", Kind: FieldKindText},
			{Key: "lesson.timestamp_day", Expr: "p.timestamp_day", Kind: FieldKindDate},
			{Key: "lesson.updated_at", Expr: "p.updated_at", Kind: FieldKindTimestamp},
		}),
	}
}

func artifactSurface(viewSchemaID string, artifactType string, fields []Field) Surface {
	return Surface{
		ViewSchemaID: viewSchemaID,
		FromSQL:      "FROM artifact_grid_projection p JOIN records r ON r.record_id = p.record_id",
		RecordExpr:   "p.record_id",
		IncidentExpr: "p.incident_id",
		WhereSQL:     "p.artifact_type = '" + artifactType + "'",
		Fields:       fields,
	}
}

func enumSortExpr(expr string, values ...string) string {
	var builder strings.Builder
	builder.WriteString("CASE ")
	builder.WriteString(expr)
	for index, value := range values {
		builder.WriteString(" WHEN '")
		builder.WriteString(value)
		builder.WriteString("' THEN ")
		builder.WriteString(strconv.Itoa(index))
	}
	builder.WriteString(" ELSE ")
	builder.WriteString(strconv.Itoa(len(values)))
	builder.WriteString(" END")
	return builder.String()
}

func recordRefCollectionExpr(fieldKey string) string {
	return recordRefCollectionExprFor("p", fieldKey, "references_record")
}

func recordRefCollectionExprFor(alias string, fieldKey string, linkType string) string {
	return `(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', ` + recordRefItemRefSQL("dst.record_id") + `,
        'item_kind', 'record_ref',
        'display_text', dst.record_type || ':' || dst.record_id::text,
        'linked_record_id', dst.record_id::text
    ) ORDER BY dst.record_type ASC, dst.record_id ASC), '[]'::jsonb)
      FROM ` + activeRecordLinksAlias("rl") + `
      JOIN records dst
        ON dst.incident_id = rl.incident_id
       AND dst.record_id = rl.dst_record_id
       AND dst.deleted_at IS NULL
     WHERE rl.incident_id = ` + alias + `.incident_id
       AND rl.src_record_id = ` + alias + `.record_id
       AND rl.link_type = '` + linkType + `'
       AND rl.field_key = '` + fieldKey + `')::text`
}

func tagCollectionExprFor(alias string) string {
	return `(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', ` + recordTagItemRefSQL("rt.record_id", "rt.record_tag_id") + `,
        'item_kind', 'tag',
        'display_text', rt.tag_name,
        'tag_id', rt.record_tag_id::text
    ) ORDER BY rt.normalized_tag_name ASC, rt.record_tag_id ASC), '[]'::jsonb)
      FROM ` + activeRecordTagsAlias("rt") + `
     WHERE rt.incident_id = ` + alias + `.incident_id
       AND rt.record_id = ` + alias + `.record_id)::text`
}

func partyRefCollectionExpr(fieldKey string) string {
	return `(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', ` + partyRefItemRefSQL("party.record_id") + `,
        'item_kind', 'party_ref',
        'display_text', party.display_name,
        'party_id', party.record_id::text
    ) ORDER BY party.display_name ASC, party.record_id ASC), '[]'::jsonb)
      FROM ` + activeRecordLinksAlias("rl") + `
      JOIN parties party
        ON party.incident_id = rl.incident_id
       AND party.record_id = rl.dst_record_id
      JOIN records dst
        ON dst.incident_id = rl.incident_id
       AND dst.record_id = rl.dst_record_id
       AND dst.deleted_at IS NULL
     WHERE rl.incident_id = p.incident_id
       AND rl.src_record_id = p.record_id
       AND rl.link_type = 'references_record'
       AND rl.field_key = '` + fieldKey + `')::text`
}

func riskRefCollectionExpr() string {
	return `(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', 'risk_ref:' || risk_ref_id::text,
        'item_kind', 'risk_ref',
        'display_text', risk_ref_text,
        'risk_ref_id', risk_ref_id::text,
        'risk_ref_text', risk_ref_text
    ) ORDER BY risk_ref_text ASC, risk_ref_id ASC), '[]'::jsonb)
      FROM handoff_risk_refs hr
     WHERE hr.incident_id = p.incident_id
       AND hr.handoff_record_id = p.record_id
       AND hr.deleted_at IS NULL)::text`
}
