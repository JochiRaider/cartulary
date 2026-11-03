// Package registryfacts owns logical fact identity and canonical owner-input
// ordering. It consumes typed JSON projections, never generated packages or
// documents. Callers retain full schema, owner-allocation and recognition gates.
package registryfacts

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const IdentitySchemaID = "cartulary.extension_owner_fact_identity.v2"

var (
	ErrInvalidFact      = errors.New("invalid owner fact identity")
	ErrInvalidFragment  = errors.New("invalid owner fragment")
	profilePattern      = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
	qualifiedPattern    = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}\.[a-z][a-z0-9_]{0,63}$`)
	publicSchemaPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,159}$`)
	routeLiteralPattern = regexp.MustCompile(`^[A-Za-z0-9._~-]+$`)
)

// Identity is constructed only from an admitted variant's logical key. Full
// values, fragments and physical names do not distinguish duplicate facts.
type Identity struct {
	profile, kind, canonical string
}

func (i Identity) CanonicalBytes() []byte { return []byte(i.canonical) }

// Key orders the profile/kind/identity tuple. Canonical JSON contains no literal
// NUL, so separators cannot collide with any component or affect tuple order.
func (i Identity) Key() string { return i.profile + "\x00" + i.kind + "\x00" + i.canonical }

func (i Identity) Token() string {
	sum := sha256.Sum256([]byte(i.canonical))
	return i.kind + ":sha256:" + hex.EncodeToString(sum[:])
}

func Identify(fact map[string]any) (Identity, error) {
	profile, _ := fact["profile_id"].(string)
	kind, _ := fact["fact_kind"].(string)
	if !profilePattern.MatchString(profile) {
		return Identity{}, ErrInvalidFact
	}
	members := map[string]any{}
	key, body := "", fact
	switch kind {
	case "recognized_profile", "claim_configuration", "state_ownership", "admission_validation", "egress", "portability", "snapshot_reporting":
	case "runtime_dependency":
		key = "dependency_profile_id"
	case "route_family", "public_schema", "worker_kind":
		key = kind
		if kind == "public_schema" {
			key = "public_schema_id"
		}
	case "workspace":
		key = "workspace_key"
	case "contribution":
		body, _ = fact["contribution"].(map[string]any)
		key = "contribution_id"
	case "migration_definition", "migration_ledger_definition":
		body, _ = fact[kind].(map[string]any)
		key = "migration_id"
	case "job_kind":
		body, _ = fact["job_kind_contract"].(map[string]any)
		key = "job_kind"
	default:
		return Identity{}, ErrInvalidFact
	}
	if key != "" {
		value, ok := body[key].(string)
		if !ok || !validMember(key, value) {
			return Identity{}, ErrInvalidFact
		}
		members[key] = value
	}
	encoded, err := canonicaljson.MarshalExtension(map[string]any{
		"schema_id": IdentitySchemaID, "profile_id": profile, "fact_kind": kind, "identity": members,
	})
	if err != nil {
		return Identity{}, ErrInvalidFact
	}
	return Identity{profile: profile, kind: kind, canonical: string(encoded)}, nil
}

func validMember(key, value string) bool {
	switch key {
	case "dependency_profile_id", "workspace_key":
		return profilePattern.MatchString(value)
	case "public_schema_id":
		return publicSchemaPattern.MatchString(value)
	case "route_family":
		_, valid := RouteSegments(value)
		return valid
	default:
		return qualifiedPattern.MatchString(value)
	}
}

// RouteSegments admits the complete route_family_template_v1 scalar. Parameter
// names remain exact in identity bytes and are neutral only during overlap tests.
func RouteSegments(route string) ([]string, bool) {
	if len(route) > 512 || !strings.HasPrefix(route, "/api/v1/") {
		return nil, false
	}
	segments := strings.Split(route[1:], "/")
	if len(segments) > 32 {
		return nil, false
	}
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return nil, false
		}
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			if !profilePattern.MatchString(segment[1 : len(segment)-1]) {
				return nil, false
			}
		} else if !routeLiteralPattern.MatchString(segment) {
			return nil, false
		}
	}
	return segments, true
}

type orderedObject struct {
	object         map[string]any
	key, canonical string
}

// normalizeFacts returns isolated, ordered objects and preserves duplicate
// declarations. Equal logical identities use full canonical fact bytes as their
// tie-break, without adding those values to the identity hash.
func normalizeFacts(facts []map[string]any) ([]map[string]any, error) {
	ordered := make([]orderedObject, len(facts))
	for index, fact := range facts {
		identity, err := Identify(fact)
		if err != nil {
			return nil, err
		}
		body, err := canonicaljson.MarshalExtension(fact)
		if err != nil {
			return nil, ErrInvalidFact
		}
		copy, err := canonicaljson.DecodeExtension(body)
		if err != nil {
			return nil, ErrInvalidFact
		}
		ordered[index] = orderedObject{object: copy.(map[string]any), key: identity.Key(), canonical: string(body)}
	}
	return orderedObjects(ordered), nil
}

// NormalizeFragments first normalizes each fact array, then orders fragments by
// owner, fragment ID and full normalized fragment bytes. Empty input stays an
// empty array; empty fact arrays remain explicit fragments. Malformed input is
// never repaired or discarded.
func NormalizeFragments(fragments []map[string]any) ([]map[string]any, error) {
	ordered := make([]orderedObject, len(fragments))
	for index, fragment := range fragments {
		owner, _ := fragment["owner_id"].(string)
		id, _ := fragment["owner_fragment_id"].(string)
		if len(fragment) != 4 || fragment["schema_id"] != "cartulary.extension_owner_fragment.v3" || !profilePattern.MatchString(owner) || !qualifiedPattern.MatchString(id) || !strings.HasPrefix(id, owner+".") {
			return nil, ErrInvalidFragment
		}
		raw, ok := fragment["facts"].([]any)
		if !ok {
			return nil, ErrInvalidFragment
		}
		facts := make([]map[string]any, len(raw))
		for i, value := range raw {
			facts[i], ok = value.(map[string]any)
			if !ok {
				return nil, ErrInvalidFact
			}
		}
		normalized, err := normalizeFacts(facts)
		if err != nil {
			return nil, err
		}
		rows := make([]any, len(normalized))
		for i, fact := range normalized {
			rows[i] = fact
		}
		copy := map[string]any{"schema_id": fragment["schema_id"], "owner_id": owner, "owner_fragment_id": id, "facts": rows}
		body, err := canonicaljson.MarshalExtension(copy)
		if err != nil {
			return nil, ErrInvalidFragment
		}
		ordered[index] = orderedObject{object: copy, key: owner + "\x00" + id, canonical: string(body)}
	}
	return orderedObjects(ordered), nil
}

func orderedObjects(ordered []orderedObject) []map[string]any {
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].key != ordered[j].key {
			return ordered[i].key < ordered[j].key
		}
		return ordered[i].canonical < ordered[j].canonical
	})
	result := make([]map[string]any, len(ordered))
	for i, object := range ordered {
		result[i] = object.object
	}
	return result
}

// ValidateIdentityDefinition requires the current transient contract exactly
// once, and forbids retired or unknown generations in its definition family.
func ValidateIdentityDefinition(set map[string]any) error {
	definitions, ok := set["definitions"].([]any)
	if !ok {
		return ErrInvalidFact
	}
	found := false
	for _, raw := range definitions {
		definition, ok := raw.(map[string]any)
		if !ok {
			return ErrInvalidFact
		}
		id, _ := definition["schema_id"].(string)
		if !strings.HasPrefix(id, "cartulary.extension_owner_fact_identity.") {
			continue
		}
		if id != IdentitySchemaID || found || definition["owner_id"] != "extensions" || definition["artifact_class"] != "generated_input" {
			return ErrInvalidFact
		}
		found = true
	}
	if !found {
		return ErrInvalidFact
	}
	return nil
}
