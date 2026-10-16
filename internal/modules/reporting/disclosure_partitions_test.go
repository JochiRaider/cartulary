package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func testEffectiveDisclosurePartitions(t *testing.T) {
	snapshot := SnapshotContent{Records: []ReportingRecordSummary{
		{RecordType: "party", SourceRecordRef: SourceRecordRef{SourceFamily: "party", SourceRecordID: "B"}},
		{RecordType: "party", SourceRecordRef: SourceRecordRef{SourceFamily: "party", SourceRecordID: "A"}},
		// An envelope or label is not another Party identity.
		{RecordType: "record_envelope", SourceRecordRef: SourceRecordRef{SourceFamily: "record_envelope", SourceRecordID: "A"}, DisclosurePartitionRefs: []string{"party:invented"}},
	}}
	for _, scope := range []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal} {
		recipients := []string(nil)
		want := []string{"public", "party:A", "party:B", "internal_only"}
		if scope == ReleaseScopeExternal {
			recipients = []string{"party:A"}
			want = []string{"public", "party:A"}
		}
		effective, err := effectiveDisclosurePartitions(snapshot, scope, recipients)
		if err != nil || !reflect.DeepEqual(effective, want) {
			t.Fatal("effective set", scope, effective, err)
		}
		for _, ref := range []string{"public", "party:A", "party:B", "internal_only", "blocked", "party:invented"} {
			eligible := false
			for _, allowed := range want {
				eligible = eligible || ref == allowed
			}
			for _, action := range []string{ActionAllow, ActionDrop, ActionMask, ActionStub, ActionTruncate} {
				t.Run(scope+"/"+ref+"/"+action, func(t *testing.T) {
					profile := internalRedactionProfile()
					profile.AllowedReleaseScopes = []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal}
					// Even an overbroad declared profile cannot authorize another recipient,
					// unknown snapshot Party, blocked content or external internal_only.
					profile.AllowedDisclosurePartitionRefs = orderedDisclosurePartitions([]string{"public", "party:A", "party:B", "party:invented", "internal_only", "blocked"})
					profile.DefaultAction = RedactionActionSpec{Type: action}
					if action == ActionTruncate {
						n := 1
						profile.DefaultAction.MaxChars = &n
					}
					digest, err := ValidateRedactionProfile(profile)
					if err != nil {
						t.Fatal(err)
					}
					view, _, _, err := BuildRedactionProfileView(profile, digest)
					if err != nil {
						t.Fatal(err)
					}
					fields := []ExportField{{Path: "/source", ContentClass: ContentClassDerivedAnalytic, Value: "private source", DisclosurePartitionRefs: []string{ref}}}
					result, err := selectRedactionFields(fields, map[string]string{"/source": "$.sections['s'].blocks['b'].fields['value']"}, view, scope, effective)
					if eligible || action == ActionDrop {
						if err != nil || len(result) != 1 || result[0].Action.Type != action || result[0].Trace.RuleID != "profile_default" {
							t.Fatal("selected action lost", err)
						}
						if action == ActionDrop && result[0].Include {
							t.Fatal("selected drop retained content")
						}
					} else {
						var failure *renderValidationError
						if !errors.As(err, &failure) || failure.FailureCode != "redaction_manifest_invalid" || failure.ReasonCode != "disclosure_partition_unresolved" || len(result) != 0 {
							t.Fatal("scope authority bypass", err)
						}
					}
				})
			}
		}
	}
	for _, refs := range [][]string{{"party:A", "party:A"}, {"party:invented"}, {"party:A:B"}, {"public"}} {
		if _, err := effectiveDisclosurePartitions(snapshot, ReleaseScopeExternal, refs); err == nil {
			t.Fatal("invalid recipient admitted", refs)
		}
	}
	if _, err := effectiveDisclosurePartitions(snapshot, ReleaseScopeInternalReview, []string{"party:A"}); err == nil {
		t.Fatal("internal recipient accepted")
	}
	profile, _, err := ResolveRedactionProfile(InternalRedactionProfileID, DefaultRedactionProfileVersion, nil, snapshot)
	if err != nil || !reflect.DeepEqual(profile.AllowedDisclosurePartitionRefs, []string{"public", "party:A", "party:B", "internal_only"}) || len(profile.Rules) != 0 || profile.DefaultAction.Type != ActionAllow {
		t.Fatal("snapshot-bound passthrough", err)
	}
	// Scope eligibility does not grant a permission withheld by the profile.
	effective := append([]string{}, profile.AllowedDisclosurePartitionRefs...)
	profile.AllowedDisclosurePartitionRefs = []string{"public"}
	digest, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	view, _, _, err := BuildRedactionProfileView(profile, digest)
	if err != nil {
		t.Fatal(err)
	}
	result, err := selectRedactionFields(
		[]ExportField{{Path: "/source", ContentClass: ContentClassSourceEvidence, Value: "private", DisclosurePartitionRefs: []string{"party:A"}}},
		map[string]string{"/source": "$.sections['s'].blocks['b'].fields['value']"},
		view, ReleaseScopeInternalReview, effective)
	var failure *renderValidationError
	if !errors.As(err, &failure) || failure.ReasonCode != "disclosure_partition_unresolved" || len(result) != 0 {
		t.Fatal("effective scope overrode the narrower profile", err)
	}
	snapshot.Records = append(snapshot.Records, snapshot.Records[0])
	if _, err := snapshotPartyPartitions(snapshot); err == nil {
		t.Fatal("duplicate Party source accepted")
	}
}

func testInternalProfileRequestDefault(t *testing.T) {
	for _, id := range []string{InternalRedactionProfileID, ExternalRedactionProfileID, TokenizedRedactionProfileID} {
		for _, version := range []string{"1", "2", "3"} {
			if _, _, err := ResolveRedactionProfile(id, version, nil, SnapshotContent{}); !errors.Is(err, ErrInvalidRedactionProfile) {
				t.Fatal("retired profile identity accepted", id, version, err)
			}
		}
	}
	base := map[string]any{"snapshot_id": fixtureSnapshotID, "client_txn_id": "default-profile", "template_id": DefaultTemplateID, "template_version": DefaultTemplateVersion, "output_kind": OutputKindSlidev}
	decode := func(body map[string]any) (CreateReleaseRequest, bool) {
		t.Helper()
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		request, apiErr := DecodeCreateReleaseRequest(bytes.NewReader(raw))
		return request, apiErr == nil
	}
	for _, scope := range []string{"", ReleaseScopeInternalDraft, ReleaseScopeInternalReview} {
		delete(base, "redaction_profile_id")
		delete(base, "redaction_profile_version")
		delete(base, "release_scope")
		if scope != "" {
			base["release_scope"] = scope
		}
		implicit, ok := decode(base)
		if !ok || implicit.RedactionProfileID != InternalRedactionProfileID || implicit.RedactionProfileVersion != DefaultRedactionProfileVersion {
			t.Fatal("internal default missing", scope)
		}
		base["redaction_profile_id"], base["redaction_profile_version"] = InternalRedactionProfileID, DefaultRedactionProfileVersion
		explicit, ok := decode(base)
		if !ok || !bytes.Equal(implicit.Normalized, explicit.Normalized) {
			t.Fatal("explicit and defaulted requests differ")
		}
	}
	for _, scope := range []string{ReleaseScopeInternalDraft, ReleaseScopeExternal} {
		base["release_scope"] = scope
		for _, partial := range []map[string]any{
			{"redaction_profile_id": InternalRedactionProfileID},
			{"redaction_profile_version": DefaultRedactionProfileVersion},
			{"redaction_profile_id": nil, "redaction_profile_version": DefaultRedactionProfileVersion},
			{"redaction_profile_id": InternalRedactionProfileID, "redaction_profile_version": ""},
		} {
			delete(base, "redaction_profile_id")
			delete(base, "redaction_profile_version")
			for key, value := range partial {
				base[key] = value
			}
			if _, ok := decode(base); ok {
				t.Fatal("incomplete/null profile tuple defaulted")
			}
		}
	}
	delete(base, "redaction_profile_id")
	delete(base, "redaction_profile_version")
	if _, ok := decode(base); ok {
		t.Fatal("external omission defaulted")
	}
}
