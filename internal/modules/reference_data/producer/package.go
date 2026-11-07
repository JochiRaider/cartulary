package producer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const FixtureRepository = "fixture.attack.native.v2"

type PackageInput struct {
	SchemaID       string `json:"schema_id"`
	SourceRoot     string `json:"source_root"`
	SourceLock     string `json:"source_lock"`
	SourceProfile  string `json:"source_profile"`
	FieldPolicy    string `json:"field_policy"`
	OutputRoot     string `json:"output_root"`
	BuiltAt        string `json:"built_at"`
	PackVersion    string `json:"pack_version"`
	LedgerRoot     string `json:"ledger_root"`
	BuilderSHA256  string `json:"builder_sha256"`
	Repository     string `json:"trust_repository_id"`
	Redistribution string `json:"redistribution"`
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

func Assemble(ctx context.Context, inputBytes []byte, builder Builder) (PackageReport, error) {
	var input PackageInput
	if err := DecodeClosed(inputBytes, &input); err != nil {
		return PackageReport{}, err
	}
	if input.SchemaID != "cartulary.attack_package_input.v1" || input.Repository != FixtureRepository || input.Redistribution != "restricted" || input.BuilderSHA256 != builder.SHA256 || !validBuildTime(input.BuiltAt) {
		return PackageReport{}, errors.New("isolated fixture publication identity and exact builder required")
	}
	lockBytes, err := os.ReadFile(input.SourceLock)
	if err != nil {
		return PackageReport{}, err
	}
	l, err := ReadLock(lockBytes)
	if err != nil {
		return PackageReport{}, err
	}
	profileBytes, err := os.ReadFile(input.SourceProfile)
	if err != nil {
		return PackageReport{}, err
	}
	profileValue, err := canonicaljson.DecodeStrict(profileBytes)
	if err != nil {
		return PackageReport{}, err
	}
	profile, ok := profileValue.(map[string]any)
	if !ok || profile["status"] != "adopted_frozen" || profile["source_profile_id"] != "cartulary.attack.all-domains.native.v2.r2" {
		return PackageReport{}, errors.New("frozen adopted source profile required")
	}
	locks, ok := profile["accepted_source_lock_sha256"].([]any)
	if !ok || !slices.Contains(locks, any(Digest(lockBytes))) {
		return PackageReport{}, errors.New("source lock outside frozen profile")
	}
	policy, err := os.ReadFile(input.FieldPolicy)
	if err != nil {
		return PackageReport{}, err
	}
	if profile["field_policy_sha256"] != Digest(policy) {
		return PackageReport{}, errors.New("field policy identity mismatch")
	}
	schemaBindings, ok := profile["adopted_schema_sha256"].(map[string]any)
	if !ok || len(schemaBindings) != 3 {
		return PackageReport{}, errors.New("native schema bindings required")
	}
	for _, name := range []string{"profiles/framework.attack.native_source.v2.schema.json", "profiles/framework.attack.object.v2.schema.json", "profiles/framework.attack.relationship.v2.schema.json"} {
		if schemaBindings[name] != packformat.ProjectionDigest(name) {
			return PackageReport{}, errors.New("frozen native schema identity mismatch")
		}
	}
	root, err := os.OpenRoot(input.SourceRoot)
	if err != nil {
		return PackageReport{}, err
	}
	defer root.Close()
	audit, err := Audit(ctx, root, lockBytes, policy, input.BuiltAt, builder)
	if err != nil {
		return PackageReport{}, err
	}
	auditBytes, err := canonicaljson.Marshal(audit)
	if err != nil {
		return PackageReport{}, err
	}
	if err = Retain(root, "audits/"+Digest(auditBytes)+".json", auditBytes, 0600); err != nil {
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
		if s["type"] == "relationship" && s["relationship_type"] == "revoked-by" && !packformat.ATTACKDeprecated(s) {
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
			rid, e := packformat.RelationshipID("cartulary.reference_pack.framework.attack.v2", typ, from, to, refs)
			if e != nil {
				return PackageReport{}, e
			}
			relationships = append(relationships, map[string]any{"schema_id": "cartulary.reference_pack.framework.attack.relationship.v2", "relationship_id": rid, "relationship_type": typ, "source_object_id": from, "target_object_id": to, "description": packformat.ATTACKDescription(s), "source_refs": refs, "extensions": map[string]any{}, "source_kind": "stix_sro", "source": s})
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
		o, e := packformat.CanonicalMap(map[string]any{"schema_id": "cartulary.reference_pack.framework.attack.object.v2", "object_id": id, "object_type": s["type"], "display_label": packformat.ATTACKLabel(s), "description": packformat.ATTACKDescription(s), "aliases": aliases, "deprecated": packformat.ATTACKDeprecated(s), "replacement_object_id": replacement, "external_refs": packformat.ATTACKExternalRefs(s), "source_refs": refsFor(id), "extensions": map[string]any{}, "source": s, "source_domains": slices.Compact(domains[id])})
		if e != nil {
			return PackageReport{}, e
		}
		objects[id] = o
	}
	embedded, err := packformat.ATTACKEmbedded(objects)
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
	m := packformat.Manifest{SchemaID: "cartulary.reference_pack_manifest.v1", Contract: packformat.ContractID, Key: "framework.attack", Kind: "framework", Version: input.PackVersion, Sequence: 1, ProfileID: "cartulary.reference_pack.framework.attack.v2", ProfileVersion: "2", SourceProfileID: profile["source_profile_id"].(string), SourceProfileSHA256: Digest(profileBytes), Repository: &input.Repository, SourceIdentifier: l.Repository, SourceVersion: l.CatalogVersion, SourceAsOf: &l.PublishedAt, BuiltAt: input.BuiltAt, Artifacts: artifacts, Dependencies: []packformat.Dependency{}, Conflicts: []packformat.Conflict{}, Extensions: map[string]any{}}
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
	for name, b := range members {
		role, media := "payload", "application/x-ndjson"
		if strings.HasPrefix(name, "notices/") {
			role, media = "notice", "text/plain"
		}
		m.Files = append(m.Files, packformat.File{Path: name, Role: role, MediaType: media, Size: int64(len(b)), SHA256: Digest(b)})
	}
	slices.SortFunc(m.Files, func(a, b packformat.File) int { return strings.Compare(a.Path, b.Path) })
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
	if err = packformat.ValidateContent(ctx, m, members, nil); err != nil {
		var paths []string
		_ = packformat.ContentSchemaFindings(ctx, m, members, func(f packformat.Finding) error {
			if len(paths) < 10 {
				paths = append(paths, f.Path)
			}
			return nil
		})
		return PackageReport{}, fmt.Errorf("canonical content: %w; schema paths=%v", err, paths)
	}
	publication, err := allocatePublication(input.LedgerRoot, input.PackVersion, Digest(publicationBytes))
	if err != nil {
		return PackageReport{}, err
	}
	m.Sequence = publication.Sequence
	manifest, err = canonicaljson.Marshal(m)
	if err != nil {
		return PackageReport{}, err
	}
	payload, err := packformat.PayloadDigest(m.Files)
	if err != nil {
		return PackageReport{}, err
	}
	members["manifest.json"] = manifest
	if err = os.MkdirAll(input.OutputRoot, 0700); err != nil {
		return PackageReport{}, err
	}
	out, err := os.OpenRoot(input.OutputRoot)
	if err != nil {
		return PackageReport{}, err
	}
	defer out.Close()
	report := PackageReport{SchemaID: "cartulary.attack_package_report.v1", ManifestSHA256: Digest(manifest), PayloadSHA256: payload, SourceProfileSHA256: Digest(profileBytes), SourceAuditSHA256: Digest(auditBytes), Builder: builder, Objects: countObjects, Relationships: countRelationships, Files: m.Files, LogicalPath: "logical/" + Digest(manifest)}
	for name, b := range members {
		if err = Retain(out, report.LogicalPath+"/"+name, b, 0600); err != nil {
			return PackageReport{}, err
		}
	}
	reportBytes, err := canonicaljson.Marshal(report)
	if err != nil {
		return PackageReport{}, err
	}
	if err = Retain(out, report.LogicalPath+"/build-report.json", reportBytes, 0600); err != nil {
		return PackageReport{}, err
	}
	for name, b := range map[string][]byte{"source-profile.json": profileBytes, "source-lock.json": lockBytes, "field-policy.json": policy, "publication-input.json": publicationBytes} {
		if err = Retain(out, report.LogicalPath+"/evidence/"+name, b, 0600); err != nil {
			return PackageReport{}, err
		}
	}
	return report, nil
}

func validBuildTime(value string) bool {
	t, e := time.Parse(time.RFC3339, value)
	return e == nil && t.UTC().Format(time.RFC3339) == value
}
