package packformat

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// TrustSnapshot is immutable attempt input. VerifyTrust returns a proposal and
// never mutates it. The application publishes proposals only in the operation's
// successful atomic publication (or for successful members of mixed refresh).
type TrustSnapshot struct {
	Root        []byte                      `json:"root"`
	RootHistory map[int64][]byte            `json:"root_history"`
	Highest     map[string]RetainedMetadata `json:"highest"` // scoped by repository/key/version by caller
}
type RetainedMetadata struct {
	Version int64  `json:"version"`
	Bytes   []byte `json:"bytes"`
}
type TrustProposal struct {
	Root            []byte                      `json:"root"`
	RootHistory     map[int64][]byte            `json:"root_history"`
	Metadata        map[string]RetainedMetadata `json:"metadata"`
	Signers         map[string][]string         `json:"signers"`
	RootTransitions []RootTransition            `json:"root_transitions"`
	ValidUntil      time.Time                   `json:"valid_until"`
	Binding         map[string]any              `json:"binding"`
}
type RootTransition struct {
	Version         int64    `json:"version"`
	PreviousSigners []string `json:"previous_signers"`
	NextSigners     []string `json:"next_signers"`
}
type signature struct {
	key   string
	value []byte
}
type metadata struct {
	signed      map[string]any
	signedBytes []byte
	bytes       []byte
	signatures  []signature
	version     int64
	expires     time.Time
}
type role struct {
	keys      map[string]ed25519.PublicKey
	threshold int
}
type rootMetadata struct {
	metadata
	roles map[string]role
}

func object(value any) (map[string]any, bool) { v, ok := value.(map[string]any); return v, ok }
func exact(value any, names ...string) (map[string]any, bool) {
	m, ok := object(value)
	if !ok || len(m) != len(names) {
		return nil, false
	}
	for _, name := range names {
		if _, ok := m[name]; !ok {
			return nil, false
		}
	}
	return m, true
}
func integer(value any, min, max int64) (int64, bool) {
	n, ok := value.(json.Number)
	if !ok {
		return 0, false
	}
	text := n.String()
	if text == "" || strings.ContainsAny(text, ".eE+") || (len(text) > 1 && text[0] == '0') || text == "-0" {
		return 0, false
	}
	v, err := strconv.ParseInt(text, 10, 64)
	return v, err == nil && v >= min && v <= max
}
func hexBytes(value any, size int) ([]byte, bool) {
	s, ok := value.(string)
	if !ok || len(s) != size*2 || s != strings.ToLower(s) {
		return nil, false
	}
	b, err := hex.DecodeString(s)
	return b, err == nil
}
func equalDigest(data []byte, value any) bool {
	want, ok := hexBytes(value, sha256.Size)
	if !ok {
		return false
	}
	sum := sha256.Sum256(data)
	return subtle.ConstantTimeCompare(sum[:], want) == 1
}
func validPackNumbers(value any) bool {
	switch v := value.(type) {
	case json.Number:
		_, ok := integer(v, -9007199254740991, 9007199254740991)
		return ok
	case map[string]any:
		for _, child := range v {
			if !validPackNumbers(child) {
				return false
			}
		}
	case []any:
		for _, child := range v {
			if !validPackNumbers(child) {
				return false
			}
		}
	}
	return true
}
func decodeMetadataShape(data []byte, kind string, maxSignatures int) (metadata, error) {
	m, err := metadataSchemaFindings(data, kind, maxSignatures, "$", func(Finding) error { return errShapeMismatch })
	if err != nil {
		if errors.Is(err, errShapeMismatch) {
			return metadata{}, fail("tuf_metadata_invalid")
		}
		return metadata{}, err
	}
	return metadataFromFields(data, m)
}

func metadataFromFields(data []byte, m map[string]any) (metadata, error) {
	signed := m["signed"].(map[string]any)
	version, _ := integer(signed["version"], 1, 9007199254740991)
	expires, _ := time.Parse("2006-01-02T15:04:05Z", signed["expires"].(string))
	signatures := []signature{}
	for _, item := range m["signatures"].([]any) {
		fields := item.(map[string]any)
		value, _ := hexBytes(fields["sig"], 64)
		signatures = append(signatures, signature{fields["keyid"].(string), value})
	}
	encoded, err := canonicaljson.Marshal(signed)
	if err != nil {
		return metadata{}, err
	}
	return metadata{signed: signed, signedBytes: encoded, bytes: bytes.Clone(data), signatures: signatures, version: version, expires: expires}, nil
}

func descriptorComplete(value any, metadataDescriptor bool) bool {
	fields, ok := object(value)
	if !ok {
		return false
	}
	if _, ok := fields["length"]; !ok {
		return false
	}
	if metadataDescriptor {
		if _, ok := fields["version"]; !ok {
			return false
		}
	}
	hashes, ok := object(fields["hashes"])
	if !ok {
		return false
	}
	_, ok = hashes["sha256"]
	return ok
}

func decodeRoot(data []byte) (rootMetadata, error) {
	root, err := decodeRootShape(data)
	if err != nil {
		return rootMetadata{}, err
	}
	canonical, err := canonicaljson.Canonicalize(data)
	if err != nil || !bytes.Equal(canonical, data) {
		return rootMetadata{}, fail("metadata_noncanonical")
	}
	return root, nil
}

func decodeRootShape(data []byte) (rootMetadata, error) {
	m, err := decodeMetadataShape(data, "root", 128)
	if err != nil {
		return rootMetadata{}, err
	}
	if m.signed["consistent_snapshot"] != false {
		return rootMetadata{}, fail("tuf_root_untrusted")
	}
	binding, ok := exact(m.signed["cartulary"], "schema_id", "trust_repository_id")
	if !ok || binding["schema_id"] != "cartulary.reference_pack_tuf_root_binding.v1" {
		return rootMetadata{}, fail("tuf_root_untrusted")
	}
	repository, ok := binding["trust_repository_id"].(string)
	if !ok || !keyPattern.MatchString(repository) || len(repository) > 128 {
		return rootMetadata{}, fail("tuf_root_untrusted")
	}
	keyObjects, ok := object(m.signed["keys"])
	if !ok || len(keyObjects) > 64 {
		return rootMetadata{}, fail("tuf_metadata_invalid")
	}
	keys := map[string]ed25519.PublicKey{}
	for id, value := range keyObjects {
		key, ok := exact(value, "keytype", "scheme", "keyval")
		if !ok || key["keytype"] != "ed25519" || key["scheme"] != "ed25519" {
			return rootMetadata{}, fail("tuf_root_untrusted")
		}
		keyval, ok := exact(key["keyval"], "public")
		if !ok {
			return rootMetadata{}, fail("tuf_root_untrusted")
		}
		public, ok := hexBytes(keyval["public"], ed25519.PublicKeySize)
		if !ok {
			return rootMetadata{}, fail("tuf_root_untrusted")
		}
		encoded, err := canonicaljson.Marshal(key)
		if err != nil || !equalDigest(encoded, id) {
			return rootMetadata{}, fail("tuf_root_untrusted")
		}
		keys[id] = ed25519.PublicKey(public)
	}
	roles, ok := exact(m.signed["roles"], "root", "timestamp", "snapshot", "targets")
	if !ok {
		return rootMetadata{}, fail("tuf_root_untrusted")
	}
	result := rootMetadata{metadata: m, roles: map[string]role{}}
	for _, name := range []string{"root", "timestamp", "snapshot", "targets"} {
		fields, ok := exact(roles[name], "keyids", "threshold")
		if !ok {
			return rootMetadata{}, fail("tuf_root_untrusted")
		}
		ids, ok := fields["keyids"].([]any)
		if !ok || len(ids) < 1 || len(ids) > 64 {
			return rootMetadata{}, fail("tuf_root_untrusted")
		}
		minimum := int64(1)
		if name == "root" {
			minimum = 2
			if len(ids) < 3 {
				return rootMetadata{}, fail("tuf_root_untrusted")
			}
		}
		threshold, ok := integer(fields["threshold"], minimum, int64(len(ids)))
		if !ok {
			return rootMetadata{}, fail("tuf_root_untrusted")
		}
		r := role{keys: map[string]ed25519.PublicKey{}, threshold: int(threshold)}
		previous := ""
		for _, value := range ids {
			id, ok := value.(string)
			if !ok || id <= previous || keys[id] == nil {
				return rootMetadata{}, fail("tuf_root_untrusted")
			}
			r.keys[id] = keys[id]
			previous = id
		}
		result.roles[name] = r
	}
	return result, nil
}

func verifySignatures(m metadata, roles ...role) ([][]string, error) {
	union := map[string]ed25519.PublicKey{}
	for _, r := range roles {
		for id, key := range r.keys {
			union[id] = key
		}
	}
	sets := make([][]string, len(roles))
	for i := range sets {
		sets[i] = []string{}
	}
	for _, signature := range m.signatures {
		key, ok := union[signature.key]
		if !ok || !ed25519.Verify(key, m.signedBytes, signature.value) {
			return nil, fail("signature_threshold_not_met")
		}
		for i, r := range roles {
			if _, ok := r.keys[signature.key]; ok {
				sets[i] = append(sets[i], signature.key)
			}
		}
	}
	for i, r := range roles {
		if len(sets[i]) < r.threshold {
			return nil, fail("signature_threshold_not_met")
		}
	}
	return sets, nil
}

// AdmitBootstrap validates the exact out-of-band root. It does not confer trust
// merely because a container supplies a self-signed root.
func AdmitBootstrap(data []byte) (TrustSnapshot, error) {
	root, err := decodeRoot(data)
	if err != nil {
		return TrustSnapshot{}, fail("tuf_root_untrusted")
	}
	if len(root.signatures) > 64 {
		return TrustSnapshot{}, fail("tuf_root_untrusted")
	}
	if _, err := verifySignatures(root.metadata, root.roles["root"]); err != nil {
		return TrustSnapshot{}, fail("tuf_root_untrusted")
	}
	return TrustSnapshot{Root: bytes.Clone(data), RootHistory: map[int64][]byte{root.version: bytes.Clone(data)}, Highest: map[string]RetainedMetadata{}}, nil
}

// VerifyHistoricalRoot authenticates a retained root against its predecessor.
// The first root is the retained out-of-band bootstrap. Historical linkage has
// no wall-clock freshness check: freshness belongs to successful envelopes.
func VerifyHistoricalRoot(previous, data []byte, repository string) (RootTransition, error) {
	next, err := decodeRoot(data)
	if err != nil || next.signed["cartulary"].(map[string]any)["trust_repository_id"] != repository {
		return RootTransition{}, fail("tuf_root_untrusted")
	}
	if previous == nil {
		if _, err := AdmitBootstrap(data); err != nil {
			return RootTransition{}, err
		}
		return RootTransition{Version: next.version, PreviousSigners: []string{}, NextSigners: []string{}}, nil
	}
	prior, err := decodeRoot(previous)
	if err != nil || prior.signed["cartulary"].(map[string]any)["trust_repository_id"] != repository || next.version != prior.version+1 {
		return RootTransition{}, fail("tuf_root_rotation_invalid")
	}
	sets, err := verifySignatures(next.metadata, prior.roles["root"], next.roles["root"])
	if err != nil {
		return RootTransition{}, fail("tuf_root_rotation_invalid")
	}
	return RootTransition{Version: next.version, PreviousSigners: sets[0], NextSigners: sets[1]}, nil
}

func descriptorMatches(descriptor map[string]any, data []byte) bool {
	length, ok := integer(descriptor["length"], 0, 9007199254740991)
	if !ok || length != int64(len(data)) {
		return false
	}
	hashes, ok := exact(descriptor["hashes"], "sha256")
	return ok && equalDigest(data, hashes["sha256"])
}
