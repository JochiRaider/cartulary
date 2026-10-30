package imports

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestCSVSourceAdmissionIsTargetNeutral(t *testing.T) {
	s := &service{limits: Limits{MaxCSVSourceBytes: 1024, MaxRows: 1, MaxColumns: 1, MaxCells: 1}}
	for _, content := range []string{"", "a,b\n1,2\n3,4\n", "header\nliteral\"quote\n", "header\n\"unclosed", "header\n\xff"} {
		units, err := s.discoverImportUnits(httpapi.UploadEnvelope{File: []byte(content)}, sourceFileKindCSV)
		if err != nil || len(units) != 1 {
			t.Fatalf("source admission interpreted %q: %v", content, err)
		}
		u := units[0]
		if u.LocatorKind != "csv_file" || u.SourceRectA1 != "" || u.Columns != nil || u.SourceRows != nil || u.PreviewRows != nil {
			t.Fatal("source admission fabricated a projection")
		}
	}
	if _, err := s.discoverImportUnits(httpapi.UploadEnvelope{File: []byte(strings.Repeat("x", 1025))}, sourceFileKindCSV); err == nil {
		t.Fatal("CSV byte overflow admitted")
	}
	for _, body := range []string{`{}`, `{"target_kind":"network_flow_table","extension_profile_id":null}`, `{"target_kind":"network_flow_table","extension_profile_id":"network_flow_activity","client_txn_id":"no"}`} {
		if _, err := decodeSourceInspectionRequest(strings.NewReader(body)); err == nil {
			t.Fatalf("invalid inspection request admitted: %s", body)
		}
	}
}
