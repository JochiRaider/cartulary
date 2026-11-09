package producer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/referencepacks/attack/internal/native"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	packformat "github.com/JochiRaider/cartulary/referencepacks/format/envelope"
)

const FixtureRepository = DatasetFixtureRepository

type BuildParameters struct {
	SourceRoot    string `json:"source_root"`
	SourceLock    string `json:"source_lock"`
	SourceProfile string `json:"source_profile"`
	FieldPolicy   string `json:"field_policy"`
	OutputRoot    string `json:"output_root"`
	BuiltAt       string `json:"built_at"`
	PackVersion   string `json:"pack_version"`
	LedgerRoot    string `json:"ledger_root"`
	BuilderSHA256 string `json:"builder_sha256"`
}

type PackageReport struct {
	SchemaID            string            `json:"schema_id"`
	ManifestSHA256      string            `json:"manifest_sha256"`
	PayloadSHA256       string            `json:"payload_sha256"`
	SourceProfileSHA256 string            `json:"source_profile_sha256"`
	SourceAuditSHA256   string            `json:"source_audit_sha256"`
	Builder             Builder           `json:"builder"`
	Objects             int64             `json:"object_count"`
	Relationships       int64             `json:"relationship_count"`
	Files               []packformat.File `json:"files"`
	LogicalPath         string            `json:"logical_path"`
}

type packageSource map[string][]byte

func (s packageSource) Open(_ context.Context, name string) (io.ReadCloser, error) {
	b, ok := s[name]
	if !ok {
		return nil, errors.New("missing logical member")
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

type FixtureBuildInput struct {
	SchemaID string `json:"schema_id"`
	BuildParameters
}
type ReleaseBuildInput struct {
	SchemaID string `json:"schema_id"`
	BuildParameters
	OutputPolicy string `json:"output_policy"`
}
type OutputPolicy struct {
	SchemaID            string `json:"schema_id"`
	SourceLockSHA256    string `json:"source_lock_sha256"`
	SourceProfileSHA256 string `json:"source_profile_sha256"`
	Redistribution      string `json:"redistribution"`
	DecisionRef         string `json:"decision_ref"`
}
type packageInput struct {
	BuildParameters
	Repository, Redistribution string
	OutputPolicy               []byte
}

func AssembleFixture(ctx context.Context, raw []byte, builder Builder) (PackageReport, error) {
	var in FixtureBuildInput
	if err := DecodeClosed(raw, &in); err != nil {
		return PackageReport{}, err
	}
	if in.SchemaID != "cartulary.attack_fixture_build_input.v3" {
		return PackageReport{}, errors.New("explicit fixture build input required")
	}
	return assemble(ctx, packageInput{BuildParameters: in.BuildParameters, Repository: DatasetFixtureRepository, Redistribution: "restricted"}, builder)
}
func AssembleRelease(ctx context.Context, raw []byte, builder Builder) (PackageReport, error) {
	var in ReleaseBuildInput
	if err := DecodeClosed(raw, &in); err != nil {
		return PackageReport{}, err
	}
	if in.SchemaID != "cartulary.attack_release_build_input.v3" {
		return PackageReport{}, errors.New("explicit release build input required")
	}
	policy, err := readRegular(in.OutputPolicy)
	if err != nil {
		return PackageReport{}, err
	}
	var decision OutputPolicy
	if err = DecodeClosed(policy, &decision); err != nil {
		return PackageReport{}, err
	}
	if decision.SchemaID != "cartulary.attack_output_policy.v1" || decision.DecisionRef == "" || (decision.Redistribution != "allowed" && decision.Redistribution != "restricted") {
		return PackageReport{}, errors.New("explicit output disposition required; candidate construction grants no distribution approval")
	}
	return assemble(ctx, packageInput{BuildParameters: in.BuildParameters, Repository: ReleaseRepository, Redistribution: decision.Redistribution, OutputPolicy: policy}, builder)
}
func assemble(ctx context.Context, input packageInput, builder Builder) (PackageReport, error) {
	if input.BuilderSHA256 != builder.SHA256 || !validBuildTime(input.BuiltAt) {
		return PackageReport{}, errors.New("exact builder and build time required")
	}

	lockBytes, err := readRegular(input.SourceLock)
	if err != nil {
		return PackageReport{}, err
	}
	profileBytes, err := readRegular(input.SourceProfile)
	if err != nil {
		return PackageReport{}, err
	}
	policy, err := readRegular(input.FieldPolicy)
	if err != nil {
		return PackageReport{}, err
	}
	inputs, err := admitATTACKInputs(lockBytes, profileBytes, policy)
	if err != nil {
		return PackageReport{}, err
	}
	if len(input.OutputPolicy) > 0 {
		var decision OutputPolicy
		if err = DecodeClosed(input.OutputPolicy, &decision); err != nil {
			return PackageReport{}, err
		}
		if decision.SourceLockSHA256 != Digest(lockBytes) || decision.SourceProfileSHA256 != Digest(profileBytes) {
			return PackageReport{}, errors.New("output policy must bind exact source lock and profile")
		}
	}
	l := inputs.lock
	root, err := os.OpenRoot(input.SourceRoot)
	if err != nil {
		return PackageReport{}, err
	}
	defer root.Close()
	audit, err := auditATTACKSources(ctx, root, inputs, input.BuiltAt, builder)
	if err != nil {
		return PackageReport{}, err
	}
	auditBytes, err := canonicaljson.Marshal(audit)
	if err != nil {
		return PackageReport{}, err
	}
	publicationBytes, err := canonicaljson.Marshal(map[string]any{"source_lock_sha256": Digest(lockBytes), "source_profile_sha256": Digest(profileBytes), "field_policy_sha256": Digest(policy), "built_at": input.BuiltAt, "builder_sha256": builder.SHA256, "pack_version": input.PackVersion, "trust_repository_id": input.Repository, "redistribution": input.Redistribution})
	if err != nil {
		return PackageReport{}, err
	}
	receipt, err := Check(root, lockBytes)
	if err != nil {
		return PackageReport{}, err
	}
	artifacts := []packformat.SourceArtifact{}
	indexes := map[string]int{}
	for _, a := range l.Artifacts {
		if a.Domain != "license" {
			artifacts = append(artifacts, packformat.SourceArtifact{Ref: a.URL, Version: l.CatalogVersion, SHA256: a.SHA256})
		}
	}
	if len(input.OutputPolicy) > 0 {
		artifacts = append(artifacts, packformat.SourceArtifact{Ref: "urn:cartulary:attack:output-policy", Version: inputs.profileID, SHA256: Digest(input.OutputPolicy)})
	}
	slices.SortFunc(artifacts, func(a, b packformat.SourceArtifact) int { return strings.Compare(a.Ref, b.Ref) })
	for i, a := range artifacts {
		for _, src := range l.Artifacts {
			if src.URL == a.Ref {
				indexes[src.Name] = i
			}
		}
	}
	records := map[string]map[string]any{}
	domains := map[string][]string{}
	members := packageSource{}
	for i, a := range l.Artifacts {
		b, e := ReadEvidence(root, receipt.Sources[i])
		if e != nil {
			return PackageReport{}, e
		}
		if a.Domain == "license" {
			members["notices/MITRE-ATTACK.txt"] = b
			continue
		}
		v, e := canonicaljson.DecodeStrict(b)
		if e != nil {
			return PackageReport{}, e
		}
		for _, value := range v.(map[string]any)["objects"].([]any) {
			o := value.(map[string]any)
			id := o["id"].(string)
			records[id] = o
			domains[id] = append(domains[id], a.Domain)
		}
	}
	replacements := map[string]string{}
	for _, s := range records {
		if s["type"] == "relationship" && s["relationship_type"] == "revoked-by" && !native.ATTACKDeprecated(s) {
			replacements[s["source_ref"].(string)] = s["target_ref"].(string)
		}
	}
	refsFor := func(id string) []map[string]any {
		refs := []map[string]any{}
		for _, locator := range audit.RecordLocators[id] {
			refs = append(refs, map[string]any{"artifact_index": indexes[locator.Artifact], "locator": locator.Pointer})
		}
		slices.SortFunc(refs, func(a, b map[string]any) int {
			if a["artifact_index"] != b["artifact_index"] {
				return a["artifact_index"].(int) - b["artifact_index"].(int)
			}
			return strings.Compare(a["locator"].(string), b["locator"].(string))
		})
		return refs
	}
	objects := map[string]map[string]any{}
	relationships := []map[string]any{}
	for id, s := range records {
		if s["type"] == "relationship" {
			refs := refsFor(id)
			typ, from, to := s["relationship_type"].(string), s["source_ref"].(string), s["target_ref"].(string)
			rid, e := native.RelationshipID("cartulary.reference_pack.framework.attack.v2", typ, from, to, refs)
			if e != nil {
				return PackageReport{}, e
			}
			relationships = append(relationships, map[string]any{"schema_id": "cartulary.reference_pack.framework.attack.relationship.v2", "relationship_id": rid, "relationship_type": typ, "source_object_id": from, "target_object_id": to, "description": native.ATTACKDescription(s), "source_refs": refs, "extensions": map[string]any{}, "source_kind": "stix_sro", "source": s})
			continue
		}
		var replacement any
		if target := replacements[id]; target != "" {
			replacement = target
		}
		slices.Sort(domains[id])
		aliases := audit.LookupAliases[id]
		if aliases == nil {
			aliases = []string{}
		}
		o, e := native.CanonicalMap(map[string]any{"schema_id": "cartulary.reference_pack.framework.attack.object.v2", "object_id": id, "object_type": s["type"], "display_label": native.ATTACKLabel(s), "description": native.ATTACKDescription(s), "aliases": aliases, "deprecated": native.ATTACKDeprecated(s), "replacement_object_id": replacement, "external_refs": native.ATTACKExternalRefs(s), "source_refs": refsFor(id), "extensions": map[string]any{}, "source": s, "source_domains": slices.Compact(domains[id])})
		if e != nil {
			return PackageReport{}, e
		}
		objects[id] = o
	}
	embedded, err := native.ATTACKEmbedded(objects)
	if err != nil {
		return PackageReport{}, err
	}
	relationships = append(relationships, embedded...)
	ids := make([]string, 0, len(objects))
	for id := range objects {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	var objectBytes, relationshipBytes bytes.Buffer
	for _, id := range ids {
		b, e := canonicaljson.Marshal(objects[id])
		if e != nil {
			return PackageReport{}, e
		}
		objectBytes.Write(b)
		objectBytes.WriteByte('\n')
	}
	slices.SortFunc(relationships, func(a, b map[string]any) int {
		for _, k := range []string{"relationship_type", "source_object_id", "target_object_id", "relationship_id"} {
			if c := strings.Compare(a[k].(string), b[k].(string)); c != 0 {
				return c
			}
		}
		return 0
	})
	for _, r := range relationships {
		b, e := canonicaljson.Marshal(r)
		if e != nil {
			return PackageReport{}, e
		}
		relationshipBytes.Write(b)
		relationshipBytes.WriteByte('\n')
	}
	members["payload/objects.ndjson"] = objectBytes.Bytes()
	members["payload/relationships.ndjson"] = relationshipBytes.Bytes()
	m := packformat.Manifest{SchemaID: "cartulary.reference_pack_manifest.v2", Contract: packformat.ContractID, Key: "framework.attack", Kind: "framework", Version: input.PackVersion, Sequence: 1, ProfileID: "cartulary.reference_pack.framework.attack.v2", ProfileVersion: "2", SourceProfileID: inputs.profileID, SourceProfileSHA256: Digest(profileBytes), Repository: &input.Repository, SourceIdentifier: l.Repository, SourceVersion: l.CatalogVersion, SourceAsOf: &l.PublishedAt, BuiltAt: input.BuiltAt, Artifacts: artifacts, Dependencies: []packformat.Dependency{}, Conflicts: []packformat.Conflict{}, Extensions: map[string]any{}}
	m.Builder.ID = builder.ID
	m.Builder.Version = builder.Version
	m.Builder.SHA256 = builder.SHA256
	m.Compatibility.Majors = []int{1}
	m.Compatibility.Capabilities = []string{}
	countObjects, countRelationships := int64(len(objects)), int64(len(relationships))
	m.Summary.Kind = "objects_relationships"
	m.Summary.Objects = &countObjects
	m.Summary.Relationships = &countRelationships
	m.License = packformat.License{Expression: "LicenseRef-MITRE-ATTACK", ListVersion: "3.28.0", Redistribution: input.Redistribution, Notices: []string{"notices/MITRE-ATTACK.txt"}, Bindings: []packformat.LicenseBinding{{Ref: "LicenseRef-MITRE-ATTACK", Path: "notices/MITRE-ATTACK.txt"}}}
	if _, err := projectDataset(ctx, &m, members); err != nil {
		return PackageReport{}, err
	}
	m.Files = logicalMemberInventory(members)
	manifest, err := canonicaljson.Marshal(m)
	if err != nil {
		return PackageReport{}, err
	}
	if _, err = packformat.DecodeManifest(manifest, true); err != nil {
		return PackageReport{}, fmt.Errorf("manifest: %w", err)
	}
	if err = packformat.ValidateNotice(members["notices/MITRE-ATTACK.txt"]); err != nil {
		return PackageReport{}, err
	}
	published, err := finishLogical(ctx, m, members, input.LedgerRoot, input.OutputRoot)
	if err != nil {
		return PackageReport{}, err
	}
	m = published.Manifest
	manifest, err = canonicaljson.Marshal(m)
	if err != nil {
		return PackageReport{}, err
	}
	payload := published.PayloadSHA256
	report := PackageReport{SchemaID: "cartulary.attack_package_report.v3", ManifestSHA256: Digest(manifest), PayloadSHA256: payload, SourceProfileSHA256: Digest(profileBytes), SourceAuditSHA256: Digest(auditBytes), Builder: builder, Objects: countObjects, Relationships: countRelationships, Files: m.Files, LogicalPath: "logical/" + Digest(manifest)}
	reportBytes, err := canonicaljson.Marshal(report)
	if err != nil {
		return PackageReport{}, err
	}
	research := map[string][]byte{"source-audit.json": auditBytes, "build-report.json": reportBytes, "source-profile.json": profileBytes, "source-lock.json": lockBytes, "field-policy.json": policy, "publication-input.json": publicationBytes}
	if len(input.OutputPolicy) > 0 {
		research["output-policy.json"] = input.OutputPolicy
	}
	if err = publishTree(input.OutputRoot, "research/"+report.ManifestSHA256, research); err != nil {
		return report, err
	}
	return report, nil
}

func validBuildTime(value string) bool {
	t, e := time.Parse(time.RFC3339, value)
	return e == nil && t.UTC().Format(time.RFC3339) == value
}
