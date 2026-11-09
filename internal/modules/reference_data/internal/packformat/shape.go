package packformat

import (
	"bytes"
	"errors"
	"regexp"
	"strings"

	"github.com/JochiRaider/cartulary/internal/gen/contractreferencepacks"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	formatschema "github.com/JochiRaider/cartulary/referencepacks/format/schema"
)

func projection(name string) []byte {
	for _, artifact := range contractreferencepacks.Artifacts {
		if artifact.Path == "contracts/reference-packs/"+name {
			return []byte(artifact.JSON)
		}
	}
	panic("reference pack build is missing a required typed projection: " + name)
}

// Application projections resolve only within compiled owner artifacts. The
// neutral format library owns schema evaluation and bounded diagnostics.
type shape struct{ compiled *formatschema.Shape }

func projectionResources() map[string][]byte {
	resources := map[string][]byte{}
	for _, a := range contractreferencepacks.Artifacts {
		resources[strings.TrimPrefix(a.Path, "contracts/reference-packs/")] = []byte(a.JSON)
	}
	return resources
}
func compileProjection(name string) *shape {
	return &shape{formatschema.Compile(projectionResources(), fixedShapeLimits, name)}
}

var projectionReferencePattern = regexp.MustCompile(`^(?:profiles/)?[a-z0-9_.-]+\.schema\.json$`)
var errShapeMismatch = errors.New("reference pack schema mismatch")
var errShapeLimit = errors.New("reference pack schema resource limit reached")

func (s *shape) matches(value any) bool { return s.compiled.Matches(value) }
func (s *shape) walk(value any, path string, emit FindingSink) error {
	return formatError(s.compiled.Walk(value, path, emit))
}
func (s *shape) property(name string) *shape {
	child := s.compiled.Property(name)
	if child == nil {
		return nil
	}
	return &shape{child}
}
func equalJSON(a, b any) bool {
	left, e1 := canonicaljson.Marshal(a)
	right, e2 := canonicaljson.Marshal(b)
	return e1 == nil && e2 == nil && bytes.Equal(left, right)
}
