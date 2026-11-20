package serverprocess

import (
	"net/http"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

// deletedRecordAdditions are record mutations whose path addresses a child
// object but whose REQ-01-705 order includes a record, with the reason.
var deletedRecordAdditions = map[string]string{
	"resolveEntityMention": "a mention action classifies its source record (REQ-01-705 step 2, REQ-01-211)",
}

// deletedRecordExclusions are record-addressed mutations outside the family,
// each with the requirement that places it there.
var deletedRecordExclusions = map[string]string{
	"deleteRecord":                "soft-delete of a soft-deleted record is record_already_deleted (REQ-01-076); asserted by its own case below",
	"restoreRecord":               "REQ-01-104 destructive-operation family: it operates on deleted records and evaluates eligibility after acquiring its protected set",
	"rollbackRecord":              "REQ-01-104 destructive-operation family: it operates on deleted records and evaluates eligibility after acquiring its protected set",
	"mergeEntityRecord":           "REQ-01-104 destructive-operation family: a deleted survivor or loser is merge_precondition_failed (REQ-01-705)",
	"issueEvidencePreviewHandle":  "handle issuance reads its record and is not a mutation; allowed on closed incidents (REQ-01-705, REQ-01-590)",
	"issueEvidenceDownloadHandle": "handle issuance reads its record and is not a mutation; allowed on closed incidents (REQ-01-705, REQ-01-590)",
}

// deletedRecordCases holds one case per catalog-derived member. Each case soft
// deletes its target through the public route and returns fresh requests that
// would succeed on the live record. Where an operation spans several record
// owners, one representative type suffices here: the per-owner order is the
// record target classification row's subject.
var deletedRecordCases = map[string]func(*conformanceEnv) []conformanceRequest{
	"patchRecord": func(e *conformanceEnv) []conformanceRequest {
		noteID := e.note("Deleted patch target")
		e.softDelete(noteID)
		return []conformanceRequest{e.notePatch(noteID)}
	},
	"resolveRecordSameFieldConflict": func(e *conformanceEnv) []conformanceRequest {
		recordID, token := e.timelineConflict("Deleted conflict")
		e.softDelete(recordID)
		return e.conflictResolutions(recordID, token)
	},
	"markTimelineRecordReviewed": func(e *conformanceEnv) []conformanceRequest {
		recordID := e.timeline("Deleted review target")
		e.softDelete(recordID)
		return []conformanceRequest{e.timelineAction(recordID, "mark-reviewed")}
	},
	"supersedeRecord": func(e *conformanceEnv) []conformanceRequest {
		recordID := e.timeline("Deleted supersede target")
		e.softDelete(recordID)
		return []conformanceRequest{e.timelineAction(recordID, "supersede")}
	},
	"mutateNoteAssociations": func(e *conformanceEnv) []conformanceRequest {
		noteID, counterpartID := e.note("Deleted association"), e.note("Live counterpart")
		e.softDelete(noteID)
		return []conformanceRequest{e.noteAssociation(noteID, counterpartID)}
	},
	"resolveEntityMention": func(e *conformanceEnv) []conformanceRequest {
		sourceID, mentionID, hostID := e.unresolvedMention("deleted")
		e.softDelete(sourceID)
		return []conformanceRequest{e.mentionResolution(mentionID, hostID)}
	},
	"appendIndicatorStateInterval": func(e *conformanceEnv) []conformanceRequest {
		indicatorID := e.indicator("deleted-interval.conformance.test")
		supportID := e.timeline("Deleted interval support")
		e.softDelete(indicatorID)
		return []conformanceRequest{e.stateInterval(indicatorID, supportID)}
	},
	"createManualIndicatorObservation": func(e *conformanceEnv) []conformanceRequest {
		text := "deleted-manual.conformance.test"
		sourceID := e.timeline(text)
		e.softDelete(sourceID)
		return []conformanceRequest{e.manualObservation(sourceID, "timeline.raw_activity_text", text)}
	},
	"createRecordLinkedNote": func(e *conformanceEnv) []conformanceRequest {
		hostID := e.host("deleted-linked-source")
		e.softDelete(hostID)
		return []conformanceRequest{e.linkedNote(hostID)}
	},
	"attachBlobToEvidenceRecord": func(e *conformanceEnv) []conformanceRequest {
		evidenceID := e.evidence("Deleted attach target")
		blobID := e.verifiedBlob("deleted-attach")
		e.softDelete(evidenceID)
		return []conformanceRequest{e.attach(evidenceID, blobID)}
	},
}

// deletedRecordOperations derives Core 04 AC-616's deleted-record family from
// the catalog: every state-changing operation whose path names a record, with
// the stated additions, less the stated exclusions. It fails on a stale
// addition or exclusion, so the stated reasons always name catalog members.
func deletedRecordOperations(t *testing.T, operations map[string]httpapi.ContractOperation) []string {
	t.Helper()
	recordTypes := publicRecordTypes()
	if !slices.Contains(recordTypes, "timeline_event") || !slices.Contains(recordTypes, "indicator") {
		t.Fatalf("public record types %v lack the view-sourced record types", recordTypes)
	}
	var addressed []string
	for id, operation := range operations {
		if operation.StateChanging && len(recordPathParameters(operation, recordTypes)) != 0 {
			addressed = append(addressed, id)
		}
	}
	for id := range deletedRecordAdditions {
		if _, ok := operations[id]; !ok || slices.Contains(addressed, id) {
			t.Fatalf("deleted-record addition %s is not a catalog operation outside the record-addressed set", id)
		}
		addressed = append(addressed, id)
	}
	var members []string
	for _, id := range addressed {
		if _, excluded := deletedRecordExclusions[id]; !excluded {
			members = append(members, id)
		}
	}
	for id := range deletedRecordExclusions {
		if !slices.Contains(addressed, id) {
			t.Fatalf("deleted-record exclusion %s does not name a record-addressed mutation", id)
		}
	}
	slices.Sort(members)
	return members
}

// TestDeletedRecordConformance verifies Core 04 AC-616's deleted-record family
// (REQ-01-705, REQ-01-076): every mutation that addresses one record by path,
// other than soft-delete and the REQ-01-104 destructive family, declares 409
// record_deleted_use_restore and returns it for a fresh request on a
// soft-deleted target; soft-delete of a soft-deleted record returns 409
// record_already_deleted. Neither commits any effect. The case table is keyed
// by the catalog-derived set, so a new record-addressed mutation without a case
// fails by name.
func TestDeletedRecordConformance(t *testing.T) {
	catalog := catalogOperationsByID(t)
	operations := deletedRecordOperations(t, catalog)
	requireCaseKeySet(t, "record_deleted_use_restore", sortedKeys(deletedRecordCases), operations)
	for _, id := range operations {
		if !declaresErrorCode(catalog[id], http.StatusConflict, "record_deleted_use_restore") {
			t.Errorf("%s addresses a record but does not declare 409 record_deleted_use_restore", id)
		}
	}
	if !declaresErrorCode(catalog["deleteRecord"], http.StatusConflict, "record_already_deleted") {
		t.Error("deleteRecord does not declare 409 record_already_deleted")
	}
	if t.Failed() {
		t.FailNow()
	}

	e := startConformanceEnv(t, "deleted-record-conformance", "IR-DELETED-CONFORMANCE")
	prepared := make(map[string][]conformanceRequest, len(operations))
	for _, id := range operations {
		prepared[id] = deletedRecordCases[id](e)
		if len(prepared[id]) == 0 {
			t.Fatalf("%s case prepared no request", id)
		}
	}
	deletedNote := e.note("Deleted delete target")
	beforeDelete := workbookscenariotest.SnapshotIncidentEffects(t, e.harness, e.incidentID)
	e.softDelete(deletedNote)
	// The snapshot observes a soft-delete, so the unchanged snapshots below
	// are evidence rather than a blind comparison.
	changed := workbookscenariotest.ChangedIncidentEffects(beforeDelete, workbookscenariotest.SnapshotIncidentEffects(t, e.harness, e.incidentID))
	if !slices.Contains(changed, "source/history/receipts") || !slices.Contains(changed, "projections") {
		t.Fatalf("soft-delete changed effect classes %v, want the envelope history and projections", changed)
	}
	repeatDelete := conformanceRequest{
		label: "soft-deleted Note", params: map[string]string{"record_id": deletedNote.String()},
		body: map[string]any{"base_row_version": e.version(deletedNote), "client_txn_id": e.txn("delete-again"), "reason": "Conformance repeat"},
	}
	e.quiesce()

	for _, id := range operations {
		operation := e.operations[id]
		t.Run(id, func(t *testing.T) {
			for _, request := range prepared[id] {
				t.Run(request.label, func(t *testing.T) {
					e.requireRejected(t, operation, request, http.StatusConflict, "record_deleted_use_restore")
				})
			}
		})
	}
	t.Run("deleteRecord", func(t *testing.T) {
		e.requireRejected(t, e.operations["deleteRecord"], repeatDelete, http.StatusConflict, "record_already_deleted")
	})
}
