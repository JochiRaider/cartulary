package native

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"github.com/JochiRaider/cartulary/referencepacks/format/schema"
	"io"
	"os"
	"strings"
	"testing"
)

type Failure = envelope.Failure
type contentFixture struct {
	Key     string            `json:"pack_key"`
	Members map[string]string `json:"members"`
}

func contentFixtureFor(t *testing.T, key string) contentFixture {
	t.Helper()
	b, err := os.ReadFile("testdata/profile.json")
	if err != nil {
		t.Fatal(err)
	}
	var f contentFixture
	if err = json.Unmarshal(b, &f); err != nil || f.Key != key {
		t.Fatal("source model fixture identity", err)
	}
	return f
}

type contentMemory map[string]string

func (m contentMemory) Open(_ context.Context, name string) (io.ReadCloser, error) {
	b, ok := m[name]
	if !ok {
		return nil, errors.New("missing source-model fixture")
	}
	return io.NopCloser(strings.NewReader(b)), nil
}
func fixtureManifest(f contentFixture) Manifest {
	m := Manifest{Key: f.Key, ProfileID: definition.ID, Artifacts: []envelope.SourceArtifact{{Ref: "fixture"}}}
	if f.Key == "framework.veris" {
		m.SourceVersion = "fixture-v2"
		m.Artifacts[0].Ref = "https://example.invalid/verisc-merged.json"
	}
	count := func(name string) *int64 { n := int64(strings.Count(f.Members[name], "\n")); return &n }
	if definition.Shape == "objects_relationships" {
		m.Summary.Objects = count("payload/objects.ndjson")
		m.Summary.Relationships = count("payload/relationships.ndjson")
	} else {
		m.Summary.Entries = count("payload/entries.ndjson")
	}
	return m
}
func compileTestReference(name string) *schema.Shape {
	var limits map[string]int
	b, err := inputs.ReadFile("limits.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(b, &limits); err != nil {
		panic(err)
	}
	return schema.Compile(resources, limits, name)
}
