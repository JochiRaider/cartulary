package reporting

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"
)

func testReferencedDisclosurePartitions(t *testing.T) {
	const host = "00000000-0000-0000-0000-000000007001"
	const first = "00000000-0000-0000-0000-000000007002"
	const second = "00000000-0000-0000-0000-000000007003"
	const unrelated = "00000000-0000-0000-0000-000000007004"
	const link = "00000000-0000-0000-0000-000000007005"
	const mention = "00000000-0000-0000-0000-000000007006"
	source := []ExportField{
		{SourceRecordID: host, SourceRecordType: "host", SourceFamily: "host", Path: "/hosts/" + host, ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "private host"}},
		{SourceRecordID: host, SourceRecordType: "host", SourceFamily: "record_envelope", Path: "/record_envelopes/" + host, ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_type": "host", "value": "host metadata"}},
		{SourceRecordID: first, SourceRecordType: "evidence", SourceFamily: "evidence", Path: "/evidence/" + first, ContentClass: ContentClassSourceEvidence, Value: "first", SupportRefs: []string{"/hosts/" + host, "/evidence/" + second}},
		{SourceRecordID: first, SourceRecordType: "evidence", SourceFamily: "record_envelope", Path: "/record_envelopes/" + first, ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_type": "evidence", "value": "first metadata"}},
		{SourceRecordID: second, SourceRecordType: "evidence", SourceFamily: "evidence", Path: "/evidence/" + second, ContentClass: ContentClassSourceEvidence, Value: "second", SupportRefs: []string{"/evidence/" + first}},
		{SourceRecordID: unrelated, SourceRecordType: "evidence", SourceFamily: "evidence", Path: "/evidence/" + unrelated, ContentClass: ContentClassSourceEvidence, Value: "independent", DisclosurePartitionRefs: []string{"public"}},
		{SourceRecordID: link, SourceRecordType: "record_link", SourceFamily: "record_link", Path: "/relationships/" + link, ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"src_record_id": host, "dst_record_id": first, "link_type": "related"}},
		{SourceRecordID: mention, SourceRecordType: "entity_mention", SourceFamily: "entity_mention", Path: "/entity_mentions/" + mention, ContentClass: ContentClassSourceEvidence, Value: map[string]any{"entity_type": "host", "raw_text": "rough host", "source_record_id": first, "source_field_key": "evidence.description", "resolution_status": "unresolved", "resolved_record_id": nil}},
	}
	snapshot, _, err := buildStructuredSnapshotModel("incident", fixtureSnapshotID, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), "boundary", fixtureSnapshotSource(source))
	if err != nil {
		t.Fatal(err)
	}
	fields, err := snapshot.RedactionFields()
	if err != nil {
		t.Fatal(err)
	}
	refs := fixtureAuthoredSourceRefs(t, snapshot.SnapshotContent, source)
	aliases := map[string]string{}
	for alias, ref := range refs {
		aliases[ref] = alias
	}
	before, _ := canonicalJSON(snapshot)
	indexed := func(fields []ExportField) map[string][]string {
		result := map[string][]string{}
		for _, field := range fields {
			result[aliases[field.Path]] = field.DisclosurePartitionRefs
		}
		return result
	}
	for _, scope := range []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal} {
		projected, err := assignSubjectFieldPartitions(context.Background(), snapshot.SnapshotContent, snapshot.SourceFacts, fields, scope)
		if err != nil {
			t.Fatal(scope, err)
		}
		restriction := "internal_only"
		if scope == ReleaseScopeExternal {
			restriction = "blocked"
		}
		for _, field := range projected {
			restricted := field.SourceRecordID == host || field.SourceRecordID == second || field.SourceRecordID == link || field.SourceRecordID == mention || field.SourceRecordID == first
			if slices.Contains(field.DisclosurePartitionRefs, restriction) != restricted {
				t.Fatal("reference did not preserve the exact dependency restriction", scope, field.Path, field.DisclosurePartitionRefs)
			}
		}
		reversed := slices.Clone(fields)
		slices.Reverse(reversed)
		reordered, err := assignSubjectFieldPartitions(context.Background(), snapshot.SnapshotContent, snapshot.SourceFacts, reversed, scope)
		if err != nil || !reflect.DeepEqual(indexed(projected), indexed(reordered)) {
			t.Fatal("reference cycle result depends on traversal order", err)
		}
	}
	after, _ := canonicalJSON(snapshot)
	if !bytes.Equal(before, after) {
		t.Fatal("reference propagation mutated the frozen model")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if result, err := assignSubjectFieldPartitions(ctx, snapshot.SnapshotContent, snapshot.SourceFacts, fields, ReleaseScopeInternalReview); !errors.Is(err, context.Canceled) || result != nil {
		t.Fatal("canceled propagation returned usable partial output", err)
	}
	// Exercise production selection. A public field may not retain a private
	// support edge even when its own scalar value has no identifying text.
	model, digest, err := bindRenderModel(snapshot, fixtureRenderIdentity(), ReleaseScopeExternal, nil)
	if err != nil {
		t.Fatal(err)
	}
	profile, _, err := ResolveRedactionProfile(ExternalRedactionProfileID, DefaultRedactionProfileVersion, nil, snapshot.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	paths, err := redactionFieldModelPaths(snapshot.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	// Drop every source except the indirect consumer and an independent field.
	// The consumer still cannot be allowed: selected drops do not rewrite
	// frozen provenance or make an originally private citation public.
	for _, field := range fields {
		if field.SourceRecordID == second || field.SourceRecordID == unrelated {
			continue
		}
		path := paths[field.Path]
		profile.Rules = slices.DeleteFunc(profile.Rules, func(rule RedactionRule) bool { return rule.Path != nil && *rule.Path == path })
		profile.Rules = append(profile.Rules, RedactionRule{RuleID: "drop-" + hashHex([]byte(path)), Path: &path, Action: RedactionActionSpec{Type: ActionDrop}})
	}
	profileDigest, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	result, err := RedactExportModel(model, snapshot.SourceFacts, profile, profileDigest, digest, ReleaseScopeExternal, nil)
	var failure *renderValidationError
	if !errors.As(err, &failure) || failure.ReasonCode != "disclosure_partition_unresolved" || result.ModelSHA256 != "" {
		t.Fatal("indirect disclosure produced renderable output", err)
	}
	path := paths[refs["/evidence/"+second]]
	profile.Rules = append(profile.Rules, RedactionRule{RuleID: "drop-consumer", Path: &path, Action: RedactionActionSpec{Type: ActionDrop}})
	profileDigest, err = ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	result, err = RedactExportModel(model, snapshot.SourceFacts, profile, profileDigest, digest, ReleaseScopeExternal, nil)
	if err != nil || len(result.Model.Fields) != 1 || result.Model.Fields[0].Value != "independent" {
		t.Fatal("selected drop did not preserve the independent public field", err)
	}
	// Endpoint identity is an exact frozen binding, not a matching ID in a
	// different family or a syntactically valid missing record.
	for _, change := range []func(*SnapshotContent){
		func(m *SnapshotContent) { m.Relationships[0].SrcRecordRef.SourceRecordRef.SourceFamily = "unknown" },
		func(m *SnapshotContent) { m.Relationships[0].SrcRecordRef.SourceRecordRef.SourceRecordID = "missing" },
		func(m *SnapshotContent) { m.Relationships[0].SrcRecordRef.SourceRecordRef.SourceSnapshotID = "foreign" },
	} {
		candidate, err := copyReportingModel(snapshot.SnapshotContent)
		if err != nil {
			t.Fatal(err)
		}
		change(&candidate)
		if result, err := assignReferencedFieldPartitions(context.Background(), candidate, fields, ReleaseScopeExternal, nil); err == nil || result != nil {
			t.Fatal("invalid endpoint yielded a partial disclosure result")
		}
	}
}
