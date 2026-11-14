package deleterestore

import (
	"encoding/json"
	"fmt"

	shape "github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/internal/sourcevalue"
)

func ValidateSnapshotSource(record, source map[string]any) error {
	if err := shape.ValidateSourceObject(record, source,
		shape.SourceFieldGroup{Names: "record_id incident_id created_by_user_id updated_by_user_id", Kind: shape.SourceUUID, Nullable: false},
		shape.SourceFieldGroup{Names: "capture_state activity_time_pair_state", Kind: shape.SourceString, Nullable: false},
		shape.SourceFieldGroup{Names: "row_version", Kind: shape.SourceInteger, Nullable: false},
		shape.SourceFieldGroup{Names: "recorded_at edited_at", Kind: shape.SourceTimestamp, Nullable: false},
		shape.SourceFieldGroup{Names: "reviewed_by_user_id superseded_by_user_id", Kind: shape.SourceUUID, Nullable: true},
		shape.SourceFieldGroup{Names: "reviewed_at superseded_at", Kind: shape.SourceTimestamp, Nullable: true},
		shape.SourceFieldGroup{Names: "date_entered_text analyst_text mitre_stage_text device_object_text ip_address_text activity_utc_text activity_local_text raw_activity_text activity_synopsis_text data_source_text", Kind: shape.SourceString, Nullable: true},
		shape.SourceFieldGroup{Names: "activity_utc_generated activity_local_generated", Kind: shape.SourceBoolean, Nullable: false},
	); err != nil {
		return err
	}
	version, _ := source["row_version"].(json.Number).Int64()
	if version < 1 {
		return fmt.Errorf("invalid historical Timeline source version")
	}
	// The full snapshot shape above admits its envelope mirrors; intrinsic
	// source semantics are shared with logical portability admission.
	row := sourcevalue.ReadAdmittedFields(source)
	if !row.VisibleTextLegal() || !row.LifecycleLegal() || !row.GeneratedTimeLegal() || !row.PairStateLegal() {
		return fmt.Errorf("invalid historical Timeline source semantics")
	}
	return nil
}
