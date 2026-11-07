package referenceadmissionfixture

import (
	"database/sql"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func seedMalformedCurrent(t testing.TB, db *sql.DB, kind string) {
	t.Helper()
	read := func(path string) map[string]any {
		t.Helper()
		b, err := os.ReadFile("../../../" + path)
		if err != nil {
			t.Fatal(err)
		}
		var value map[string]any
		if err := json.Unmarshal(b, &value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	artifact := "set"
	vector := read("contracts/reference-pack-fixtures/fixtures/complete-set.v1.json")
	id := vector["expected_set_id"].(string)
	value := map[string]any{"schema_id": "cartulary.reference_pack_set.v2", "pack_set_id": id, "pack_set_sha256": strings.TrimPrefix(id, "rpset_"), "members": vector["members"]}
	if strings.HasPrefix(kind, "current manifest") {
		artifact = "manifest"
		release := read("fixtures/entra-app-ids/release.v2.json")
		value = read(release["logical_root"].(string) + "/manifest.json")
		if kind == "current manifest extra" {
			value["extra"] = true
		} else {
			delete(value, "source_identifier")
		}
	} else if kind == "current set unordered" {
		members := value["members"].([]any)
		members[0], members[1] = members[1], members[0]
	}
	raw, err := canonicaljson.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if kind == "current set noncanonical" {
		raw = append(raw, '\n')
	}
	seedWith(t, db, artifact, nil, map[string][]byte{artifact: raw})
}
