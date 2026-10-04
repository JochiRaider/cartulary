package reporting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
)

func TestPortableArtifactCatalogClosure(t *testing.T) {
	incident := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	snapshot := "40000000-0000-0000-0000-000000000002"
	release := "40000000-0000-0000-0000-000000000003"
	c := emptyPortableCatalog(incident)
	c.Snapshots = append(c.Snapshots, portableSnapshot{snapshot, portableModelPath(snapshot), strings.Repeat("a", 64)})
	c.Releases = append(c.Releases, portableRelease{release, snapshot, portableManifestPath(release), strings.Repeat("b", 64)})
	c.OmittedFiles = append(c.OmittedFiles, portableOmission{release, "internal/reveal-map.json", "sensitive_reveal_map"})
	data, err := canonicaljson.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodePortableCatalog(data, incident); err != nil {
		t.Fatal(err)
	}
	type objectSelector func(map[string]any) map[string]any
	for label, selectObject := range map[string]objectSelector{
		"catalog":  func(m map[string]any) map[string]any { return m },
		"snapshot": func(m map[string]any) map[string]any { return m["snapshots"].([]any)[0].(map[string]any) },
		"release":  func(m map[string]any) map[string]any { return m["releases"].([]any)[0].(map[string]any) },
		"omission": func(m map[string]any) map[string]any { return m["omitted_files"].([]any)[0].(map[string]any) },
	} {
		var pristine map[string]any
		if err := json.Unmarshal(data, &pristine); err != nil {
			t.Fatal(err)
		}
		for field := range selectObject(pristine) {
			for _, mutation := range []string{"omitted", "null", "wrong type"} {
				t.Run(label+"/"+field+"/"+mutation, func(t *testing.T) {
					var m map[string]any
					_ = json.Unmarshal(data, &m)
					o := selectObject(m)
					switch mutation {
					case "omitted":
						delete(o, field)
					case "null":
						o[field] = nil
					default:
						o[field] = false
					}
					raw, _ := canonicaljson.Marshal(m)
					if _, err := decodePortableCatalog(raw, incident); err == nil {
						t.Fatal("accepted invalid member")
					}
				})
			}
		}
		t.Run(label+"/unknown", func(t *testing.T) {
			var m map[string]any
			_ = json.Unmarshal(data, &m)
			selectObject(m)["hostile-secret"] = true
			raw, _ := canonicaljson.Marshal(m)
			if _, err := decodePortableCatalog(raw, incident); err == nil || strings.Contains(err.Error(), "hostile-secret") {
				t.Fatal("unknown member was admitted or echoed")
			}
		})
	}
	for name, raw := range map[string][]byte{
		"duplicate key":     bytes.Replace(data, []byte(`"schema_id":`), []byte(`"schema_id":"other","schema_id":`), 1),
		"malformed unicode": bytes.Replace(data, []byte(`"sensitive_reveal_map"`), []byte(`"\ud800"`), 1),
		"whitespace":        append([]byte(" "), data...),
		"wrong incident":    bytes.Replace(data, []byte(incident.String()), []byte(snapshot), 1),
		"noncanonical ID":   bytes.ReplaceAll(data, []byte(snapshot), []byte(strings.ToUpper("40000000-0000-0000-a000-000000000002"))),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := decodePortableCatalog(raw, incident); err == nil {
				t.Fatal("accepted invalid catalog")
			}
		})
	}
	t.Run("reachable cardinality boundaries", func(t *testing.T) {
		full := emptyPortableCatalog(incident)
		for n := 1; n <= portableSnapshotLimit; n++ {
			id := fmt.Sprintf("40000000-0000-0001-0000-%012d", n)
			full.Snapshots = append(full.Snapshots, portableSnapshot{id, portableModelPath(id), strings.Repeat("a", 64)})
		}
		for n := 1; n <= portableReleaseLimit; n++ {
			id := fmt.Sprintf("40000000-0000-0002-0000-%012d", n)
			full.Releases = append(full.Releases, portableRelease{id, full.Snapshots[0].SnapshotID, portableManifestPath(id), strings.Repeat("b", 64)})
			full.OmittedFiles = append(full.OmittedFiles, portableOmission{id, "internal/reveal-map.json", "sensitive_reveal_map"})
		}
		raw, _ := canonicaljson.Marshal(full)
		if _, err := decodePortableCatalog(raw, incident); err != nil {
			t.Fatal("reachable catalog maximum rejected", err)
		}
		for _, which := range []string{"snapshot", "release", "omission"} {
			overflow := full
			switch which {
			case "snapshot":
				overflow.Snapshots = append(append([]portableSnapshot{}, full.Snapshots...), full.Snapshots[0])
			case "release":
				overflow.Releases = append(append([]portableRelease{}, full.Releases...), full.Releases[0])
			default:
				overflow.OmittedFiles = append(append([]portableOmission{}, full.OmittedFiles...), full.OmittedFiles[0])
			}
			raw, _ := canonicaljson.Marshal(overflow)
			if _, err := decodePortableCatalog(raw, incident); err == nil {
				t.Fatal("accepted cardinality overflow", which)
			}
		}
	})
	for _, name := range []string{"../escape", "x/../escape", "x//file", "/absolute", "x\\file", "x:stream", "x\x00file", strings.Repeat("x", 513)} {
		if portableMemberPath(name) {
			t.Fatal("unsafe member admitted")
		}
	}
	if !portableMemberPath(strings.Repeat("x", 512)) {
		t.Fatal("path equality rejected")
	}
}
