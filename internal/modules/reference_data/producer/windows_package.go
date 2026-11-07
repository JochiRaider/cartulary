package producer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	windowsinputs "github.com/JochiRaider/cartulary/contracts/reference-pack-producers/windows-event-ids"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type WindowsBuildInput struct {
	SchemaID        string `json:"schema_id"`
	SourceRoot      string `json:"source_root"`
	SourceInventory string `json:"source_inventory"`
	Editorial       string `json:"editorial"`
	OutputRoot      string `json:"output_root"`
	LedgerRoot      string `json:"ledger_root"`
	BuiltAt         string `json:"built_at"`
	PackVersion     string `json:"pack_version"`
	BuilderSHA256   string `json:"builder_sha256"`
	Repository      string `json:"trust_repository_id"`
	LicenseDecision string `json:"license_decision"`
}
type WindowsNotice struct {
	Path     string  `json:"path"`
	SourceID string  `json:"source_id,omitempty"`
	Content  *string `json:"content,omitempty"`
}
type WindowsLicenseDecision struct {
	SchemaID         string             `json:"schema_id"`
	Purpose          string             `json:"purpose"`
	SourceLockSHA256 string             `json:"source_lock_sha256"`
	DecisionRef      string             `json:"decision_ref"`
	ApprovedBy       string             `json:"approved_by"`
	License          packformat.License `json:"license"`
	Notices          []WindowsNotice    `json:"notices"`
}
type WindowsPackageReport struct {
	SchemaID            string  `json:"schema_id"`
	ManifestSHA256      string  `json:"manifest_sha256"`
	PayloadSHA256       string  `json:"payload_sha256"`
	PayloadFileSHA256   string  `json:"payload_file_sha256"`
	SourceProfileSHA256 string  `json:"source_profile_sha256"`
	SourceLockSHA256    string  `json:"source_lock_sha256"`
	SourceAuditSHA256   string  `json:"source_audit_sha256"`
	EvidenceSHA256      string  `json:"evidence_sha256"`
	Builder             Builder `json:"builder"`
	Entries             int     `json:"entry_count"`
	LogicalPath         string  `json:"logical_path"`
}

func AssembleWindows(ctx context.Context, inputBytes []byte, builder Builder) (WindowsPackageReport, error) {
	var in WindowsBuildInput
	if err := windowsDecode(inputBytes, "build", &in); err != nil {
		return WindowsPackageReport{}, err
	}
	for _, paths := range [][2]string{{in.SourceRoot, in.OutputRoot}, {in.OutputRoot, in.SourceRoot}} {
		if err := signingOutside(paths[0], paths[1]); err != nil {
			return WindowsPackageReport{}, errors.New("research retention and distribution roots must be disjoint")
		}
	}
	if in.BuilderSHA256 != builder.SHA256 || !validBuildTime(in.BuiltAt) {
		return WindowsPackageReport{}, errors.New("exact builder and explicit UTC build time required")
	}
	inventory, err := readRegular(in.SourceInventory)
	if err != nil {
		return WindowsPackageReport{}, err
	}
	editorial, err := readRegular(in.Editorial)
	if err != nil {
		return WindowsPackageReport{}, err
	}
	root, err := os.OpenRoot(in.SourceRoot)
	if err != nil {
		return WindowsPackageReport{}, err
	}
	defer root.Close()
	projection, err := WindowsReconcile(ctx, root, inventory, editorial, builder)
	if err != nil {
		return WindowsPackageReport{}, err
	}
	lock, err := WindowsCheck(root, inventory)
	if err != nil {
		return WindowsPackageReport{}, err
	}
	// Licensing applies to the projected output, never to undistributed research bytes.
	decisionBytes, err := readRegular(in.LicenseDecision)
	if err != nil {
		return WindowsPackageReport{}, windowsFailure("missing_publication_inputs", "license_decision", "recorded output licensing decision required")
	}
	var decision WindowsLicenseDecision
	if err = windowsDecode(decisionBytes, "license", &decision); err != nil {
		return WindowsPackageReport{}, err
	}
	if decision.Purpose != lock.Purpose || decision.SourceLockSHA256 != projection.Report.SourceLockSHA256 {
		return WindowsPackageReport{}, windowsFailure("missing_publication_inputs", "license_decision", "decision must cover exact corpus and purpose")
	}
	if lock.Purpose == "real" {
		approved, e := windowsinputs.Files.ReadFile("license.project.v2.json")
		if e != nil {
			return WindowsPackageReport{}, e
		}
		var policy WindowsLicenseDecision
		if e = windowsDecode(approved, "license", &policy); e != nil {
			return WindowsPackageReport{}, e
		}
		actual, e := canonicaljson.Marshal(decision)
		if e != nil {
			return WindowsPackageReport{}, e
		}
		expected, e := canonicaljson.Marshal(policy)
		if e != nil {
			return WindowsPackageReport{}, e
		}
		if !bytes.Equal(actual, expected) {
			return WindowsPackageReport{}, errors.New("real output requires the recorded project-license decision; research notices cannot be distributed")
		}
	}
	if lock.Purpose == "real" && in.Repository != WindowsReleaseRepository {
		return WindowsPackageReport{}, errors.New("real Windows pack requires its independent release repository")
	}
	if lock.Purpose == "synthetic" && in.Repository != WindowsFixtureRepository {
		return WindowsPackageReport{}, errors.New("synthetic inputs require Windows fixture repository")
	}
	sourceAsOf := ""
	sources := map[string]WindowsAcquired{}
	for _, a := range lock.Sources {
		sourceAsOf = max(sourceAsOf, a.ReceivedAt)
		sources[a.Source.ID] = a
	}
	if in.BuiltAt < sourceAsOf {
		return WindowsPackageReport{}, errors.New("build time precedes source acquisition")
	}
	members := packageSource{"payload/entries.ndjson": projection.Payload}
	for _, notice := range decision.Notices {
		if _, duplicate := members[notice.Path]; duplicate || !strings.HasPrefix(notice.Path, "notices/") || !strings.HasSuffix(notice.Path, ".txt") {
			return WindowsPackageReport{}, errors.New("invalid notice inventory")
		}
		var b []byte
		var e error
		if notice.Content != nil {
			b = []byte(*notice.Content)
		} else {
			a, ok := sources[notice.SourceID]
			if !ok || a.Source.Kind != "license" || lock.Purpose != "synthetic" {
				return WindowsPackageReport{}, errors.New("only synthetic fixtures may distribute an acquired notice")
			}
			b, e = ReadEvidence(root, a.Evidence)
			if e != nil {
				return WindowsPackageReport{}, e
			}
		}
		if e = packformat.ValidateNotice(b); e != nil {
			return WindowsPackageReport{}, e
		}
		members[notice.Path] = b
	}
	manifest := packformat.Manifest{SchemaID: "cartulary.reference_pack_manifest.v1", Contract: packformat.ContractID, Key: WindowsKey, Kind: "enrichment", Version: in.PackVersion, Sequence: 1, ProfileID: "cartulary.reference_pack.enrichment.windows_event_ids.v2", ProfileVersion: "2", SourceProfileID: WindowsProfile, SourceProfileSHA256: projection.Report.SourceProfileSHA256, Repository: &in.Repository, SourceIdentifier: "Cartulary factual Windows event definitions and original summaries", SourceVersion: "microsoft-references:" + projection.Report.SourceLockSHA256, SourceAsOf: &sourceAsOf, BuiltAt: in.BuiltAt, Artifacts: []packformat.SourceArtifact{{Ref: "urn:cartulary:windows-event-evidence:" + projection.Report.EvidenceSHA256, Version: "microsoft-references:" + projection.Report.SourceLockSHA256, SHA256: projection.Report.EvidenceSHA256}}, Dependencies: []packformat.Dependency{}, Conflicts: []packformat.Conflict{}, License: decision.License, Extensions: map[string]any{}}
	manifest.Builder.ID = builder.ID
	if lock.Purpose == "synthetic" {
		manifest.SourceIdentifier = "Cartulary synthetic Windows producer fixtures; not Microsoft source evidence"
	}
	manifest.Builder.Version = builder.Version
	manifest.Builder.SHA256 = builder.SHA256
	manifest.Compatibility.Majors = []int{1}
	manifest.Compatibility.Capabilities = []string{}
	count := int64(projection.Report.EntryCount)
	manifest.Summary.Kind = "entries"
	manifest.Summary.Entries = &count
	manifest.Files = logicalMemberInventory(members)
	raw, err := canonicaljson.Marshal(manifest)
	if err != nil {
		return WindowsPackageReport{}, err
	}
	if _, err = packformat.DecodeManifest(raw, true); err != nil {
		return WindowsPackageReport{}, fmt.Errorf("windows manifest: %w", err)
	}
	if err = packformat.ValidateContent(ctx, manifest, members, nil); err != nil {
		return WindowsPackageReport{}, fmt.Errorf("windows content: %w", err)
	}
	audit, err := canonicaljson.Marshal(projection.Report)
	if err != nil {
		return WindowsPackageReport{}, err
	}
	// Bind every output-affecting input, not local filesystem paths.
	publicationInput, err := canonicaljson.Marshal(map[string]any{"source_lock_sha256": projection.Report.SourceLockSHA256, "source_profile_sha256": projection.Report.SourceProfileSHA256, "editorial_sha256": Digest(editorial), "license_decision_sha256": Digest(decisionBytes), "builder_sha256": builder.SHA256, "built_at": in.BuiltAt, "pack_version": in.PackVersion, "trust_repository_id": in.Repository, "payload_file_sha256": projection.Report.PayloadFileSHA256, "evidence_sha256": projection.Report.EvidenceSHA256})
	if err != nil {
		return WindowsPackageReport{}, err
	}
	publication, err := allocateScopedPublication(in.LedgerRoot, in.PackVersion, Digest(publicationInput), "cartulary.windows_event_publication.v2", in.Repository, WindowsKey)
	if err != nil {
		return WindowsPackageReport{}, err
	}
	manifest.Sequence = publication.Sequence
	raw, err = canonicaljson.Marshal(manifest)
	if err != nil {
		return WindowsPackageReport{}, err
	}
	payload, err := packformat.PayloadDigest(manifest.Files)
	if err != nil {
		return WindowsPackageReport{}, err
	}
	report := WindowsPackageReport{"cartulary.windows_event_package_report.v2", Digest(raw), payload, projection.Report.PayloadFileSHA256, projection.Report.SourceProfileSHA256, projection.Report.SourceLockSHA256, Digest(audit), projection.Report.EvidenceSHA256, builder, projection.Report.EntryCount, "logical/" + Digest(raw)}
	members["manifest.json"] = raw
	reportBytes, err := canonicaljson.Marshal(report)
	if err != nil {
		return report, err
	}
	research := map[string][]byte{"evidence.tar": projection.Archive, "source-profile.json": projection.SourceProfile, "license-decision.json": decisionBytes, "reconciliation.json": audit, "publication-input.json": publicationInput, "build-report.json": reportBytes}
	if err = publishTree(in.SourceRoot, "publications/"+report.ManifestSHA256, research); err != nil {
		return report, err
	}

	if err = publishTree(in.OutputRoot, report.LogicalPath, members); err != nil {
		return WindowsPackageReport{}, err
	}
	return report, nil
}
