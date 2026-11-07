package producer

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type SigningContext struct {
	SchemaID   string            `json:"schema_id"`
	Repository string            `json:"trust_repository_id"`
	Seeds      map[string]string `json:"private_seeds"`
}

type EnvelopeInput struct {
	SchemaID             string `json:"schema_id"`
	LogicalRoot          string `json:"logical_root"`
	OutputRoot           string `json:"output_root"`
	SigningContext       string `json:"signing_context"`
	CreateSigningContext bool   `json:"create_signing_context"`
	IssuedAt             string `json:"issued_at"`
	RootExpires          string `json:"root_expires"`
	OnlineExpires        string `json:"online_expires"`
	MetadataVersion      int64  `json:"metadata_version"`
	LedgerRoot           string `json:"ledger_root"`
	PreviousEnvelope     string `json:"previous_envelope"`
}

type EnvelopeReport struct {
	SchemaID             string  `json:"schema_id"`
	Repository           string  `json:"trust_repository_id"`
	ManifestSHA256       string  `json:"manifest_sha256"`
	PayloadSHA256        string  `json:"payload_sha256"`
	ContainerSHA256      string  `json:"container_sha256"`
	ContainerBytes       int64   `json:"container_bytes"`
	BootstrapSHA256      string  `json:"bootstrap_sha256"`
	MetadataVersion      int64   `json:"metadata_version"`
	IssuedAt             string  `json:"issued_at"`
	EnvelopePath         string  `json:"envelope_path"`
	PreviousReportSHA256 *string `json:"previous_report_sha256"`
}

var signingRoles = []string{"root-1", "root-2", "root-3", "targets", "snapshot", "timestamp"}

type publicationPolicy struct {
	Repository, Key, Prefix, Version string
	Fixture                          bool
}

func (p publicationPolicy) contextSchema() string {
	if p.Fixture {
		return p.Prefix + "_fixture_signing_context." + p.Version
	}
	return p.Prefix + "_signing_context." + p.Version
}
func (p publicationPolicy) allocationSchema() string {
	if p.Fixture {
		return p.Prefix + "_fixture_envelope_allocation." + p.Version
	}
	return p.Prefix + "_envelope_allocation." + p.Version
}

var attackFixturePolicy = publicationPolicy{FixtureRepository, "framework.attack", "cartulary.attack", "v1", true}
var windowsFixturePolicy = publicationPolicy{WindowsFixtureRepository, WindowsKey, "cartulary.windows_event", "v2", true}
var entraFixturePolicy = publicationPolicy{EntraFixtureRepository, EntraKey, "cartulary.entra", "v2", true}
var windowsReleasePolicy = publicationPolicy{WindowsReleaseRepository, WindowsKey, "cartulary.windows_event_release", "v2", false}
var entraReleasePolicy = publicationPolicy{EntraRepository, EntraKey, "cartulary.entra_release", "v2", false}

func fixturePolicyForRepository(repository string) (publicationPolicy, error) {
	for _, p := range []publicationPolicy{attackFixturePolicy, windowsFixturePolicy, entraFixturePolicy} {
		if p.Repository == repository {
			return p, nil
		}
	}
	return publicationPolicy{}, errors.New("isolated fixture repository required")
}

// NewFixtureSigningContext creates local fixture authority, never deployment
// authority. Callers must keep these bytes outside runtime and public evidence.
func NewFixtureSigningContext() (SigningContext, error) {
	return newSigningContext(attackFixturePolicy)
}

// NewWindowsFixtureSigningContext creates isolated test authority only.
func NewWindowsFixtureSigningContext() (SigningContext, error) {
	return newSigningContext(windowsFixturePolicy)
}

// NewEntraFixtureSigningContext creates isolated authority for canonical corpus tests.
func NewEntraFixtureSigningContext() (SigningContext, error) {
	return newSigningContext(entraFixturePolicy)
}

func newSigningContext(policy publicationPolicy) (SigningContext, error) {
	c := SigningContext{SchemaID: policy.contextSchema(), Repository: policy.Repository, Seeds: map[string]string{}}
	for _, role := range signingRoles {
		seed := make([]byte, ed25519.SeedSize)
		if _, err := rand.Read(seed); err != nil {
			return c, err
		}
		c.Seeds[role] = hex.EncodeToString(seed)
	}
	return c, nil
}

func loadSigningContext(path string, create bool, policy publicationPolicy) (SigningContext, error) {
	var c SigningContext
	b, err := readRegular(path)
	if errors.Is(err, os.ErrNotExist) && create {
		c, err = newSigningContext(policy)
		if err != nil {
			return c, err
		}
		b, err = canonicaljson.Marshal(c)
		if err != nil {
			return c, err
		}
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return c, err
		}
		root, e := os.OpenRoot(filepath.Dir(path))
		if e != nil {
			return c, e
		}
		err = Retain(root, filepath.Base(path), b, 0600)
		root.Close()
		if err != nil {
			return c, err
		}
	} else if err != nil {
		return c, err
	}
	if err = DecodeClosed(b, &c); err != nil {
		return c, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return c, err
	}
	if info.Mode().Perm()&0077 != 0 || c.SchemaID != policy.contextSchema() || c.Repository != policy.Repository || len(c.Seeds) != len(signingRoles) {
		return c, errors.New("private signing context must match the selected publisher policy")
	}
	seen := map[string]bool{}
	for _, role := range signingRoles {
		seed, err := hex.DecodeString(c.Seeds[role])
		if err != nil || len(seed) != ed25519.SeedSize || seen[c.Seeds[role]] {
			return c, errors.New("distinct signing keys required")
		}
		seen[c.Seeds[role]] = true
	}
	return c, nil
}

// FixtureEnvelope shares production canonical/digest/archive/trust primitives.
// The caller supplies all signing times and versions. Verification is a separate
// production operation with a fresh trusted wall clock.
func FixtureEnvelope(ctx context.Context, members map[string][]byte, c SigningContext, rootExpires, onlineExpires string, version int64) ([]byte, []byte, error) {
	policy, err := fixturePolicyForRepository(c.Repository)
	if err != nil {
		return nil, nil, err
	}
	return buildEnvelope(ctx, members, c, policy, rootExpires, onlineExpires, version)
}

func buildEnvelope(ctx context.Context, members map[string][]byte, c SigningContext, policy publicationPolicy, rootExpires, onlineExpires string, version int64) ([]byte, []byte, error) {
	manifest, err := packformat.DecodeManifest(members["manifest.json"], true)
	if err != nil {
		return nil, nil, err
	}
	if manifest.Repository == nil || *manifest.Repository != policy.Repository || manifest.Key != policy.Key || c.Repository != policy.Repository || c.SchemaID != policy.contextSchema() || len(c.Seeds) != len(signingRoles) {
		return nil, nil, errors.New("matching publisher and pack identity required")
	}
	if policy.Key == "framework.attack" && manifest.License.Redistribution != "restricted" {
		return nil, nil, errors.New("ATT&CK fixture remains restricted")
	}
	payload, err := packformat.PayloadDigest(manifest.Files)
	if err != nil {
		return nil, nil, err
	}
	keys := map[string]any{}
	ids := map[string]string{}
	private := map[string]ed25519.PrivateKey{}
	for _, role := range signingRoles {
		seed, err := hex.DecodeString(c.Seeds[role])
		if err != nil || len(seed) != 32 {
			return nil, nil, errors.New("invalid signing key")
		}
		key := ed25519.NewKeyFromSeed(seed)
		public := map[string]any{"keytype": "ed25519", "scheme": "ed25519", "keyval": map[string]any{"public": hex.EncodeToString(key.Public().(ed25519.PublicKey))}}
		b, err := canonicaljson.Marshal(public)
		if err != nil {
			return nil, nil, err
		}
		id := Digest(b)
		if _, exists := keys[id]; exists {
			return nil, nil, errors.New("distinct signing keys required")
		}
		keys[id] = public
		ids[role] = id
		private[id] = key
	}
	sign := func(signed map[string]any, signers []string) ([]byte, error) {
		b, err := canonicaljson.Marshal(signed)
		if err != nil {
			return nil, err
		}
		slices.Sort(signers)
		sigs := []any{}
		for _, id := range signers {
			sigs = append(sigs, map[string]any{"keyid": id, "sig": hex.EncodeToString(ed25519.Sign(private[id], b))})
		}
		return canonicaljson.Marshal(map[string]any{"signed": signed, "signatures": sigs})
	}
	rootIDs := []string{ids["root-1"], ids["root-2"], ids["root-3"]}
	slices.Sort(rootIDs)
	roles := map[string]any{"root": map[string]any{"keyids": rootIDs, "threshold": 2}}
	for _, role := range []string{"targets", "snapshot", "timestamp"} {
		roles[role] = map[string]any{"keyids": []string{ids[role]}, "threshold": 1}
	}
	root, err := sign(map[string]any{"_type": "root", "spec_version": "1.0.35", "version": 1, "expires": rootExpires, "consistent_snapshot": false, "keys": keys, "roles": roles, "cartulary": map[string]any{"schema_id": "cartulary.reference_pack_tuf_root_binding.v1", "trust_repository_id": policy.Repository}}, []string{ids["root-1"], ids["root-2"]})
	if err != nil {
		return nil, nil, err
	}
	files := map[string][]byte{"manifest.json": members["manifest.json"]}
	for _, f := range manifest.Files {
		b, ok := members[f.Path]
		if !ok || int64(len(b)) != f.Size || Digest(b) != f.SHA256 {
			return nil, nil, errors.New("logical member integrity mismatch")
		}
		files[f.Path] = b
	}
	files["bundle.json"], err = canonicaljson.Marshal(map[string]any{"schema_id": "cartulary.reference_pack_bundle_hint.v1", "trust_repository_id": policy.Repository})
	if err != nil {
		return nil, nil, err
	}
	descriptor := func(b []byte) map[string]any {
		return map[string]any{"length": len(b), "hashes": map[string]any{"sha256": Digest(b)}}
	}
	targets := map[string]any{}
	for name, b := range files {
		targets[name] = descriptor(b)
	}
	binding := map[string]any{"schema_id": "cartulary.reference_pack_tuf_targets_binding.v1", "trust_repository_id": policy.Repository, "pack_key": manifest.Key, "pack_version": manifest.Version, "pack_release_sequence": manifest.Sequence, "manifest_sha256": Digest(files["manifest.json"]), "payload_sha256": payload}
	files["metadata/targets.json"], err = sign(map[string]any{"_type": "targets", "spec_version": "1.0.35", "version": version, "expires": onlineExpires, "targets": targets, "cartulary": binding}, []string{ids["targets"]})
	if err != nil {
		return nil, nil, err
	}
	for _, role := range []string{"snapshot", "timestamp"} {
		child := "targets.json"
		if role == "timestamp" {
			child = "snapshot.json"
		}
		d := descriptor(files["metadata/"+child])
		d["version"] = version
		files["metadata/"+role+".json"], err = sign(map[string]any{"_type": role, "spec_version": "1.0.35", "version": version, "expires": onlineExpires, "meta": map[string]any{child: d}}, []string{ids[role]})
		if err != nil {
			return nil, nil, err
		}
	}
	bootstrap, err := canonicaljson.Marshal(map[string]any{"schema_id": "cartulary.reference_pack_trust_bootstrap.v1", "repositories": []any{map[string]any{"repository_id": policy.Repository, "trusted_root": json.RawMessage(root), "trusted_root_sha256": Digest(root)}}})
	if err != nil {
		return nil, nil, err
	}
	if _, err := packformat.DecodeBootstrap(bootstrap); err != nil {
		return nil, nil, err
	}
	var container bytes.Buffer
	if err = packformat.WriteUSTAR(ctx, &container, files); err != nil {
		return nil, nil, err
	}
	return container.Bytes(), bootstrap, nil
}

func Envelope(ctx context.Context, inputBytes []byte, renew bool) (EnvelopeReport, error) {
	var in EnvelopeInput
	if err := DecodeClosed(inputBytes, &in); err != nil {
		return EnvelopeReport{}, err
	}
	policy := attackFixturePolicy
	if in.SchemaID == "cartulary.entra_envelope_input.v2" {
		policy = entraFixturePolicy
	} else if in.SchemaID == "cartulary.entra_release_envelope_input.v2" {
		policy = entraReleasePolicy
	} else if in.SchemaID == "cartulary.windows_event_envelope_input.v2" {
		policy = windowsFixturePolicy
	} else if in.SchemaID == "cartulary.windows_event_release_envelope_input.v2" {
		policy = windowsReleasePolicy
		if err := windowsDecode(inputBytes, "release_envelope", &in); err != nil {
			return EnvelopeReport{}, err
		}
	}
	if in.SchemaID != policy.Prefix+"_envelope_input."+policy.Version || in.LedgerRoot == "" || in.MetadataVersion < 1 || in.MetadataVersion > 9007199254740991 || !validBuildTime(in.IssuedAt) || !validBuildTime(in.RootExpires) || !validBuildTime(in.OnlineExpires) || in.OnlineExpires <= in.IssuedAt || in.RootExpires < in.OnlineExpires || renew != (in.PreviousEnvelope != "") {
		return EnvelopeReport{}, errors.New("invalid explicit envelope inputs")
	}
	for _, public := range []string{in.OutputRoot, in.LogicalRoot, in.LedgerRoot} {
		if err := signingOutside(in.SigningContext, public); err != nil {
			return EnvelopeReport{}, err
		}
	}
	b, err := readRegular(filepath.Join(in.LogicalRoot, "manifest.json"))
	if err != nil {
		return EnvelopeReport{}, err
	}
	m, err := packformat.DecodeManifest(b, true)
	if err != nil {
		return EnvelopeReport{}, err
	}
	members := map[string][]byte{"manifest.json": b}
	for _, f := range m.Files {
		b, err := readRegular(filepath.Join(in.LogicalRoot, filepath.FromSlash(f.Path)))
		if err != nil {
			return EnvelopeReport{}, err
		}
		members[f.Path] = b
	}
	if m.Repository == nil || *m.Repository != policy.Repository || m.Key != policy.Key {
		return EnvelopeReport{}, errors.New("logical pack does not match publisher policy")
	}
	c, err := loadSigningContext(in.SigningContext, in.CreateSigningContext, policy)
	if err != nil {
		return EnvelopeReport{}, err
	}
	container, bootstrap, err := buildEnvelope(ctx, members, c, policy, in.RootExpires, in.OnlineExpires, in.MetadataVersion)
	if err != nil {
		return EnvelopeReport{}, err
	}
	payload, err := packformat.PayloadDigest(m.Files)
	if err != nil {
		return EnvelopeReport{}, err
	}
	r := EnvelopeReport{SchemaID: policy.Prefix + "_envelope_report." + policy.Version, Repository: policy.Repository, ManifestSHA256: Digest(b), PayloadSHA256: payload, ContainerSHA256: Digest(container), ContainerBytes: int64(len(container)), BootstrapSHA256: Digest(bootstrap), MetadataVersion: in.MetadataVersion, IssuedAt: in.IssuedAt, EnvelopePath: "envelopes/" + Digest(container)}
	if renew {
		previousBytes, err := readRegular(filepath.Join(in.PreviousEnvelope, "envelope-report.json"))
		if err != nil {
			return r, err
		}
		var previous EnvelopeReport
		if err = DecodeClosed(previousBytes, &previous); err != nil {
			return r, err
		}
		if previous.SchemaID != r.SchemaID || previous.ManifestSHA256 != r.ManifestSHA256 || previous.PayloadSHA256 != r.PayloadSHA256 || previous.BootstrapSHA256 != r.BootstrapSHA256 || previous.MetadataVersion >= r.MetadataVersion || previous.IssuedAt > r.IssuedAt {
			return r, errors.New("renewal must preserve logical bytes and root while advancing metadata")
		}
		priorContainer, err := readRegular(filepath.Join(in.PreviousEnvelope, "container.pack"))
		if err != nil || Digest(priorContainer) != previous.ContainerSHA256 || int64(len(priorContainer)) != previous.ContainerBytes {
			return r, errors.New("prior envelope integrity mismatch")
		}
		priorBootstrap, err := readRegular(filepath.Join(in.PreviousEnvelope, "bootstrap.json"))
		if err != nil || Digest(priorBootstrap) != previous.BootstrapSHA256 {
			return r, errors.New("prior bootstrap integrity mismatch")
		}
		d := Digest(previousBytes)
		r.PreviousReportSHA256 = &d
	}
	// Metadata versions are scoped by immutable pack identity, matching the
	// production trust store. Immutable reservations also survive failed writes.
	reservation, err := canonicaljson.Marshal(map[string]any{"schema_id": policy.allocationSchema(), "trust_repository_id": policy.Repository, "manifest_sha256": r.ManifestSHA256, "metadata_version": r.MetadataVersion, "container_sha256": r.ContainerSHA256, "bootstrap_sha256": r.BootstrapSHA256, "issued_at": r.IssuedAt})
	if err != nil {
		return r, err
	}
	if err = os.MkdirAll(in.LedgerRoot, 0700); err != nil {
		return r, err
	}
	ledger, err := os.OpenRoot(in.LedgerRoot)
	if err != nil {
		return r, err
	}
	err = Retain(ledger, r.ManifestSHA256+"/"+fmt.Sprintf("%016d.json", r.MetadataVersion), reservation, 0600)
	ledger.Close()
	if err != nil {
		return r, err
	}
	reportBytes, err := canonicaljson.Marshal(r)
	if err != nil {
		return r, err
	}
	if !policy.Fixture {
		if err = verifyReleaseEnvelope(ctx, container, bootstrap, r, in.LedgerRoot, policy); err != nil {
			return r, err
		}
	}
	if err = publishTree(in.OutputRoot, r.EnvelopePath, map[string][]byte{"container.pack": container, "bootstrap.json": bootstrap, "envelope-report.json": reportBytes}); err != nil {
		return r, err
	}
	return r, nil
}

func signingOutside(privatePath, publicPath string) error {
	resolve := func(p string) (string, error) {
		absolute, err := filepath.Abs(p)
		if err != nil {
			return "", err
		}
		// Resolve existing ancestors as well as the final component, including
		// the creation case where the private file does not exist yet.
		parts := []string{}
		current := absolute
		for {
			resolved, err := filepath.EvalSymlinks(current)
			if err == nil {
				return filepath.Join(append([]string{resolved}, parts...)...), nil
			}
			if !errors.Is(err, os.ErrNotExist) {
				return "", err
			}
			parent := filepath.Dir(current)
			if parent == current {
				return "", err
			}
			parts = append([]string{filepath.Base(current)}, parts...)
			current = parent
		}
	}
	private, err := resolve(privatePath)
	if err != nil {
		return err
	}
	public, err := resolve(publicPath)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(public, private)
	if err != nil {
		return err
	}
	if relative == "." || relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("private signing context must be outside retained public roots")
	}
	return nil
}
