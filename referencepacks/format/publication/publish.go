package publication

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"github.com/JochiRaider/cartulary/referencepacks/format/verification"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type EnvelopeInput struct {
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

// ApprovedOutput binds a producer-reviewed licensing decision to these exact
// manifest bytes. Its decision must also occur in the manifest provenance.
// Source interpretation and the authority to approve remain producer obligations.
type ApprovedOutput struct {
	ManifestSHA256    string `json:"manifest_sha256"`
	DecisionSHA256    string `json:"decision_sha256"`
	DecisionSourceRef string `json:"decision_source_ref"`
}
type ReleaseInput struct {
	Envelope EnvelopeInput
	Approval ApprovedOutput
}
type FixtureInput struct{ Envelope EnvelopeInput }
type EnvelopeReport struct {
	SchemaID             string          `json:"schema_id"`
	Class                AuthorityClass  `json:"authority_class"`
	Repository           string          `json:"trust_repository_id"`
	PackKey              string          `json:"pack_key"`
	ManifestSHA256       string          `json:"manifest_sha256"`
	PayloadSHA256        string          `json:"payload_sha256"`
	ContainerSHA256      string          `json:"container_sha256"`
	ContainerBytes       int64           `json:"container_bytes"`
	BootstrapSHA256      string          `json:"bootstrap_sha256"`
	MetadataVersion      int64           `json:"metadata_version"`
	IssuedAt             string          `json:"issued_at"`
	EnvelopePath         string          `json:"envelope_path"`
	PreviousReportSHA256 *string         `json:"previous_report_sha256"`
	Approval             *ApprovedOutput `json:"approved_output"`
}

func PublishRelease(ctx context.Context, input ReleaseInput, scope Scope, clock Clock) (EnvelopeReport, error) {
	if scope.class != ReleaseAuthority {
		return EnvelopeReport{}, errors.New("release authority required")
	}
	return publish(ctx, input.Envelope, &input.Approval, scope, clock)
}
func PublishFixture(ctx context.Context, input FixtureInput, scope Scope, clock Clock) (EnvelopeReport, error) {
	if scope.class != FixtureAuthority {
		return EnvelopeReport{}, errors.New("fixture authority required")
	}
	return publish(ctx, input.Envelope, nil, scope, clock)
}
func publish(ctx context.Context, in EnvelopeInput, approval *ApprovedOutput, scope Scope, clock Clock) (EnvelopeReport, error) {
	var report EnvelopeReport
	if !scope.valid() || clock == nil || in.LogicalRoot == "" || in.OutputRoot == "" || in.LedgerRoot == "" || in.SigningContext == "" || in.MetadataVersion < 1 || in.MetadataVersion > 9007199254740991 || !validBuildTime(in.IssuedAt) || !validBuildTime(in.RootExpires) || !validBuildTime(in.OnlineExpires) || in.OnlineExpires <= in.IssuedAt || in.RootExpires < in.OnlineExpires {
		return report, errors.New("invalid explicit publication inputs")
	}
	if err := ctx.Err(); err != nil {
		return report, err
	}
	for _, root := range []string{in.LogicalRoot, in.OutputRoot, in.LedgerRoot} {
		if err := signingOutside(in.SigningContext, root); err != nil {
			return report, err
		}
	}
	manifestBytes, err := ReadRegular(filepath.Join(in.LogicalRoot, "manifest.json"), 1048576)
	if err != nil {
		return report, err
	}
	manifest, err := envelope.DecodeManifest(manifestBytes, true)
	if err != nil {
		return report, err
	}
	if manifest.Repository == nil || *manifest.Repository != scope.repository || manifest.Key != scope.key {
		return report, errors.New("logical pack does not match publisher")
	}
	if approval != nil {
		if !digestPattern.MatchString(approval.DecisionSHA256) || approval.ManifestSHA256 != Digest(manifestBytes) {
			return report, errors.New("exact approved output required")
		}
		bound := false
		for _, artifact := range manifest.Artifacts {
			if artifact.Ref == approval.DecisionSourceRef && artifact.SHA256 == approval.DecisionSHA256 {
				bound = true
			}
		}
		if !bound {
			return report, errors.New("license decision provenance is missing")
		}
	}
	members := map[string][]byte{"manifest.json": manifestBytes}
	total := int64(len(manifestBytes))
	root, err := os.OpenRoot(in.LogicalRoot)
	if err != nil {
		return report, err
	}
	for _, file := range manifest.Files {
		if file.Size < 0 || file.Size > 536870912-total {
			root.Close()
			return report, errors.New("logical inventory size limit")
		}
		total += file.Size
		if err = noSymlinks(root, file.Path, false); err != nil {
			root.Close()
			return report, err
		}
		b, e := readRootRegular(root, file.Path, file.Size)
		if e != nil {
			root.Close()
			return report, e
		}
		members[file.Path] = b
	}
	if err = root.Close(); err != nil {
		return report, err
	}
	signer, err := LoadFileSigner(in.SigningContext, in.CreateSigningContext, scope, in.LogicalRoot, in.OutputRoot, in.LedgerRoot)
	if err != nil {
		return report, err
	}
	container, bootstrap, err := buildEnvelope(ctx, members, signer, scope, in.RootExpires, in.OnlineExpires, in.MetadataVersion)
	if err != nil {
		return report, err
	}
	payload, err := envelope.PayloadDigest(manifest.Files)
	if err != nil {
		return report, err
	}
	report = EnvelopeReport{SchemaID: "cartulary.reference_pack_envelope_report.v1", Class: scope.class, Repository: scope.repository, PackKey: scope.key, ManifestSHA256: Digest(manifestBytes), PayloadSHA256: payload, ContainerSHA256: Digest(container), ContainerBytes: int64(len(container)), BootstrapSHA256: Digest(bootstrap), MetadataVersion: in.MetadataVersion, IssuedAt: in.IssuedAt, EnvelopePath: "envelopes/" + Digest(container), Approval: approval}
	if in.PreviousEnvelope != "" {
		previousBytes, err := ReadRegular(filepath.Join(in.PreviousEnvelope, "envelope-report.json"), 65536)
		if err != nil {
			return report, err
		}
		var previous EnvelopeReport
		if err = DecodeClosed(previousBytes, &previous); err != nil {
			return report, err
		}
		if previous.SchemaID != report.SchemaID || previous.Class != report.Class || previous.Repository != report.Repository || previous.PackKey != report.PackKey || previous.ManifestSHA256 != report.ManifestSHA256 || previous.PayloadSHA256 != report.PayloadSHA256 || previous.BootstrapSHA256 != report.BootstrapSHA256 || previous.MetadataVersion >= report.MetadataVersion || previous.IssuedAt > report.IssuedAt {
			return report, errors.New("renewal must preserve logical bytes and root while advancing metadata")
		}
		prior, err := ReadRegular(filepath.Join(in.PreviousEnvelope, "container.pack"), 536870912)
		if err != nil || Digest(prior) != previous.ContainerSHA256 || int64(len(prior)) != previous.ContainerBytes {
			return report, errors.New("prior envelope integrity mismatch")
		}
		prior, err = ReadRegular(filepath.Join(in.PreviousEnvelope, "bootstrap.json"), 2097152)
		if err != nil || Digest(prior) != previous.BootstrapSHA256 {
			return report, errors.New("prior bootstrap integrity mismatch")
		}
		hash := Digest(previousBytes)
		report.PreviousReportSHA256 = &hash
	}
	encoded, err := canonicaljson.Marshal(report)
	if err != nil {
		return report, err
	}
	if err = reserveEnvelope(in.LedgerRoot, report, encoded); err != nil {
		return report, err
	}
	roots, err := envelope.DecodeBootstrap(bootstrap)
	if err != nil {
		return report, err
	}
	workspace := memoryWorkspace{}
	verified, err := verification.Dataset(ctx, bytes.NewReader(container), int64(len(container)), roots, clock.Now(), workspace)
	if err != nil {
		return report, err
	}
	if verified.ManifestSHA256 != report.ManifestSHA256 || verified.PayloadSHA256 != report.PayloadSHA256 {
		return report, errors.New("standalone verification identity mismatch")
	}
	if err = PublishTree(in.OutputRoot, report.EnvelopePath, map[string][]byte{"container.pack": container, "bootstrap.json": bootstrap, "envelope-report.json": encoded}); err != nil {
		return report, err
	}
	return report, nil
}

type memoryWorkspace map[string][]byte

func (w memoryWorkspace) Write(ctx context.Context, name string, write func(io.Writer) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, exists := w[name]; exists {
		return errors.New("duplicate workspace member")
	}
	var b bytes.Buffer
	if err := write(&b); err != nil {
		return err
	}
	w[name] = b.Bytes()
	return nil
}
func (w memoryWorkspace) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	b, ok := w[name]
	if !ok {
		return nil, errors.New("missing workspace member")
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func reserveEnvelope(directory string, report EnvelopeReport, encoded []byte) error {
	root, err := openLedgerNamespace(directory, "envelopes")
	if err != nil {
		return err
	}
	defer root.Close()
	if err = noSymlinks(root, report.ManifestSHA256, true); err != nil {
		return err
	}
	if err = root.MkdirAll(report.ManifestSHA256, 0700); err != nil {
		return err
	}
	ledger, err := root.OpenRoot(report.ManifestSHA256)
	if err != nil {
		return err
	}
	defer ledger.Close()
	lock, err := ledger.OpenFile(".lock", os.O_CREATE|os.O_RDWR|unix.O_NOFOLLOW, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err = unix.Flock(int(lock.Fd()), unix.LOCK_EX); err != nil {
		return err
	}
	defer unix.Flock(int(lock.Fd()), unix.LOCK_UN) //nolint:errcheck // close releases lock
	entries, err := readRootDirectory(ledger)
	if err != nil {
		return err
	}
	slices.SortFunc(entries, func(a, b os.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
	var highest int64
	replay := false
	for _, entry := range entries {
		if entry.Name() == ".lock" {
			continue
		}
		if entry.Type() != 0 {
			return errors.New("invalid envelope ledger member")
		}
		b, err := readRootRegular(ledger, entry.Name(), 65536)
		if err != nil {
			return err
		}
		var previous EnvelopeReport
		if err = DecodeClosed(b, &previous); err != nil {
			return err
		}
		canonical, err := canonicaljson.Marshal(previous)
		if err != nil || !bytes.Equal(canonical, b) || previous.SchemaID != report.SchemaID || previous.ManifestSHA256 != report.ManifestSHA256 || previous.Class != report.Class || previous.Repository != report.Repository || previous.PackKey != report.PackKey || previous.MetadataVersion <= highest || entry.Name() != fmt.Sprintf("%016d.json", previous.MetadataVersion) {
			return errors.New("invalid envelope reservation chain")
		}
		highest = previous.MetadataVersion
		if previous.MetadataVersion == report.MetadataVersion {
			if !bytes.Equal(b, encoded) {
				return errors.New("immutable envelope reservation collision")
			}
			replay = true
		}
	}
	if replay {
		return nil
	}
	if report.MetadataVersion <= highest {
		return errors.New("metadata reservation rollback")
	}
	return Retain(ledger, fmt.Sprintf("%016d.json", report.MetadataVersion), encoded, 0600)
}
