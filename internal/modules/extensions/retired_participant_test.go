package extensions

import "testing"

func checkRetiredBackupParticipant(t *testing.T) {
	for _, retired := range []string{
		"cartulary.extension_participant_contract.v1",
		"cartulary.extension_participant_specialization.v3",
		"cartulary.extension_snapshot_reporting_participant_context.v1",
		"cartulary.extension_snapshot_reporting_participant_result.v1",
	} {
		t.Run(retired, func(t *testing.T) {
			source := newMutableGeneratedSource()
			changed := false
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				if object["schema_id"] == "cartulary.extension_reporting_source_contract.v1" && object["participant_id"] != nil {
					object["schema_id"] = retired
					changed = true
				}
			})
			if !changed {
				t.Fatal("current source fixture missing")
			}
			_, err := NewCoordinator(source)
			requireFinding(t, err, "extension_registry_invalid", "registry_generation")
		})
	}
}
