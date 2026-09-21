package revisions_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/historytest"
	"github.com/google/uuid"
)

func TestHistorySourceFamilyPaging_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "history-source-pages")
	login, actor := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident, _ := seedRecord(t, harness.DB, harness.Server, login, actor, "IR-HISTORY-SOURCES")
	contributions, err := revisionassembly.CurrentProviderContributions()
	if err != nil {
		t.Fatal(err)
	}
	check := func(t *testing.T, recordType string, snapshot map[string]any, kind string, facts map[string]any) {
		t.Helper()
		record := uuid.New()
		envelopetest.SeedRecordEnvelope(t, harness.DB, incident, actor, record, recordType)
		mustExec(t, harness.DB, `UPDATE records SET deleted_at=now(),deleted_by_user_id=$2 WHERE record_id=$1`, record, actor)
		encode := func(value map[string]any) []byte {
			raw, e := json.Marshal(value)
			if e != nil {
				t.Fatal(e)
			}
			return []byte(strings.ReplaceAll(string(raw), historytest.RecordID, record.String()))
		}
		at := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
		for i := 1; i <= 2; i++ {
			cs := uuid.New()
			mustExec(t, harness.DB, `INSERT INTO change_sets(change_set_id,incident_id,actor_user_id,source,created_at) VALUES($1,$2,$3,'history-source-test',$4)`, cs, incident, actor, at.Add(time.Duration(i)*time.Second))
			if snapshot != nil {
				mustExec(t, harness.DB, `INSERT INTO record_revisions(change_set_id,record_id,row_version,before_json,after_json,created_at) VALUES($1,$2,$3,NULL,$4,$5)`, cs, record, i, encode(snapshot), at)
			}
			// The first row-family event is revision-only; the second uses the source's
			// row mutation projector. Non-row families have two complete collection events.
			if snapshot == nil || i == 2 {
				target := record.String()
				if kind == "record_tag" {
					target = fmt.Sprintf("record_tag:%s:%s", record, record)
				}
				mustExec(t, harness.DB, `INSERT INTO change_set_mutations(change_set_id,sequence_no,target_kind,target_id,operation_kind,before_value,after_value,history_record_ids,history_entry_record_ids) VALUES($1,1,$2,$3,'create',NULL,$4,ARRAY[$5::uuid],'{}')`, cs, kind, target, encode(facts), record)
			}
		}
		full := historyItems(getHistory(t, harness.Server.HTTP.URL, login, record, "?limit=100"))
		if len(full) != 2 {
			t.Fatalf("events=%d, want 2", len(full))
		}
		pages := collectHistoryPages(t, harness.Server.HTTP.URL, login, record, 1)
		if !reflect.DeepEqual(pages, full) {
			t.Fatal("paging changed complete semantic events")
		}
		for _, item := range full {
			diff := item.(map[string]any)["diff_summary"].(map[string]any)
			if len(diff["units"].([]any)) == 0 {
				t.Fatal("empty semantic event")
			}
		}
		raw, _ := json.Marshal(full)
		for _, private := range []string{"private-upload-token", "private-credential"} {
			if strings.Contains(string(raw), private) {
				t.Fatal("private retained fact exposed")
			}
		}
	}
	count := 0
	for _, owner := range contributions {
		for _, record := range owner.Records {
			for _, variant := range historytest.Variants(record.RecordType) {
				t.Run(record.RecordType+variant, func(t *testing.T) {
					snapshot := historytest.Snapshot(record.RecordType, record.SnapshotSchemaID, variant)
					check(t, record.RecordType, snapshot, "record", snapshot)
				})
				count++
			}
		}
		for _, target := range owner.NonRowTargets {
			t.Run(target.TargetKind, func(t *testing.T) {
				check(t, "host", nil, target.TargetKind, historytest.NonRowFacts(target.TargetKind))
			})
			count++
		}
	}
	if count != 24 {
		t.Fatalf("source configurations=%d, want 24", count)
	}
}
