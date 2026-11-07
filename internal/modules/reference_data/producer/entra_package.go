package producer

import (
	"context"
	"errors"
	"fmt"
	"os"

	entrainputs "github.com/JochiRaider/cartulary/contracts/reference-pack-producers/entra-app-ids"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type EntraBuildInput struct {
	SchemaID        string `json:"schema_id"`
	SourceRoot      string `json:"source_root"`
	SourceInventory string `json:"source_inventory"`
	OutputRoot      string `json:"output_root"`
	LedgerRoot      string `json:"ledger_root"`
	BuiltAt         string `json:"built_at"`
	PackVersion     string `json:"pack_version"`
	BuilderSHA256   string `json:"builder_sha256"`
	Repository      string `json:"trust_repository_id"`
}
type EntraPackageReport struct {
	SchemaID              string  `json:"schema_id"`
	ManifestSHA256        string  `json:"manifest_sha256"`
	PayloadSHA256         string  `json:"payload_sha256"`
	EntriesSHA256         string  `json:"entries_sha256"`
	SourceProfileSHA256   string  `json:"source_profile_sha256"`
	SourceLockSHA256      string  `json:"source_lock_sha256"`
	ReconciliationSHA256  string  `json:"reconciliation_sha256"`
	LicenseDecisionSHA256 string  `json:"license_decision_sha256"`
	Builder               Builder `json:"builder"`
	Entries               int     `json:"entry_count"`
	LogicalPath           string  `json:"logical_path"`
}

func AssembleEntra(ctx context.Context, input []byte, builder Builder) (EntraPackageReport, error) {
	var in EntraBuildInput
	if err := DecodeClosed(input, &in); err != nil {
		return EntraPackageReport{}, err
	}
	if in.SchemaID != "cartulary.entra_build_input.v2" || in.Repository != EntraRepository || in.BuilderSHA256 != builder.SHA256 || !validBuildTime(in.BuiltAt) {
		return EntraPackageReport{}, errors.New("entra assembly requires explicit current publisher, builder and build time")
	}
	for _, pair := range [][2]string{{in.SourceRoot, in.OutputRoot}, {in.OutputRoot, in.SourceRoot}, {in.LedgerRoot, in.OutputRoot}, {in.OutputRoot, in.LedgerRoot}} {
		if err := signingOutside(pair[0], pair[1]); err != nil {
			return EntraPackageReport{}, errors.New("entra research, ledger and distribution roots must be disjoint")
		}
	}
	inventory, err := readRegular(in.SourceInventory)
	if err != nil {
		return EntraPackageReport{}, err
	}
	root, err := os.OpenRoot(in.SourceRoot)
	if err != nil {
		return EntraPackageReport{}, err
	}
	defer root.Close()
	projection, err := EntraReconcile(ctx, root, inventory, builder)
	if err != nil {
		return EntraPackageReport{}, err
	}
	lock, err := EntraCheck(root, inventory)
	if err != nil {
		return EntraPackageReport{}, err
	}
	if in.BuiltAt < lock.ObservedAt {
		return EntraPackageReport{}, errors.New("entra build precedes source acceptance")
	}
	policyBytes, err := entrainputs.Files.ReadFile("license.project.v2.json")
	if err != nil {
		return EntraPackageReport{}, err
	}
	var policy struct {
		SchemaID    string             `json:"schema_id"`
		DecisionRef string             `json:"decision_ref"`
		ApprovedBy  string             `json:"approved_by"`
		Scope       string             `json:"scope"`
		License     packformat.License `json:"license"`
	}
	if err = DecodeClosed(policyBytes, &policy); err != nil {
		return EntraPackageReport{}, err
	}
	decision, err := canonicaljson.Marshal(map[string]any{"schema_id": "cartulary.entra_license_decision.v2", "source_lock_sha256": projection.Report.SourceLockSHA256, "output_policy_sha256": Digest(policyBytes), "decision_ref": policy.DecisionRef, "approved_by": policy.ApprovedBy, "scope": policy.Scope, "license": policy.License})
	if err != nil {
		return EntraPackageReport{}, err
	}
	members := packageSource{"payload/entries.ndjson": projection.Payload}
	for _, notice := range []struct{ name, path string }{{"Apache-2.0.txt", "notices/Apache-2.0.txt"}, {"attribution.txt", "notices/attribution.txt"}} {
		b, e := entrainputs.Files.ReadFile(notice.name)
		if e != nil {
			return EntraPackageReport{}, e
		}
		if e = packformat.ValidateNotice(b); e != nil {
			return EntraPackageReport{}, e
		}
		members[notice.path] = b
	}
	// All additional decision references sort after HTTPS source artifacts, so
	// exact source-row indices assigned by reconciliation remain unchanged.
	artifacts := append(projection.Artifacts, packformat.SourceArtifact{Ref: "urn:cartulary:entra-input:license-decision.json", Version: EntraSourceProfile, SHA256: Digest(decision)})
	sortEntraArtifacts(artifacts)
	m := packformat.Manifest{SchemaID: "cartulary.reference_pack_manifest.v1", Contract: packformat.ContractID, Key: EntraKey, Kind: "enrichment", Version: in.PackVersion, Sequence: 1, ProfileID: EntraContentProfile, ProfileVersion: "2", SourceProfileID: EntraSourceProfile, SourceProfileSHA256: projection.Report.SourceProfileSHA256, Repository: &in.Repository, SourceIdentifier: "Cartulary factual Entra application IDs and original summaries; current and historical Microsoft documentation", SourceVersion: "snapshot-2026-11-06." + projection.Report.SourceLockSHA256, SourceAsOf: &lock.ObservedAt, BuiltAt: in.BuiltAt, Artifacts: artifacts, Dependencies: []packformat.Dependency{}, Conflicts: []packformat.Conflict{}, License: policy.License, Extensions: map[string]any{}}
	m.Builder.ID = builder.ID
	m.Builder.Version = builder.Version
	m.Builder.SHA256 = builder.SHA256
	m.Compatibility.Majors = []int{1}
	m.Compatibility.Capabilities = []string{}
	count := int64(projection.Report.Counts["entries"])
	m.Summary.Kind = "entries"
	m.Summary.Entries = &count
	m.Files = logicalMemberInventory(members)
	raw, err := canonicaljson.Marshal(m)
	if err != nil {
		return EntraPackageReport{}, err
	}
	if _, err = packformat.DecodeManifest(raw, true); err != nil {
		return EntraPackageReport{}, fmt.Errorf("entra manifest: %w", err)
	}
	if err = packformat.ValidateContent(ctx, m, members, nil); err != nil {
		return EntraPackageReport{}, fmt.Errorf("entra content: %w", err)
	}
	reconciliation, err := canonicaljson.Marshal(projection.Report)
	if err != nil {
		return EntraPackageReport{}, err
	}
	// Sequence allocation is after validation. The provisional manifest binds
	// every output input while deliberately excluding destination filesystem paths.
	publication, err := allocateScopedPublication(in.LedgerRoot, in.PackVersion, Digest(raw), "cartulary.entra_publication.v2", in.Repository, EntraKey)
	if err != nil {
		return EntraPackageReport{}, err
	}
	m.Sequence = publication.Sequence
	raw, err = canonicaljson.Marshal(m)
	if err != nil {
		return EntraPackageReport{}, err
	}
	payload, err := packformat.PayloadDigest(m.Files)
	if err != nil {
		return EntraPackageReport{}, err
	}
	report := EntraPackageReport{"cartulary.entra_package_report.v2", Digest(raw), payload, Digest(projection.Payload), projection.Report.SourceProfileSHA256, projection.Report.SourceLockSHA256, Digest(reconciliation), Digest(decision), builder, int(count), "logical/" + Digest(raw)}
	reportBytes, err := canonicaljson.Marshal(report)
	if err != nil {
		return report, err
	}
	projection.Research["license-decision.json"] = decision
	projection.Research["reconciliation.json"] = reconciliation
	projection.Research["build-report.json"] = reportBytes
	if err = publishTree(in.SourceRoot, "publications/"+report.ManifestSHA256, projection.Research); err != nil {
		return report, err
	}
	members["manifest.json"] = raw
	if err = publishTree(in.OutputRoot, report.LogicalPath, members); err != nil {
		return report, err
	}
	return report, nil
}
