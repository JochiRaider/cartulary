package reporting

import (
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	"sort"
	"strings"
	"testing"
	"time"
)

func fixtureStructuredRenderModel(t testing.TB, fields []ExportField) ExportModel {
	t.Helper()
	model, _, err := buildStructuredTestExportModel("00000000-0000-0000-0000-000000000001", fixtureSnapshotID,
		time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC), SourceBoundaryTokenPrefix+strings.Repeat("0", 64), ReleaseScopeInternalReview, nil, fields)
	if err != nil {
		t.Fatal(err)
	}
	return model
}

func fixtureWithFields(t testing.TB, model ExportModel, additions ...ExportField) ExportModel {
	t.Helper()
	fields := fixtureSourceFields(t, model.SnapshotContent)
	return fixtureStructuredRenderModel(t, append(fields, additions...))
}

// Rebuilding an authored test source is distinct from reading frozen refs.
// Keep this explicit translation out of production admission and replay.
func fixtureSourceFields(t testing.TB, model SnapshotContent) []ExportField {
	t.Helper()
	fields, err := model.RedactionFields()
	if err != nil {
		t.Fatal(err)
	}
	selectors := map[string]string{}
	roots := map[string]string{"incident_metadata": "incident", "record_envelope": "record_envelopes", "host": "hosts", "identity": "identities", "party": "parties", "timeline_event": "timeline", "evidence": "evidence"}
	for _, ref := range model.SupportIndex {
		_, id, _ := strings.Cut(ref.SupportTargetRef, ":")
		var selector string
		switch ref.SupportKind {
		case "evidence_item":
			selector = "/evidence/" + id
		case "timeline_event":
			selector = "/timeline/" + id
		case "relationship":
			selector = "/relationships/" + id
		case "source_record":
			family, record, _ := strings.Cut(id, ":")
			root := roots[family]
			if root == "" {
				t.Fatalf("fixture has no source selector for %q", family)
			}
			selector = "/" + root + "/" + record
		default:
			t.Fatalf("fixture has no source selector for %q", ref.SupportKind)
		}
		selectors[ref.SupportRefID] = selector
	}
	for i := range fields {
		source := reportingFieldSources(model)[fields[i].Path].Field
		ref := snapshotSourceIndex(model)[fields[i].Path]
		fields[i].Path = *ref.SourcePath
		if source.FieldKey != "value" {
			fields[i].Value = map[string]any{source.FieldKey: fields[i].Value}
		}
		for j, ref := range fields[i].SupportRefs {
			fields[i].SupportRefs[j] = selectors[ref]
		}
	}
	return fields
}

const fixtureSnapshotID = "00000000-0000-0000-0000-000000000402"

func fixtureRenderIdentity() RenderIdentity {
	id := "00000000-0000-0000-0000-000000000403"
	return RenderIdentity{ReleaseID: &id, RenderAdmittedAt: "2026-05-23T12:00:00.000000Z"}
}

func bindFixtureRenderModel(content SnapshotContent) ExportModel {
	if content.SnapshotID == "" {
		content.SnapshotID = fixtureSnapshotID
	}
	if content.SnapshotAt.IsZero() {
		content.SnapshotAt = time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	}
	content.DerivationVersion = DerivationVersion
	id, err := snapshotModelID(content.SnapshotID, content.DerivationVersion, content.SnapshotAt)
	if err != nil {
		panic(err)
	}
	model, _, err := bindRenderModel(SnapshotModel{SourceFacts: fixtureSourceFactsForContent(content), SchemaID: SnapshotModelSchemaID, SnapshotModelID: id, SnapshotContent: content}, fixtureRenderIdentity(), ReleaseScopeInternalReview, nil)
	if err != nil {
		panic(err)
	}
	return model
}

func buildStructuredTestExportModel(incidentID, snapshotID string, at time.Time, watermark, scope string, partitions []string, fields []ExportField) (ExportModel, string, error) {
	snapshot, _, err := buildStructuredSnapshotModel(incidentID, snapshotID, at, watermark, fixtureSnapshotSource(fixtureOwnerFields(fields)))
	if err != nil {
		return ExportModel{}, "", err
	}
	return bindRenderModel(snapshot, fixtureRenderIdentity(), scope, partitions)
}

func fixtureOutputOptions(kind, scope string) []byte {
	options, err := materializeOutputOptions(nil, kind, scope)
	if err != nil {
		panic(err)
	}
	return options
}

// Test-only shorthand for fixtures whose authored source IDs are their second
// path segments. Production materialization consumes declared RecordFacts.
func fixtureOwnerFields(fields []ExportField) []ExportField {
	fields = cloneExportFields(fields)
	for i := range fields {
		if fields[i].SourceFamily == "" {
			fields[i].SourceFamily = sourceFamilyForPath(fields[i].Path)
		}
		if fields[i].SourceFamily == "incident_metadata" {
			continue
		}
		if fields[i].SourceRecordID == "" {
			parts := strings.Split(strings.TrimPrefix(fields[i].Path, "/"), "/")
			if len(parts) >= 2 {
				fields[i].SourceRecordID = parts[1]
			}
		}
		if fields[i].SourceRecordType == "" {
			fields[i].SourceRecordType = coreTypeForSourceFamily(fields[i].SourceFamily)
			if fields[i].SourceRecordType == "" {
				fields[i].SourceRecordType = fields[i].SourceFamily
			}
		}
	}
	return fields
}

// Explicit fixture facts for tests concerned with other model obligations.
// Production always obtains these facts from the Timeline source owner.
func fixtureSnapshotSource(fields []ExportField) SnapshotSource {
	source := SnapshotSource{SourceFacts: exportprovider.EmptySourceFacts(), Fields: fields}
	seen := map[string]bool{}
	for _, field := range fields {
		if field.SourceFamily == "party" && field.SourceRecordID != "" && !seen["party:"+field.SourceRecordID] {
			seen["party:"+field.SourceRecordID] = true
			source.SourceFacts.Parties = append(source.SourceFacts.Parties, exportprovider.PartyFact{PartyID: field.SourceRecordID, PartitionSegment: field.SourceRecordID})
		}
		if field.SourceFamily != "timeline_event" || field.SourceRecordID == "" || seen[field.SourceRecordID] {
			continue
		}
		seen[field.SourceRecordID] = true
		source.Timeline = append(source.Timeline, exportprovider.TimelineSortFact{
			RecordID: field.SourceRecordID, RecordCreatedAt: "2026-05-01T00:00:00.000000Z", ActivityParseState: "missing", TimeBucket: "unresolved_time",
		})
	}
	source.SourceFacts.Sort()
	return source
}

func fixtureSourceFactsForContent(content SnapshotContent) exportprovider.SourceFacts {
	facts := exportprovider.EmptySourceFacts()
	for _, record := range content.Records {
		if record.SourceRecordRef.SourceFamily == "party" && record.RecordType == "party" {
			facts.Parties = append(facts.Parties, exportprovider.PartyFact{PartyID: record.RecordID, PartitionSegment: record.RecordID})
		}
	}
	facts.Sort()
	return facts
}

func fixtureSourceFactsDigest(content SnapshotContent) string {
	digest, err := sourceFactsDigest(fixtureSourceFactsForContent(content))
	if err != nil {
		panic(err)
	}
	return digest
}

// Assertions describe the source obligation while artifacts use model
// addresses. Keep the translation in fixtures, never in an artifact reader.
func fixtureSourcesByModelPath(t testing.TB, content SnapshotContent, selectors ...string) map[string]string {
	t.Helper()
	paths := fixtureModelPaths(t, content, selectors...)
	sources := map[string]string{}
	for source, path := range paths {
		sources[path] = source
	}
	return sources
}

// Tests supply their own authored selectors in sorted projection order. The
// model does not retain them, and assertions must not reconstruct them from
// production reference IDs or paths.
func fixtureModelPaths(t testing.TB, content SnapshotContent, selectors ...string) map[string]string {
	t.Helper()
	paths, err := redactionFieldModelPaths(content)
	if err != nil {
		t.Fatal(err)
	}
	ordered := make([]string, 0, len(paths))
	for _, path := range paths {
		ordered = append(ordered, path)
	}
	sort.Strings(ordered)
	if len(ordered) != len(selectors) {
		t.Fatalf("fixture source count %d differs from model %d", len(selectors), len(ordered))
	}
	result := map[string]string{}
	for i, selector := range selectors {
		result[selector] = ordered[i]
	}
	return result
}

func fixtureAuthoredSourceRefs(t testing.TB, content SnapshotContent, source []ExportField) map[string]string {
	t.Helper()
	source = cloneExportFields(source)
	sort.Slice(source, func(i, j int) bool { return source[i].Path < source[j].Path })
	selectors := []string{}
	for _, field := range source {
		_, refs, err := projectReportingSourceFields(field)
		if err != nil {
			t.Fatal(err)
		}
		selectors = append(selectors, refs...)
	}
	paths := fixtureModelPaths(t, content, selectors...)
	refs := map[string]string{}
	for _, ref := range content.SourceIndex {
		if ref.SourcePath != nil {
			refs[*ref.SourcePath] = ref.SourceRefID
		}
	}
	result := map[string]string{}
	for selector, path := range paths {
		result[selector] = refs[path]
	}
	return result
}

// Scalar-action fixtures must explicitly preserve the typed envelopes. This
// fixture policy is not a production fallback for unsupported container actions.
func fixtureAllowContainerRules(t testing.TB, profile *RedactionProfile, source SnapshotContent) {
	t.Helper()
	candidates, err := redactionTreeCandidates(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range candidates {
		if candidate.kind == "field" {
			continue
		}
		path := candidate.path
		exists := false
		for _, rule := range profile.Rules {
			if rule.Path != nil && *rule.Path == path {
				exists = true
			}
		}
		if exists {
			continue
		}
		profile.Rules = append(profile.Rules, RedactionRule{RuleID: "fixture-container-allow-" + hashHex([]byte(path)), Path: &path, Action: RedactionActionSpec{Type: ActionAllow}})
	}
}
