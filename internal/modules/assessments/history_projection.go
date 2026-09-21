package assessments

import "github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"

func projectRecordHistory(facts historycontract.Facts) ([]historycontract.Unit, error) {
	fields := historycontract.Fields("assessment", "text", "subject_type assessment_state rationale assessed_at")
	fields = append(fields, historycontract.Field{Key: "assessment.subject_ref", Member: "subject_record_id", Type: "uuid", Required: true}, historycontract.Field{Key: "assessment.assessor", Member: "assessor_user_id", Type: "text", Required: true}, historycontract.Field{Key: "assessment.confidence_score", Member: "confidence_score", Type: "number"})
	return historycontract.Row(facts, fields)
}
