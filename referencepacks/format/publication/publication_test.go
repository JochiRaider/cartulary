package publication

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/dataset"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"github.com/JochiRaider/cartulary/referencepacks/format/verification"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type fixedClock struct{ time.Time }

func (c fixedClock) Now() time.Time { return c.Time }
func testClock() Clock {
	at, _ := time.Parse(time.RFC3339, "2030-01-01T00:00:00Z")
	return fixedClock{at}
}
func encode(t *testing.T, v any) []byte {
	t.Helper()
	b, err := canonicaljson.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func logicalFixture(t *testing.T, scope Scope) (EnvelopeInput, ApprovedOutput) {
	t.Helper()
	directory := t.TempDir()
	logical := filepath.Join(directory, "logical")
	public := filepath.Join(directory, "public")
	ledger := filepath.Join(directory, "ledger")
	d := dataset.Descriptor{SchemaID: dataset.DescriptorSchema, Lookups: []dataset.Lookup{{Name: "alias", InputType: "scalar", Components: []dataset.Component{{Name: "value", Type: "string", Normalization: "ascii_case_insensitive"}}, Coverage: dataset.Coverage{Objects: 1, Postings: 1, DistinctKeys: 1}, Unique: true}}}
	o := dataset.Object{SchemaID: dataset.ObjectSchema, ID: "sample-1", Type: "sample", DisplayLabel: "Sample", SourceRefs: []dataset.SourceRef{{ArtifactIndex: 0, Locator: "#"}}, Data: json.RawMessage(`{"factual":"source"}`), LookupValues: map[string][]json.RawMessage{"alias": {json.RawMessage(`"EXAMPLE"`)}}}
	members := map[string][]byte{dataset.DescriptorPath: encode(t, d), dataset.ObjectsPath: append(encode(t, o), '\n'), dataset.RelationshipsPath: {}, "notices/license.txt": []byte("Test-only material. CC0-1.0.\n")}
	sourceAsOf := "2029-12-31T00:00:00Z"
	one, zero := int64(1), int64(0)
	decision := Digest([]byte("test-only approved output"))
	m := envelope.Manifest{SchemaID: "cartulary.reference_pack_manifest.v2", Contract: envelope.ContractID, Key: scope.key, Kind: "enrichment", Version: "2030.1", Sequence: 1, ProfileID: dataset.ProfileID, ProfileVersion: "1", SourceProfileID: "example.source.v1", SourceProfileSHA256: Digest([]byte("profile")), Repository: &scope.repository, SourceIdentifier: "Example source", SourceVersion: "1", SourceAsOf: &sourceAsOf, BuiltAt: "2030-01-01T00:00:00Z", Artifacts: []envelope.SourceArtifact{{Ref: "urn:cartulary:reference-pack:license-decision", Version: "1", SHA256: decision}}, Dependencies: []envelope.Dependency{}, Conflicts: []envelope.Conflict{}, Files: LogicalMemberInventory(members), License: envelope.License{Expression: "CC0-1.0", ListVersion: "3.28.0", Redistribution: "allowed", Notices: []string{"notices/license.txt"}, Bindings: []envelope.LicenseBinding{}}, Extensions: map[string]any{}}
	m.Builder.ID = "example.producer"
	m.Builder.Version = "1"
	m.Builder.SHA256 = Digest([]byte("builder"))
	m.Compatibility.Majors = []int{1}
	m.Compatibility.Capabilities = []string{}
	m.Summary.Kind = "objects_relationships"
	m.Summary.Objects = &one
	m.Summary.Relationships = &zero
	raw := encode(t, m)
	members["manifest.json"] = raw
	for name, b := range members {
		file := filepath.Join(logical, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	input := EnvelopeInput{LogicalRoot: logical, OutputRoot: public, SigningContext: filepath.Join(directory, "private", "signing.json"), CreateSigningContext: true, IssuedAt: "2030-01-01T00:00:00Z", RootExpires: "2031-01-01T00:00:00Z", OnlineExpires: "2030-01-31T00:00:00Z", MetadataVersion: 1, LedgerRoot: ledger}
	return input, ApprovedOutput{ManifestSHA256: Digest(raw), DecisionSHA256: decision, DecisionSourceRef: "urn:cartulary:reference-pack:license-decision"}
}

func TestPublicationAuthorityAndRenewal(t *testing.T) {
	ctx := context.Background()
	release, _ := ReleaseScope("example.release", "enrichment.example")
	input, approval := logicalFixture(t, release)
	report, err := PublishRelease(ctx, ReleaseInput{input, approval}, release, testClock())
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(input.OutputRoot, report.EnvelopePath)
	archive, err := ReadRegular(filepath.Join(output, "container.pack"), 536870912)
	if err != nil {
		t.Fatal(err)
	}
	bootstrap, err := ReadRegular(filepath.Join(output, "bootstrap.json"), 2097152)
	if err != nil {
		t.Fatal(err)
	}
	roots, err := envelope.DecodeBootstrap(bootstrap)
	if err != nil {
		t.Fatal(err)
	}
	workspace := memoryWorkspace{}
	result, err := verification.Dataset(ctx, bytes.NewReader(archive), int64(len(archive)), roots, testClock().Now(), workspace)
	if err != nil || result.Dataset.Objects != 1 || result.Dataset.Postings != 1 {
		t.Fatal("standalone import", err)
	}
	if result.ManifestSHA256 != approval.ManifestSHA256 || Digest(archive) != report.ContainerSHA256 {
		t.Fatal("exact artifact identity")
	}
	if key, err := dataset.Normalize(result.Dataset.Descriptor, "alias", []byte(`"example"`)); err != nil || key != `["example"]` {
		t.Fatal("normalized declared lookup", key, err)
	}
	replay, err := PublishRelease(ctx, ReleaseInput{input, approval}, release, testClock())
	if err != nil || replay.ContainerSHA256 != report.ContainerSHA256 {
		t.Fatal("identical replay", err)
	}
	input.CreateSigningContext = false
	input.PreviousEnvelope = output
	input.MetadataVersion = 2
	input.IssuedAt = "2030-01-02T00:00:00Z"
	input.OnlineExpires = "2030-02-01T00:00:00Z"
	renewed, err := PublishRelease(ctx, ReleaseInput{input, approval}, release, testClock())
	if err != nil {
		t.Fatal(err)
	}
	if renewed.ManifestSHA256 != report.ManifestSHA256 || renewed.PayloadSHA256 != report.PayloadSHA256 || renewed.BootstrapSHA256 != report.BootstrapSHA256 || renewed.ContainerSHA256 == report.ContainerSHA256 || renewed.PreviousReportSHA256 == nil {
		t.Fatal("renewal changed logical identity or failed to advance")
	}
	if err = os.WriteFile(filepath.Join(output, "container.pack"), []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	input.MetadataVersion = 3
	if _, err = PublishRelease(ctx, ReleaseInput{input, approval}, release, testClock()); err == nil {
		t.Fatal("tampered predecessor accepted")
	}
}

func TestPublicationRejectsAuthorityAndOutputMismatch(t *testing.T) {
	ctx := context.Background()
	release, _ := ReleaseScope("example.release", "enrichment.example")
	fixture, _ := FixtureScope("fixture.example", "enrichment.example")
	for _, name := range []string{"fixture entrypoint", "fixture signing helper", "fixture authority", "unapproved output", "unbound decision", "private output", "expired", "changed logical member", "wrong key"} {
		t.Run(name, func(t *testing.T) {
			in, approval := logicalFixture(t, release)
			scope := release
			clock := testClock()
			switch name {
			case "fixture signing helper":
				signer, err := LoadFileSigner(in.SigningContext, true, release, in.OutputRoot)
				if err != nil {
					t.Fatal(err)
				}
				if _, _, err = BuildFixtureEnvelope(ctx, nil, signer, in.RootExpires, in.OnlineExpires, 1); err == nil {
					t.Fatal("release signer entered unvalidated fixture helper")
				}
				return
			case "fixture entrypoint":
				if _, err := PublishFixture(ctx, FixtureInput{in}, release, clock); err == nil {
					t.Fatal("release through fixture entrypoint")
				}
				return
			case "fixture authority":
				scope = fixture
			case "unapproved output":
				approval.ManifestSHA256 = Digest(nil)
			case "unbound decision":
				approval.DecisionSHA256 = Digest(nil)
			case "private output":
				in.SigningContext = filepath.Join(in.OutputRoot, "key.json")
			case "expired":
				at, _ := time.Parse(time.RFC3339, in.OnlineExpires)
				clock = fixedClock{at}
			case "changed logical member":
				if err := os.WriteFile(filepath.Join(in.LogicalRoot, dataset.ObjectsPath), []byte("{}\n"), 0600); err != nil {
					t.Fatal(err)
				}
			case "wrong key":
				scope, _ = ReleaseScope("example.release", "enrichment.other")
			}
			if _, err := PublishRelease(ctx, ReleaseInput{in, approval}, scope, clock); err == nil {
				t.Fatal("invalid release accepted")
			}
			if _, err := os.Stat(filepath.Join(in.OutputRoot, "envelopes")); !os.IsNotExist(err) {
				t.Fatal("failed release was exposed", err)
			}
		})
	}
	in, _ := logicalFixture(t, fixture)
	if _, err := PublishFixture(ctx, FixtureInput{in}, fixture, testClock()); err != nil {
		t.Fatal("fixture path", err)
	}
	if _, err := LoadFileSigner(in.SigningContext, false, release, in.OutputRoot); err == nil {
		t.Fatal("fixture keys admitted as release authority")
	}
	if _, err := ReleaseScope("fixture.example", "enrichment.example"); err == nil {
		t.Fatal("fixture namespace accepted for release")
	}
	if _, err := FixtureScope("example.release", "enrichment.example"); err == nil {
		t.Fatal("release namespace accepted for fixture")
	}
}

func TestPublicationReservationsAndImmutableTrees(t *testing.T) {
	directory := t.TempDir()
	scope, _ := ReleaseScope("example.release", "enrichment.example")
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			p, err := ReserveRelease(directory, "1", Digest([]byte("inputs")), scope)
			if err != nil || p.Sequence != 1 {
				t.Errorf("concurrent reservation: %v %v", p, err)
			}
		}()
	}
	group.Wait()
	if _, err := ReserveRelease(directory, "1", Digest([]byte("changed")), scope); err == nil {
		t.Fatal("sequence collision accepted")
	}
	p, err := ReserveRelease(directory, "2", Digest([]byte("next")), scope)
	if err != nil || p.Sequence != 2 {
		t.Fatal("continuity", p, err)
	}
	if err = os.WriteFile(filepath.Join(directory, "logical", "0000000000000002.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = ReserveRelease(directory, "3", Digest(nil), scope); err == nil {
		t.Fatal("tampered ledger accepted")
	}
	tree := t.TempDir()
	members := map[string][]byte{"a/file": []byte("one"), "b": []byte("two")}
	if err = PublishTree(tree, "release", members); err != nil {
		t.Fatal(err)
	}
	if err = PublishTree(tree, "release", members); err != nil {
		t.Fatal("identical replay", err)
	}
	members["b"] = []byte("changed")
	if err = PublishTree(tree, "release", members); err == nil {
		t.Fatal("mutated immutable tree accepted")
	}
	if err = PublishTree(tree, "../escape", members); err == nil {
		t.Fatal("escaping output accepted")
	}
	if err = os.Symlink(t.TempDir(), filepath.Join(tree, "link")); err != nil {
		t.Fatal(err)
	}
	if err = PublishTree(tree, "link/release", members); err == nil {
		t.Fatal("symlink output accepted")
	}
	if err = unix.Mkfifo(filepath.Join(tree, "fifo"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = ReadRegular(filepath.Join(tree, "fifo"), 1024); err == nil {
		t.Fatal("nonregular input admitted")
	}
	if err = os.WriteFile(filepath.Join(tree, "oversized"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(tree)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err = Retain(root, "oversized", []byte("one"), 0600); err == nil {
		t.Fatal("oversized existing evidence admitted")
	}
}

func TestPublicationMetadataReservations(t *testing.T) {
	scope, _ := ReleaseScope("example.release", "enrichment.example")
	in, approval := logicalFixture(t, scope)
	ctx := context.Background()
	first, err := PublishRelease(ctx, ReleaseInput{in, approval}, scope, testClock())
	if err != nil {
		t.Fatal(err)
	}
	in.OnlineExpires = "2030-01-30T00:00:00Z"
	if _, err = PublishRelease(ctx, ReleaseInput{in, approval}, scope, testClock()); err == nil {
		t.Fatal("same version different envelope accepted")
	}
	in.PreviousEnvelope = filepath.Join(in.OutputRoot, first.EnvelopePath)
	in.MetadataVersion = 3
	if _, err = PublishRelease(ctx, ReleaseInput{in, approval}, scope, testClock()); err != nil {
		t.Fatal(err)
	}
	in.MetadataVersion = 2
	if _, err = PublishRelease(ctx, ReleaseInput{in, approval}, scope, testClock()); err == nil {
		t.Fatal("metadata rollback reservation accepted")
	}
	// An identical earlier reservation cannot bypass verification of later
	// retained ledger records. Replay must not conceal post-publication damage.
	if err = os.WriteFile(filepath.Join(in.LedgerRoot, "envelopes", first.ManifestSHA256, "0000000000000003.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = reserveEnvelope(in.LedgerRoot, first, encode(t, first)); err == nil {
		t.Fatal("replay concealed damaged later reservation")
	}
}

func TestPublicationSharedLedgerLifecycle(t *testing.T) {
	ctx := context.Background()
	scope, _ := ReleaseScope("example.release", "enrichment.example")
	in, approval := logicalFixture(t, scope)
	raw, err := ReadRegular(filepath.Join(in.LogicalRoot, "manifest.json"), 1048576)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := envelope.DecodeManifest(raw, true)
	if err != nil {
		t.Fatal(err)
	}
	members := map[string][]byte{}
	for _, file := range manifest.Files {
		members[file.Path], err = ReadRegular(filepath.Join(in.LogicalRoot, file.Path), 4194304)
		if err != nil {
			t.Fatal(err)
		}
	}
	logical := LogicalInput{Scope: scope, Manifest: manifest, Members: members, LedgerRoot: in.LedgerRoot, OutputRoot: in.OutputRoot}
	first, err := PublishLogical(ctx, logical)
	if err != nil || first.Manifest.Sequence != 1 {
		t.Fatal("first logical release", err)
	}
	in.LogicalRoot = filepath.Join(in.OutputRoot, first.LogicalPath)
	approval.ManifestSHA256 = first.ManifestSHA256
	signed, err := PublishRelease(ctx, ReleaseInput{in, approval}, scope, testClock())
	if err != nil {
		t.Fatal(err)
	}
	if replay, err := PublishLogical(ctx, logical); err != nil || replay.ManifestSHA256 != first.ManifestSHA256 {
		t.Fatal("logical replay after signing", err)
	}
	in.CreateSigningContext = false
	in.PreviousEnvelope = filepath.Join(in.OutputRoot, signed.EnvelopePath)
	in.MetadataVersion = 2
	in.IssuedAt, in.OnlineExpires = "2030-01-02T00:00:00Z", "2030-02-01T00:00:00Z"
	if renewed, err := PublishRelease(ctx, ReleaseInput{in, approval}, scope, testClock()); err != nil || renewed.ManifestSHA256 != first.ManifestSHA256 {
		t.Fatal("renewal after assembly", err)
	}
	logical.Manifest.Version = "2030.2"
	next, err := PublishLogical(ctx, logical)
	if err != nil || next.Manifest.Sequence != 2 {
		t.Fatal("next logical release after renewal", err)
	}
	in.LogicalRoot = filepath.Join(in.OutputRoot, next.LogicalPath)
	in.PreviousEnvelope = ""
	approval.ManifestSHA256 = next.ManifestSHA256
	if _, err := PublishRelease(ctx, ReleaseInput{in, approval}, scope, testClock()); err != nil {
		t.Fatal("next signed release", err)
	}
	old := t.TempDir()
	if err := os.WriteFile(filepath.Join(old, "0000000000000001.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReserveRelease(old, "1", Digest(nil), scope); err == nil {
		t.Fatal("flat predecessor ledger silently reset")
	}
	if err := reserveEnvelope(old, signed, encode(t, signed)); err == nil {
		t.Fatal("flat predecessor envelope ledger silently reset")
	}
}
