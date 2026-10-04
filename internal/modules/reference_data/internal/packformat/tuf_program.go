package packformat

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// Both historical readers and live attempts use this complete registry range.
// The program has no trust mutation and no network access. Only a successful
// result can become a proposal for the coordinator's publication transaction.
func VerifyTrust(metadataFiles map[string][]byte, inventory Inventory, repository string, at time.Time, snapshot TrustSnapshot) (TrustProposal, error) {
	return verifyTrust(context.Background(), metadataFiles, inventory, repository, at, snapshot, nil, false)
}

func VerifyTrustWithDiagnostics(ctx context.Context, metadataFiles map[string][]byte, inventory Inventory, repository string, at time.Time, snapshot TrustSnapshot, scratch DiagnosticScratch) (TrustProposal, error) {
	return verifyTrust(ctx, metadataFiles, inventory, repository, at, snapshot, scratch, true)
}

type trustProgram struct {
	files      map[string][]byte
	paths      []string
	inventory  Inventory
	repository string
	at         time.Time
	snapshot   TrustSnapshot
	admitted   map[string]metadata
	root       rootMetadata
	proposal   TrustProposal
}

func verifyTrust(ctx context.Context, files map[string][]byte, inventory Inventory, repository string, at time.Time, snapshot TrustSnapshot, scratch DiagnosticScratch, diagnostics bool) (TrustProposal, error) {
	program, result := TrustChecks(files, inventory, repository, at, snapshot)
	if err := runCheckRange(ctx, "tuf_schema", "target_hash", scratch, program, diagnostics); err != nil {
		return TrustProposal{}, err
	}
	return result(), nil
}

// TrustChecks supplies the complete offline-trust range. The proposal accessor
// is meaningful only after every check has succeeded.
func TrustChecks(files map[string][]byte, inventory Inventory, repository string, at time.Time, snapshot TrustSnapshot) (map[string]CheckFunc, func() TrustProposal) {
	p := &trustProgram{files: files, paths: sortedTrustKeys(files), inventory: inventory, repository: repository, at: at, snapshot: snapshot, admitted: map[string]metadata{},
		proposal: TrustProposal{Root: bytes.Clone(snapshot.Root), RootHistory: map[int64][]byte{}, Metadata: map[string]RetainedMetadata{}, Signers: map[string][]string{}, RootTransitions: []RootTransition{}}}
	program := map[string]CheckFunc{
		"tuf_schema": p.schema, "tuf_canonical": p.canonical,
		"root_trust": p.rootTrust, "root_rotation": p.rotation,
		"role_signature": p.signatures, "metadata_rollback": p.rollback,
		"metadata_integrity": p.integrity, "metadata_links": p.links,
		"metadata_expiry": p.expiry, "metadata_expiry_policy": p.expiryPolicy,
		"target_missing": p.targetMissing, "target_extra": p.targetExtra,
		"target_length": p.targetLength, "target_hash": p.targetHash,
	}
	return program, func() TrustProposal { return p.proposal }
}

func sortedTrustKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func (p *trustProgram) role(name string) metadata { return p.admitted["metadata/"+name+".json"] }
func (p *trustProgram) path(name string) string {
	i, _ := slices.BinarySearch(p.paths, name)
	return "$.metadata[" + strconv.Itoa(i) + "]"
}
func (p *trustProgram) rolePath(name string) string { return p.path("metadata/" + name + ".json") }

func (p *trustProgram) schema(ctx context.Context, emit FindingSink) error {
	return trustSchemaFindings(ctx, p.files, emit, p.admitted)
}

func (p *trustProgram) canonical(ctx context.Context, emit FindingSink) error {
	for _, name := range p.paths {
		if err := ctx.Err(); err != nil {
			return err
		}
		data := p.files[name]
		canonical, err := canonicaljson.Canonicalize(data)
		if err != nil || !bytes.Equal(canonical, data) {
			if err := emit(Finding{Path: p.path(name)}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *trustProgram) rootTrust(_ context.Context, emit FindingSink) error {
	root, err := decodeRoot(p.snapshot.Root)
	if err != nil || !bytes.Equal(p.snapshot.RootHistory[root.version], p.snapshot.Root) || root.signed["cartulary"].(map[string]any)["trust_repository_id"] != p.repository {
		return emit(Finding{Path: "metadata/root.json"})
	}
	// Retained transitions may include predecessor-only signatures. They were
	// checked against both roots when retained. The current root must still
	// carry its complete valid successor-authorized quorum before any rotation.
	if len(currentRootSigners(root)) < root.roles["root"].threshold {
		return emit(Finding{Path: "metadata/root.json"})
	}
	p.root = root
	return nil
}

func currentRootSigners(root rootMetadata) []string {
	result := []string{}
	for _, sig := range root.signatures {
		if key, ok := root.roles["root"].keys[sig.key]; ok && ed25519.Verify(key, root.signedBytes, sig.value) {
			result = append(result, sig.key)
		}
	}
	return result
}

func (p *trustProgram) rotation(ctx context.Context, emit FindingSink) error {
	updates := map[int64]string{}
	for _, name := range p.paths {
		if !strings.HasSuffix(name, ".root.json") {
			continue
		}
		token := strings.TrimSuffix(strings.TrimPrefix(name, "metadata/"), ".root.json")
		version, err := strconv.ParseInt(token, 10, 64)
		if err != nil || version < 1 || version > 9007199254740991 || strconv.FormatInt(version, 10) != token {
			if err := emit(Finding{Path: p.path(name)}); err != nil {
				return err
			}
			continue
		}
		updates[version] = name
	}
	versions := make([]int64, 0, len(updates))
	for v := range updates {
		versions = append(versions, v)
	}
	slices.Sort(versions)
	chainValid := true
	for _, version := range versions {
		if err := ctx.Err(); err != nil {
			return err
		}
		name := updates[version]
		data := p.files[name]
		if version <= p.root.version {
			if !bytes.Equal(p.snapshot.RootHistory[version], data) {
				if err := emit(Finding{Path: p.path(name)}); err != nil {
					return err
				}
			}
			continue
		}
		// Later thresholds require the authenticated predecessor. After a gap
		// or rejected transition they are inapplicable, not speculative errors.
		if !chainValid {
			continue
		}
		if version != p.root.version+1 {
			chainValid = false
			if err := emit(Finding{Path: p.path(name)}); err != nil {
				return err
			}
			continue
		}
		next, err := decodeRootShape(data)
		if err != nil || next.version != version || next.signed["cartulary"].(map[string]any)["trust_repository_id"] != p.repository {
			chainValid = false
			if err := emit(Finding{Path: p.path(name)}); err != nil {
				return err
			}
			continue
		}
		sets, valid, err := signatureFindings(next.metadata, p.path(name), emit, p.root.roles["root"], next.roles["root"])
		if err != nil {
			return err
		}
		if !valid {
			chainValid = false
			continue
		}
		p.proposal.RootTransitions = append(p.proposal.RootTransitions, RootTransition{Version: version, PreviousSigners: sets[0], NextSigners: sets[1]})
		p.proposal.RootHistory[version] = bytes.Clone(data)
		p.proposal.Root = bytes.Clone(data)
		p.root = next
	}
	p.proposal.Signers["root"] = currentRootSigners(p.root)
	return nil
}

// Invalid supplied signatures and unsatisfied thresholds are independent
// findings. Every signature is examined; the threshold never short-circuits
// either validation or the complete sorted signer sets.
func signatureFindings(m metadata, path string, emit FindingSink, roles ...role) ([][]string, bool, error) {
	union := map[string]ed25519.PublicKey{}
	sets := make([][]string, len(roles))
	for i, r := range roles {
		sets[i] = []string{}
		for id, key := range r.keys {
			union[id] = key
		}
	}
	valid := true
	for i, sig := range m.signatures {
		key, ok := union[sig.key]
		if !ok || !ed25519.Verify(key, m.signedBytes, sig.value) {
			valid = false
			if err := emit(Finding{Path: path + ".signatures[" + strconv.Itoa(i) + "]"}); err != nil {
				return nil, false, err
			}
			continue
		}
		for j, r := range roles {
			if _, ok := r.keys[sig.key]; ok {
				sets[j] = append(sets[j], sig.key)
			}
		}
	}
	for i, r := range roles {
		if len(sets[i]) < r.threshold {
			valid = false
			if err := emit(Finding{Path: path + ".signatures"}); err != nil {
				return nil, false, err
			}
		}
	}
	return sets, valid, nil
}

func (p *trustProgram) signatures(ctx context.Context, emit FindingSink) error {
	for _, name := range []string{"timestamp", "snapshot", "targets"} {
		if err := ctx.Err(); err != nil {
			return err
		}
		m := p.role(name)
		sets, valid, err := signatureFindings(m, p.rolePath(name), emit, p.root.roles[name])
		if err != nil {
			return err
		}
		if valid {
			p.proposal.Signers[name] = sets[0]
			p.proposal.Metadata[name] = RetainedMetadata{Version: m.version, Bytes: bytes.Clone(m.bytes)}
		}
	}
	return nil
}

func (p *trustProgram) rollback(_ context.Context, emit FindingSink) error {
	for _, name := range []string{"timestamp", "snapshot", "targets"} {
		m := p.role(name)
		if old, ok := p.snapshot.Highest[name]; ok && (m.version < old.Version || m.version == old.Version && !bytes.Equal(m.bytes, old.Bytes)) {
			if err := emit(Finding{Path: p.rolePath(name) + ".signed.version"}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *trustProgram) integrity(ctx context.Context, emit FindingSink) error {
	for _, name := range []string{"timestamp", "snapshot", "targets"} {
		field := "meta"
		if name == "targets" {
			field = "targets"
		}
		descriptors, _ := object(p.role(name).signed[field])
		for i, key := range sortedTrustKeys(descriptors) {
			if err := ctx.Err(); err != nil {
				return err
			}
			if !descriptorComplete(descriptors[key], name != "targets") {
				if err := emit(Finding{Path: p.rolePath(name) + ".signed." + field + "[" + strconv.Itoa(i) + "]"}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (p *trustProgram) links(_ context.Context, emit FindingSink) error {
	for _, link := range []struct{ parent, child string }{{"timestamp", "snapshot"}, {"snapshot", "targets"}} {
		meta, ok := exact(p.role(link.parent).signed["meta"], link.child+".json")
		matches := ok
		if ok {
			descriptor, _ := object(meta[link.child+".json"])
			v, _ := integer(descriptor["version"], 1, 9007199254740991)
			matches = v == p.role(link.child).version && descriptorMatches(descriptor, p.role(link.child).bytes)
		}
		if !matches {
			if err := emit(Finding{Path: p.rolePath(link.parent) + ".signed.meta"}); err != nil {
				return err
			}
		}
	}
	binding, ok := exact(p.role("targets").signed["cartulary"], "schema_id", "trust_repository_id", "pack_key", "pack_version", "pack_release_sequence", "manifest_sha256", "payload_sha256")
	if !ok || binding["schema_id"] != "cartulary.reference_pack_tuf_targets_binding.v1" || binding["trust_repository_id"] != p.repository {
		return emit(Finding{Path: p.rolePath("targets") + ".signed.cartulary"})
	}
	p.proposal.Binding = binding
	return nil
}

func (p *trustProgram) allRoles() []struct {
	name, path string
	value      metadata
	maximum    int64
} {
	return []struct {
		name, path string
		value      metadata
		maximum    int64
	}{
		{"root", "metadata/root.json", p.root.metadata, 63072000},
		{"targets", p.rolePath("targets"), p.role("targets"), 31622400},
		{"snapshot", p.rolePath("snapshot"), p.role("snapshot"), 8035200},
		{"timestamp", p.rolePath("timestamp"), p.role("timestamp"), 2678400},
	}
}
func (p *trustProgram) expiry(_ context.Context, emit FindingSink) error {
	for _, role := range p.allRoles() {
		if !role.value.expires.After(p.at) {
			if err := emit(Finding{Path: role.path}); err != nil {
				return err
			}
		}
	}
	return nil
}
func (p *trustProgram) expiryPolicy(_ context.Context, emit FindingSink) error {
	p.proposal.ValidUntil = p.root.expires
	roles := p.allRoles()
	for i, role := range roles {
		if role.value.expires.Sub(p.at) > time.Duration(role.maximum)*time.Second || i > 0 && role.value.expires.After(roles[i-1].value.expires) {
			if err := emit(Finding{Path: role.path}); err != nil {
				return err
			}
		}
		if role.value.expires.Before(p.proposal.ValidUntil) {
			p.proposal.ValidUntil = role.value.expires
		}
	}
	return nil
}

func (p *trustProgram) targets() map[string]any {
	v, _ := object(p.role("targets").signed["targets"])
	return v
}
func (p *trustProgram) payloadPaths() []string {
	paths := []string{}
	for _, path := range sortedTrustKeys(p.inventory) {
		if !strings.HasPrefix(path, "metadata/") {
			paths = append(paths, path)
		}
	}
	return paths
}
func (p *trustProgram) targetPath(i int) string {
	return p.rolePath("targets") + ".signed.targets[" + strconv.Itoa(i) + "]"
}
func (p *trustProgram) targetMissing(_ context.Context, emit FindingSink) error {
	targets := p.targets()
	for i, path := range p.payloadPaths() {
		if _, ok := targets[path]; !ok {
			if err := emit(Finding{Path: "$.archive_members[" + strconv.Itoa(i) + "]"}); err != nil {
				return err
			}
		}
	}
	return nil
}
func (p *trustProgram) targetExtra(_ context.Context, emit FindingSink) error {
	for i, path := range sortedTrustKeys(p.targets()) {
		if _, ok := p.inventory[path]; !ok || strings.HasPrefix(path, "metadata/") {
			if err := emit(Finding{Path: p.targetPath(i)}); err != nil {
				return err
			}
		}
	}
	return nil
}
func (p *trustProgram) targetLength(_ context.Context, emit FindingSink) error {
	targets := p.targets()
	for i, path := range sortedTrustKeys(targets) {
		d, _ := object(targets[path])
		length, _ := integer(d["length"], 0, 9007199254740991)
		if length != p.inventory[path].Size {
			if err := emit(Finding{Path: p.targetPath(i) + ".length"}); err != nil {
				return err
			}
		}
	}
	return nil
}
func (p *trustProgram) targetHash(_ context.Context, emit FindingSink) error {
	targets := p.targets()
	for i, path := range sortedTrustKeys(targets) {
		d, _ := object(targets[path])
		hashes, _ := object(d["hashes"])
		if hashes["sha256"] != p.inventory[path].SHA256 {
			if err := emit(Finding{Path: p.targetPath(i) + ".hashes.sha256"}); err != nil {
				return err
			}
		}
	}
	return nil
}
