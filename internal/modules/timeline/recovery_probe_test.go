package timeline_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

func TestRestoreWorkbookProbeRegistrationMatchesCanonicalFixture(t *testing.T) {
	var fixtureJSON string
	for _, artifact := range contractrecovery.Artifacts {
		if artifact.Path == "contracts/recovery/fixtures/restore-workbook-probe-registration.v1.json" {
			fixtureJSON = artifact.JSON
			break
		}
	}
	if fixtureJSON == "" {
		t.Fatal("generated canonical restore workbook probe fixture is missing")
	}
	body, err := json.Marshal(timeline.RestoreWorkbookProbeRegistration())
	if err != nil {
		t.Fatalf("encode Timeline restore workbook probe registration: %v", err)
	}
	var got any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode registration: %v", err)
	}
	var want any
	if err := json.Unmarshal([]byte(fixtureJSON), &want); err != nil {
		t.Fatalf("decode canonical fixture: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Timeline restore workbook probe registration got %s want %s", body, fixtureJSON)
	}
}
