// Package packstate projects Reference Data conditions and resolves effective sets.
// It has no storage, transaction, transport, or Jobs dependencies.
package packstate

import (
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

type Health string

const (
	Available Health = "verified_available"
	Failed    Health = "failed"
	Missing   Health = "missing"
)

type Version struct {
	Member        packformat.SetMember
	Manifest      packformat.Manifest
	Health        Health
	Disabled      bool
	Builtin       bool
	Removed       bool
	MissingReason *string
	Envelope      *Envelope
}
type Envelope struct {
	ID              string
	ContainerSHA256 string
	VerifiedAt      time.Time
	ValidUntil      *time.Time
}
type Rejection struct{ Reason string }

func (e *Rejection) Error() string { return "reference pack operation rejected: " + e.Reason }
func reject(reason string) error   { return &Rejection{Reason: reason} }

func cloneVersion(v Version) Version {
	if v.Envelope != nil {
		e := *v.Envelope
		if e.ValidUntil != nil {
			stamp := *e.ValidUntil
			e.ValidUntil = &stamp
		}
		v.Envelope = &e
	}
	if v.MissingReason != nil {
		reason := *v.MissingReason
		v.MissingReason = &reason
	}
	v.Manifest.Dependencies = slices.Clone(v.Manifest.Dependencies)
	v.Manifest.Conflicts = slices.Clone(v.Manifest.Conflicts)
	v.Manifest.Artifacts = slices.Clone(v.Manifest.Artifacts)
	v.Manifest.Files = slices.Clone(v.Manifest.Files)
	v.Manifest.Compatibility.Majors = slices.Clone(v.Manifest.Compatibility.Majors)
	v.Manifest.Compatibility.Capabilities = slices.Clone(v.Manifest.Compatibility.Capabilities)
	v.Manifest.License.Notices = slices.Clone(v.Manifest.License.Notices)
	v.Manifest.License.Bindings = slices.Clone(v.Manifest.License.Bindings)
	if v.Manifest.Repository != nil {
		value := *v.Manifest.Repository
		v.Manifest.Repository = &value
	}
	if v.Manifest.SourceAsOf != nil {
		value := *v.Manifest.SourceAsOf
		v.Manifest.SourceAsOf = &value
	}
	if v.Manifest.Summary.Entries != nil {
		value := *v.Manifest.Summary.Entries
		v.Manifest.Summary.Entries = &value
	}
	if v.Manifest.Summary.Objects != nil {
		value := *v.Manifest.Summary.Objects
		v.Manifest.Summary.Objects = &value
	}
	if v.Manifest.Summary.Relationships != nil {
		value := *v.Manifest.Summary.Relationships
		v.Manifest.Summary.Relationships = &value
	}
	if v.Manifest.Extensions != nil {
		v.Manifest.Extensions = cloneJSON(v.Manifest.Extensions).(map[string]any)
	}
	for i, c := range v.Manifest.Conflicts {
		if c.Version != nil {
			version := *c.Version
			v.Manifest.Conflicts[i].Version = &version
		}
	}
	return v
}
func cloneJSON(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := map[string]any{}
		for k, child := range v {
			out[k] = cloneJSON(child)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, child := range v {
			out[i] = cloneJSON(child)
		}
		return out
	default:
		return value
	}
}

// Condition projects health without losing the independent administrative
// disable bit. Invalid or missing content remains visible even while disabled.
func Condition(v Version) string {
	if v.Health == Available && v.Disabled {
		return "disabled"
	}
	return string(v.Health)
}

var ErrRequiredRegistryUnavailable = errors.New("reference pack: required registry unavailable")

// ResolveEffectiveSet applies safety fallback and prunes optional dependency
// chains to a fixed point. Record-compatibility checks belong to the initiating
// normal replacement; emergency fallback never rewrites historical tokens.
func ResolveEffectiveSet(selected []Version, base []Version) ([]Version, []string, error) {
	active := map[string]Version{}
	fallback := map[string]Version{}
	seenBase := map[string]bool{}
	required := map[string]bool{}
	for _, p := range packformat.Profiles() {
		if p.Required {
			required[p.Key] = true
		}
	}
	for _, v := range base {
		if !required[v.Member.Key] || !v.Builtin || len(v.Manifest.Dependencies) != 0 || len(v.Manifest.Conflicts) != 0 {
			return nil, nil, errors.New("reference pack: invalid Base registry")
		}
		if seenBase[v.Member.Key] {
			return nil, nil, errors.New("reference pack: duplicate Base registry")
		}
		seenBase[v.Member.Key] = true
		if healthy(v) {
			fallback[v.Member.Key] = cloneVersion(v)
		}
	}
	if len(seenBase) != len(required) {
		return nil, nil, errors.New("reference pack: incomplete Base registries")
	}
	removed := map[string]bool{}
	for _, v := range selected {
		if _, exists := active[v.Member.Key]; exists {
			return nil, nil, reject("contract_incompatible")
		}
		active[v.Member.Key] = cloneVersion(v)
	}
	changed := true
	for changed {
		changed = false
		for key, v := range active {
			if !healthy(v) {
				delete(active, key)
				removed[key] = true
				changed = true
			}
		}
		for key, v := range fallback {
			if _, exists := active[key]; !exists {
				active[key] = v
				changed = true
			}
		}
		keys := sortedKeys(active)
		for _, k := range keys {
			v := active[k]
			if required[k] && (len(v.Manifest.Dependencies) != 0 || len(v.Manifest.Conflicts) != 0) {
				return nil, nil, reject("required_registry_gap")
			}
			if !required[k] {
				for _, c := range v.Manifest.Conflicts {
					if required[c.Key] {
						return nil, nil, reject("pack_conflict")
					}
				}
			}
			missing := false
			for _, dependency := range v.Manifest.Dependencies {
				target, ok := active[dependency.Key]
				if !ok || target.Member.Version != dependency.Version || target.Member.PayloadSHA256 != dependency.SHA256 {
					missing = true
					break
				}
			}
			if missing {
				if required[k] {
					return nil, nil, reject("required_registry_gap")
				}
				delete(active, k)
				removed[k] = true
				changed = true
			}
		}
	}
	for key := range required {
		if _, ok := active[key]; !ok {
			return nil, nil, ErrRequiredRegistryUnavailable
		}
	}
	// Reject remaining cycles and conflicts after pruning. Sorted roots and
	// dependency edges make the selected failure independent of map traversal.
	visiting, visited := map[string]bool{}, map[string]bool{}
	var visit func(string) bool
	visit = func(k string) bool {
		if visiting[k] {
			return false
		}
		if visited[k] {
			return true
		}
		visiting[k] = true
		v := active[k]
		dependencies := slices.Clone(v.Manifest.Dependencies)
		slices.SortFunc(dependencies, func(a, b packformat.Dependency) int { return strings.Compare(a.Key, b.Key) })
		for _, d := range dependencies {
			if !visit(d.Key) {
				return false
			}
		}
		visiting[k] = false
		visited[k] = true
		return true
	}
	for _, k := range sortedKeys(active) {
		if !visit(k) {
			return nil, nil, reject("dependency_cycle")
		}
	}
	result := []Version{}
	for _, k := range sortedKeys(active) {
		v := active[k]
		for _, c := range v.Manifest.Conflicts {
			if target, exists := active[c.Key]; exists && (c.Version == nil || target.Member.Version == *c.Version) {
				return nil, nil, reject("pack_conflict")
			}
		}
		result = append(result, v)
	}
	pruned := make([]string, 0, len(removed))
	for key := range removed {
		pruned = append(pruned, key)
	}
	slices.Sort(pruned)
	return result, pruned, nil
}
func healthy(v Version) bool {
	return !v.Removed && !v.Disabled && v.Health == Available && v.Envelope != nil
}
func sortedKeys(m map[string]Version) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
