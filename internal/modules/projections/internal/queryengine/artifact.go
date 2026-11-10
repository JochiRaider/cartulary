package queryengine

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
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
SELECT a.record_id, a.artifact_type, a.finding_kind, jsonb_build_object(
       'record_id', a.record_id,
       'row_version', a.row_version,
       'artifact_type', a.artifact_type,
       'title', a.title,
       'body', a.body,
       'timestamp_utc', a.timestamp_utc,
       'updated_at', a.updated_at,
       'created_at', a.created_at,
       'created_by_user_id', a.created_by_user_id,
       'linked_record_count', a.linked_record_count
   ) || CASE a.artifact_type
     WHEN 'note' THEN '{}'::jsonb
     WHEN 'comm_log' THEN jsonb_build_object(
       'comm_id', a.comm_id,
       'comm_type', a.comm_type,
       'audience', a.audience,
       'channel_or_meeting', a.channel_or_meeting,
       'summary', a.summary,
       'next_report_at', a.next_report_at,
       'privilege_tag', a.privilege_tag
   )
     WHEN 'handoff' THEN jsonb_build_object(
       'handoff_id', a.handoff_id,
       'outgoing_owner_user_id', a.outgoing_owner_user_id,
       'incoming_owner_user_id', a.incoming_owner_user_id,
       'current_state_summary', a.current_state_summary,
       'next_checks', a.next_checks,
       'acknowledged_at', a.acknowledged_at,
       'ack_state', a.ack_state
   )
     WHEN 'status_review' THEN jsonb_build_object(
       'status_review_id', a.status_review_id,
       'review_owner_user_id', a.review_owner_user_id,
       'current_state_summary', a.current_state_summary,
       'active_risks_summary', a.active_risks_summary,
       'next_report_at', a.next_report_at
   )
     WHEN 'lesson' THEN jsonb_build_object(
       'lesson_id', a.lesson_id,
       'owner_user_id', a.owner_user_id,
       'closure_state', a.closure_state
   )
     WHEN 'finding' THEN jsonb_build_object(
       'finding_statement', a.finding_statement,
       'finding_kind', a.finding_kind,
       'finding_state', a.finding_state,
       'finding_owner_user_id', a.finding_owner_user_id,
       'finding_confidence_score', a.finding_confidence_score,
       'finding_closed_at', a.finding_closed_at,
       'finding_updated_at', a.finding_updated_at,
       'finding_confidence_band', a.finding_confidence_band
   )
     ELSE '{}'::jsonb
   END
  FROM artifact_grid_projection a
  JOIN records r
    ON r.incident_id = a.incident_id
   AND r.record_id = a.record_id
   AND r.deleted_at IS NULL
 WHERE a.incident_id = $1
   AND a.artifact_type IN ('note', 'comm_log', 'handoff', 'status_review', 'lesson', 'finding')
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
		fact.Value, err = decodeReportingFactObject(raw)
		if err != nil {
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
			{Key: "finding.statement", Expr: sqlshape.Literal("p.finding_statement"), Kind: FieldKindText},
			{Key: "finding.kind", Expr: sqlshape.Literal("p.finding_kind"), SortExpr: enumSortExpr(sqlshape.Literal("p.finding_kind"), "finding", "hypothesis"), Kind: FieldKindText},
			{Key: "finding.state", Expr: sqlshape.Literal("p.finding_state"), SortExpr: enumSortExpr(sqlshape.Literal("p.finding_state"), "open", "closed"), Kind: FieldKindText},
			{Key: "finding.owner_user_id", Expr: sqlshape.Literal("p.finding_owner_user_id"), Kind: FieldKindText},
			{Key: "finding.confidence_score", Expr: sqlshape.Literal("p.finding_confidence_score"), Kind: FieldKindNumber},
			{Key: "finding.closed_at", Expr: sqlshape.Literal("p.finding_closed_at"), Kind: FieldKindTimestamp},
			{Key: "finding.updated_at", Expr: sqlshape.Literal("p.finding_updated_at"), Kind: FieldKindTimestamp},
			{Key: "finding.supporting_refs", Expr: recordRefCollectionExprFor("finding.supporting_refs", "supported_by"), Kind: FieldKindCollection},
			{Key: "finding.contradictory_refs", Expr: recordRefCollectionExprFor("finding.contradictory_refs", "references_record"), Kind: FieldKindCollection},
			{Key: "finding.confidence_band", Expr: sqlshape.Literal("p.finding_confidence_band"), SortExpr: enumSortExpr(sqlshape.Literal("p.finding_confidence_band"), "unset", "low", "medium", "high"), Kind: FieldKindText},
		}),
		artifactSurface(artifactForensicKeywordsViewSchemaID, "forensic_keyword", []Field{
			{Key: "forensic_keyword.pattern", Expr: sqlshape.Literal("p.forensic_keyword_pattern"), Kind: FieldKindText},
			{Key: "forensic_keyword.reason", Expr: sqlshape.Literal("p.forensic_keyword_reason"), Kind: FieldKindText},
			{Key: "forensic_keyword.match_mode", Expr: sqlshape.Literal("p.forensic_keyword_match_mode"), SortExpr: enumSortExpr(sqlshape.Literal("p.forensic_keyword_match_mode"), "literal", "regex"), Kind: FieldKindText},
			{Key: "forensic_keyword.case_sensitive", Expr: sqlshape.Literal("p.forensic_keyword_case_sensitive"), Kind: FieldKindBool},
			{Key: "forensic_keyword.created_at", Expr: sqlshape.Literal("p.forensic_keyword_created_at"), Kind: FieldKindTimestamp},
			{Key: "forensic_keyword.keyword_id", Expr: sqlshape.Literal("p.forensic_keyword_keyword_id"), Kind: FieldKindText},
			{Key: "forensic_keyword.created_day", Expr: sqlshape.Literal("p.forensic_keyword_created_day"), Kind: FieldKindDate},
		}),
		artifactSurface(artifactInvestigativeQueriesViewSchemaID, "investigative_query", []Field{
			{Key: "investigative_query.platform", Expr: sqlshape.Literal("p.investigative_query_platform"), Kind: FieldKindText},
			{Key: "investigative_query.purpose", Expr: sqlshape.Literal("p.investigative_query_purpose"), Kind: FieldKindText},
			{Key: "investigative_query.query_text", Expr: sqlshape.Literal("p.investigative_query_query_text"), Kind: FieldKindText},
			{Key: "investigative_query.created_by_user_id", Expr: sqlshape.Literal("p.investigative_query_created_by_user_id"), Kind: FieldKindText},
			{Key: "investigative_query.created_at", Expr: sqlshape.Literal("p.investigative_query_created_at"), Kind: FieldKindTimestamp},
			{Key: "investigative_query.query_id", Expr: sqlshape.Literal("p.investigative_query_query_id"), Kind: FieldKindText},
			{Key: "investigative_query.created_day", Expr: sqlshape.Literal("p.investigative_query_created_day"), Kind: FieldKindDate},
		}),
		artifactSurface(artifactNotesViewSchemaID, "note", []Field{
			{Key: "note.title", Expr: sqlshape.Literal("p.title"), Kind: FieldKindText},
			{Key: "note.body", Expr: sqlshape.Literal("p.body"), Kind: FieldKindText},
			{Key: "note.tags", Expr: tagCollectionExprFor(), Kind: FieldKindCollection},
			{Key: "note.linked_record_count", Expr: sqlshape.Literal("p.linked_record_count"), Kind: FieldKindNumber},
			{Key: "note.updated_at", Expr: sqlshape.Literal("p.updated_at"), Kind: FieldKindTimestamp},
			{Key: "note.created_by_user_id", Expr: sqlshape.Literal("p.created_by_user_id"), Kind: FieldKindText},
		}),
		artifactSurface(artifactCommLogViewSchemaID, "comm_log", []Field{
			{Key: "comm_log.timestamp_utc", Expr: sqlshape.Literal("p.timestamp_utc"), Kind: FieldKindTimestamp},
			{Key: "comm_log.comm_type", Expr: sqlshape.Literal("p.comm_type"), Kind: FieldKindText},
			{Key: "comm_log.audience", Expr: sqlshape.Literal("p.audience"), Kind: FieldKindText},
			{Key: "comm_log.channel_or_meeting", Expr: sqlshape.Literal("p.channel_or_meeting"), Kind: FieldKindText},
			{Key: "comm_log.summary", Expr: sqlshape.Literal("p.summary"), Kind: FieldKindText},
			{Key: "comm_log.next_report_at", Expr: sqlshape.Literal("p.next_report_at"), Kind: FieldKindTimestamp},
			{Key: "comm_log.privilege_tag", Expr: sqlshape.Literal("p.privilege_tag"), Kind: FieldKindText},
			{Key: "comm_log.decision_ids", Expr: recordRefCollectionExpr("comm_log.decision_ids"), Kind: FieldKindCollection},
			{Key: "comm_log.action_task_ids", Expr: recordRefCollectionExpr("comm_log.action_task_ids"), Kind: FieldKindCollection},
			{Key: "comm_log.audience_party_ids", Expr: partyRefCollectionExpr("comm_log.audience_party_ids"), Kind: FieldKindCollection},
			{Key: "comm_log.attendee_party_ids", Expr: partyRefCollectionExpr("comm_log.attendee_party_ids"), Kind: FieldKindCollection},
			{Key: "comm_log.comm_id", Expr: sqlshape.Literal("p.comm_id"), Kind: FieldKindText},
			{Key: "comm_log.timestamp_day", Expr: sqlshape.Literal("p.timestamp_day"), Kind: FieldKindDate},
			{Key: "comm_log.next_report_day", Expr: sqlshape.Literal("p.next_report_day"), Kind: FieldKindDate},
			{Key: "comm_log.updated_at", Expr: sqlshape.Literal("p.updated_at"), Kind: FieldKindTimestamp},
		}),
		artifactSurface(artifactHandoffViewSchemaID, "handoff", []Field{
			{Key: "handoff.timestamp_utc", Expr: sqlshape.Literal("p.timestamp_utc"), Kind: FieldKindTimestamp},
			{Key: "handoff.outgoing_owner_user_id", Expr: sqlshape.Literal("p.outgoing_owner_user_id"), Kind: FieldKindText},
			{Key: "handoff.incoming_owner_user_id", Expr: sqlshape.Literal("p.incoming_owner_user_id"), Kind: FieldKindText},
			{Key: "handoff.current_state_summary", Expr: sqlshape.Literal("p.current_state_summary"), Kind: FieldKindText},
			{Key: "handoff.open_task_ids", Expr: recordRefCollectionExpr("handoff.open_task_ids"), Kind: FieldKindCollection},
			{Key: "handoff.open_decision_ids", Expr: recordRefCollectionExpr("handoff.open_decision_ids"), Kind: FieldKindCollection},
			{Key: "handoff.open_risk_refs", Expr: riskRefCollectionExpr(), Kind: FieldKindCollection},
			{Key: "handoff.next_checks", Expr: sqlshape.Literal("p.next_checks"), Kind: FieldKindText},
			{Key: "handoff.acknowledged_at", Expr: sqlshape.Literal("p.acknowledged_at"), Kind: FieldKindTimestamp},
			{Key: "handoff.handoff_id", Expr: sqlshape.Literal("p.handoff_id"), Kind: FieldKindText},
			{Key: "handoff.timestamp_day", Expr: sqlshape.Literal("p.timestamp_day"), Kind: FieldKindDate},
			{Key: "handoff.ack_state", Expr: sqlshape.Literal("p.ack_state"), SortExpr: enumSortExpr(sqlshape.Literal("p.ack_state"), "pending", "acknowledged"), Kind: FieldKindText},
			{Key: "handoff.updated_at", Expr: sqlshape.Literal("p.updated_at"), Kind: FieldKindTimestamp},
		}),
		artifactSurface(artifactStatusReviewViewSchemaID, "status_review", []Field{
			{Key: "status_review.timestamp_utc", Expr: sqlshape.Literal("p.timestamp_utc"), Kind: FieldKindTimestamp},
			{Key: "status_review.review_owner_user_id", Expr: sqlshape.Literal("p.review_owner_user_id"), Kind: FieldKindText},
			{Key: "status_review.current_state_summary", Expr: sqlshape.Literal("p.current_state_summary"), Kind: FieldKindText},
			{Key: "status_review.blocked_task_ids", Expr: recordRefCollectionExpr("status_review.blocked_task_ids"), Kind: FieldKindCollection},
			{Key: "status_review.pending_evidence_ids", Expr: recordRefCollectionExpr("status_review.pending_evidence_ids"), Kind: FieldKindCollection},
			{Key: "status_review.open_decision_ids", Expr: recordRefCollectionExpr("status_review.open_decision_ids"), Kind: FieldKindCollection},
			{Key: "status_review.active_risks_summary", Expr: sqlshape.Literal("p.active_risks_summary"), Kind: FieldKindText},
			{Key: "status_review.next_report_at", Expr: sqlshape.Literal("p.next_report_at"), Kind: FieldKindTimestamp},
			{Key: "status_review.status_review_id", Expr: sqlshape.Literal("p.status_review_id"), Kind: FieldKindText},
			{Key: "status_review.timestamp_day", Expr: sqlshape.Literal("p.timestamp_day"), Kind: FieldKindDate},
			{Key: "status_review.next_report_day", Expr: sqlshape.Literal("p.next_report_day"), Kind: FieldKindDate},
			{Key: "status_review.updated_at", Expr: sqlshape.Literal("p.updated_at"), Kind: FieldKindTimestamp},
		}),
		artifactSurface(artifactLessonViewSchemaID, "lesson", []Field{
			{Key: "lesson.timestamp_utc", Expr: sqlshape.Literal("p.timestamp_utc"), Kind: FieldKindTimestamp},
			{Key: "lesson.summary", Expr: sqlshape.Literal("p.summary"), Kind: FieldKindText},
			{Key: "lesson.owner_user_id", Expr: sqlshape.Literal("p.owner_user_id"), Kind: FieldKindText},
			{Key: "lesson.closure_state", Expr: sqlshape.Literal("p.closure_state"), Kind: FieldKindText},
			{Key: "lesson.follow_up_task_ids", Expr: recordRefCollectionExpr("lesson.follow_up_task_ids"), Kind: FieldKindCollection},
			{Key: "lesson.evidence_refs", Expr: recordRefCollectionExpr("lesson.evidence_refs"), Kind: FieldKindCollection},
			{Key: "lesson.lesson_id", Expr: sqlshape.Literal("p.lesson_id"), Kind: FieldKindText},
			{Key: "lesson.timestamp_day", Expr: sqlshape.Literal("p.timestamp_day"), Kind: FieldKindDate},
			{Key: "lesson.updated_at", Expr: sqlshape.Literal("p.updated_at"), Kind: FieldKindTimestamp},
		}),
	}
}

func artifactSurface(viewSchemaID string, artifactType string, fields []Field) Surface {
	return Surface{
		ViewSchemaID: viewSchemaID,
		FromSQL:      sqlshape.Literal("FROM artifact_grid_projection p JOIN records r ON r.record_id = p.record_id"),
		RecordExpr:   sqlshape.Literal("p.record_id"),
		IncidentExpr: sqlshape.Literal("p.incident_id"),
		WhereSQL:     sqlshape.Format("p.artifact_type = %s", sqlshape.Quote(artifactType)),
		Fields:       fields,
	}
}

func enumSortExpr(expr sqlshape.Fragment, values ...string) sqlshape.Fragment {
	var builder sqlshape.Builder
	builder.Write(sqlshape.Literal("CASE"))
	builder.Write(expr)
	for index, value := range values {
		builder.Write(sqlshape.Literal("WHEN"))
		builder.Write(sqlshape.Quote(value))
		builder.Write(sqlshape.Literal("THEN"))
		builder.Write(sqlshape.Number(index))
	}
	builder.Write(sqlshape.Literal("ELSE"))
	builder.Write(sqlshape.Number(len(values)))
	builder.Write(sqlshape.Literal("END"))
	return builder.Fragment()
}

func recordRefCollectionExpr(fieldKey string) sqlshape.Fragment {
	return recordRefCollectionExprFor(fieldKey, "references_record")
}

func recordRefCollectionExprFor(fieldKey string, linkType string) sqlshape.Fragment {
	return sqlshape.Format(`(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', %s,
        'item_kind', 'record_ref',
        'display_text', dst.record_type || ':' || dst.record_id::text,
        'linked_record_id', dst.record_id::text
    ) ORDER BY dst.record_type ASC, dst.record_id ASC), '[]'::jsonb)
      FROM active_record_links_v1 rl
      JOIN records dst
        ON dst.incident_id = rl.incident_id
       AND dst.record_id = rl.dst_record_id
       AND dst.deleted_at IS NULL
     WHERE rl.incident_id = p.incident_id
       AND rl.src_record_id = p.record_id
       AND rl.link_type = %s
       AND rl.field_key = %s)::text`, recordRefItemRefSQL(sqlshape.Literal("dst.record_id")), sqlshape.Quote(linkType), sqlshape.Quote(fieldKey))
}

func tagCollectionExprFor() sqlshape.Fragment {
	return sqlshape.Format(`(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', %s,
        'item_kind', 'tag',
        'display_text', rt.tag_name,
        'tag_id', rt.record_tag_id::text
    ) ORDER BY rt.normalized_tag_name ASC, rt.record_tag_id ASC), '[]'::jsonb)
      FROM active_record_tags_v1 rt
     WHERE rt.incident_id = p.incident_id
       AND rt.record_id = p.record_id)::text`, recordTagItemRefSQL(sqlshape.Literal("rt.record_id"), sqlshape.Literal("rt.record_tag_id")))
}

func partyRefCollectionExpr(fieldKey string) sqlshape.Fragment {
	return sqlshape.Format(`(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', %s,
        'item_kind', 'party_ref',
        'display_text', party.display_name,
        'party_id', party.record_id::text
    ) ORDER BY party.display_name ASC, party.record_id ASC), '[]'::jsonb)
      FROM active_record_links_v1 rl
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
       AND rl.field_key = %s)::text`, partyRefItemRefSQL(sqlshape.Literal("party.record_id")), sqlshape.Quote(fieldKey))
}

func riskRefCollectionExpr() sqlshape.Fragment {
	return sqlshape.Literal(`(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', 'risk_ref:' || risk_ref_id::text,
        'item_kind', 'risk_ref',
        'display_text', risk_ref_text,
        'risk_ref_id', risk_ref_id::text,
        'risk_ref_text', risk_ref_text
    ) ORDER BY risk_ref_text ASC, risk_ref_id ASC), '[]'::jsonb)
      FROM handoff_risk_refs hr
     WHERE hr.incident_id = p.incident_id
       AND hr.handoff_record_id = p.record_id
       AND hr.deleted_at IS NULL)::text`)
}
