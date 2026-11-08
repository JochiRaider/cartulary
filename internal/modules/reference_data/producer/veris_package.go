package producer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	verisinputs "github.com/JochiRaider/cartulary/contracts/reference-pack-producers/veris"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type VERISBuildInput struct {
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

func VERISAuditSources(ctx context.Context, root *os.Root, inventory []byte, builder Builder) (VERISAudit, error) {
	p, err := VERISProject(ctx, root, inventory, builder)
	if err != nil {
		return VERISAudit{}, err
	}
	b, err := canonicaljson.Marshal(p.Report)
	if err != nil {
		return p.Report, err
	}
	for name, data := range p.Members {
		if err := Retain(root, "candidates/"+Digest(b)+"/"+name, data, 0600); err != nil {
			return p.Report, err
		}
	}
	return p.Report, Retain(root, "audits/"+Digest(b)+".json", b, 0600)
}

func AssembleVERIS(ctx context.Context, input []byte, builder Builder) (PackageReport, error) {
	var in VERISBuildInput
	var report PackageReport
	if err := DecodeClosed(input, &in); err != nil {
		return report, err
	}
	if in.SchemaID != "cartulary.veris_build_input.v1" || in.Repository != VERISRepository || in.BuilderSHA256 != builder.SHA256 || !validBuildTime(in.BuiltAt) {
		return report, errors.New("VERIS assembly requires current publisher, builder and build time")
	}
	for _, pair := range [][2]string{{in.SourceRoot, in.OutputRoot}, {in.OutputRoot, in.SourceRoot}, {in.LedgerRoot, in.OutputRoot}, {in.OutputRoot, in.LedgerRoot}, {in.SourceRoot, in.LedgerRoot}, {in.LedgerRoot, in.SourceRoot}} {
		if err := signingOutside(pair[0], pair[1]); err != nil {
			return report, errors.New("VERIS source, ledger and distribution roots must be disjoint")
		}
	}
	inventoryBytes, err := readRegular(in.SourceInventory)
	if err != nil {
		return report, err
	}
	inventory, err := verisInventory(inventoryBytes)
	if err != nil {
		return report, err
	}
	root, err := os.OpenRoot(in.SourceRoot)
	if err != nil {
		return report, err
	}
	defer root.Close()
	p, err := VERISProject(ctx, root, inventoryBytes, builder)
	if err != nil {
		return report, err
	}
	lock, err := VERISCheck(root, inventoryBytes)
	if err != nil {
		return report, err
	}
	if in.BuiltAt < lock.ObservedAt {
		return report, errors.New("VERIS build precedes acquisition")
	}
	policyBytes, err := verisinputs.Files.ReadFile("license.project.v1.json")
	if err != nil {
		return report, err
	}
	var policy struct {
		SchemaID      string             `json:"schema_id"`
		DecisionRef   string             `json:"decision_ref"`
		Authorization string             `json:"authorization"`
		Scope         string             `json:"scope"`
		License       packformat.License `json:"license"`
	}
	if err := DecodeClosed(policyBytes, &policy); err != nil {
		return report, err
	}
	if policy.SchemaID != "cartulary.veris_output_policy.v1" || policy.DecisionRef == "" || policy.Authorization == "" || policy.License.Expression != "CC-BY-SA-4.0" || policy.License.Redistribution != "allowed" {
		return report, errors.New("VERIS distribution disposition required")
	}
	for _, name := range []string{"CC-BY-SA-4.0.txt", "attribution.txt"} {
		b, err := verisinputs.Files.ReadFile(name)
		if err != nil {
			return report, err
		}
		if err := packformat.ValidateNotice(b); err != nil {
			return report, err
		}
		p.Members["notices/"+name] = b
	}
	decision, err := canonicaljson.Marshal(map[string]any{"schema_id": "cartulary.veris_license_decision.v1", "source_lock_sha256": p.Report.SourceLockSHA256, "output_policy_sha256": Digest(policyBytes), "decision_ref": policy.DecisionRef, "authorization": policy.Authorization, "scope": policy.Scope, "license": policy.License})
	if err != nil {
		return report, err
	}
	p.Artifacts = append(p.Artifacts, packformat.SourceArtifact{Ref: "urn:cartulary:veris:license-decision", Version: VERISSourceProfile, SHA256: Digest(decision)})
	slices.SortFunc(p.Artifacts, func(a, b packformat.SourceArtifact) int { return strings.Compare(a.Ref, b.Ref) })
	m := packformat.Manifest{SchemaID: "cartulary.reference_pack_manifest.v1", Contract: packformat.ContractID, Key: VERISKey, Kind: "framework", Version: in.PackVersion, Sequence: 1, ProfileID: packformat.VERISProfile, ProfileVersion: "2", SourceProfileID: VERISSourceProfile, SourceProfileSHA256: p.Report.SourceProfileSHA256, Repository: &in.Repository, SourceIdentifier: "VERIS Community; vz-risk/veris@" + inventory.Commit, SourceVersion: inventory.Version, SourceAsOf: &inventory.AsOf, BuiltAt: in.BuiltAt, Artifacts: p.Artifacts, Dependencies: []packformat.Dependency{}, Conflicts: []packformat.Conflict{}, License: policy.License, Extensions: map[string]any{}}
	m.Builder.ID, m.Builder.Version, m.Builder.SHA256 = builder.ID, builder.Version, builder.SHA256
	m.Compatibility.Majors, m.Compatibility.Capabilities = []int{1}, []string{}
	objects, relationships := int64(p.Report.Counts["objects"]), int64(p.Report.Counts["relationships"])
	m.Summary.Kind, m.Summary.Objects, m.Summary.Relationships = "objects_relationships", &objects, &relationships
	m.Files = logicalMemberInventory(p.Members)
	raw, err := canonicaljson.Marshal(m)
	if err != nil {
		return report, err
	}
	if _, err := packformat.DecodeManifest(raw, true); err != nil {
		return report, fmt.Errorf("VERIS manifest: %w", err)
	}
	if err := packformat.ValidateContent(ctx, m, p.Members, nil); err != nil {
		return report, fmt.Errorf("VERIS content: %w", err)
	}
	// Publication allocation cannot occur until full production validation passes.
	publication, err := allocateScopedPublication(in.LedgerRoot, in.PackVersion, Digest(raw), "cartulary.veris_publication.v1", in.Repository, VERISKey)
	if err != nil {
		return report, err
	}
	m.Sequence = publication.Sequence
	raw, err = canonicaljson.Marshal(m)
	if err != nil {
		return report, err
	}
	payload, err := packformat.PayloadDigest(m.Files)
	if err != nil {
		return report, err
	}
	audit, err := canonicaljson.Marshal(p.Report)
	if err != nil {
		return report, err
	}
	report = PackageReport{"cartulary.veris_package_report.v1", Digest(raw), payload, p.Report.SourceProfileSHA256, Digest(audit), builder, objects, relationships, m.Files, "logical/" + Digest(raw)}
	reportBytes, err := canonicaljson.Marshal(report)
	if err != nil {
		return report, err
	}
	p.Research["license-decision.json"], p.Research["audit.json"], p.Research["build-report.json"] = decision, audit, reportBytes
	if err := publishTree(in.SourceRoot, "publications/"+report.ManifestSHA256, p.Research); err != nil {
		return report, err
	}
	p.Members["manifest.json"] = raw
	if err := publishTree(in.OutputRoot, report.LogicalPath, p.Members); err != nil {
		return report, err
	}
	return report, nil
}
