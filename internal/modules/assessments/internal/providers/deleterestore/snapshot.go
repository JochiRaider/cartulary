package deleterestore

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/assessments/internal/policy"
	shape "github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

func ValidateSnapshotSource(record, source map[string]any) error {
	if err := shape.ValidateSourceObject(record, source,
		shape.SourceFieldGroup{Names: "record_id incident_id subject_record_id assessor_user_id", Kind: shape.SourceUUID, Nullable: false},
		shape.SourceFieldGroup{Names: "subject_type assessment_state rationale", Kind: shape.SourceString, Nullable: false},
		shape.SourceFieldGroup{Names: "confidence_score", Kind: shape.SourceInteger, Nullable: true},
		shape.SourceFieldGroup{Names: "assessed_at created_at updated_at", Kind: shape.SourceTimestamp, Nullable: false},
		shape.SourceFieldGroup{Names: "deleted_at", Kind: shape.SourceTimestamp, Nullable: true},
		shape.SourceFieldGroup{Names: "deleted_by_user_id", Kind: shape.SourceUUID, Nullable: true},
	); err != nil {
		return err
	}
	if !policy.ValidSubjectType(source["subject_type"].(string)) || source["subject_record_id"] == source["record_id"] {
		return fmt.Errorf("invalid assessment snapshot subject")
	}
	if !policy.ValidState(source["assessment_state"].(string)) {
		return fmt.Errorf("invalid assessment snapshot state")
	}
	if raw := source["confidence_score"]; raw != nil {
		value, err := raw.(json.Number).Int64()
		score := int(value)
		if _, valid := policy.ConfidenceBand(&score); err != nil || int64(score) != value || !valid {
			return fmt.Errorf("invalid assessment snapshot confidence")
		}
	}
	rationale := source["rationale"].(string)
	if normalized, valid := fieldnorm.NormalizeNote(rationale); !valid || normalized != rationale {
		return fmt.Errorf("noncanonical assessment snapshot rationale")
	}
	// Scalar admission has checked timestamp syntax. assessed_at is an explicit
	// judgment instant and is intentionally independent of the persistence clocks.
	created, _ := time.Parse(time.RFC3339Nano, source["created_at"].(string))
	updated, _ := time.Parse(time.RFC3339Nano, source["updated_at"].(string))
	if updated.Before(created) {
		return fmt.Errorf("assessment snapshot clocks are reversed")
	}
	deletedAt, deletedBy := source["deleted_at"], source["deleted_by_user_id"]
	if (deletedAt == nil) != (deletedBy == nil) {
		return fmt.Errorf("incomplete assessment snapshot deletion tuple")
	}
	if deletedAt != nil {
		deleted, _ := time.Parse(time.RFC3339Nano, deletedAt.(string))
		if deleted.Before(created) || deleted.After(updated) {
			return fmt.Errorf("assessment snapshot deletion outside source lifetime")
		}
	}
	return nil
}
