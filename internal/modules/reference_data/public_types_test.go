package reference_data

import (
	"bytes"
	"context"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func testOwnerDTOCanonicalRoundTrips(t *testing.T) {
	equalBytes := func(a, b any) {
		t.Helper()
		left, err := canonicaljson.Marshal(a)
		if err != nil {
			t.Fatal(err)
		}
		right, err := canonicaljson.Marshal(b)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(left, right) {
			t.Fatalf("owner representation changed canonical bytes: %s != %s", left, right)
		}
	}
	_, versions, err := loadBuiltinRelease(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var members []packformat.SetMember
	for _, version := range versions {
		members = append(members, packformat.SetMember(memberFor(version.Content)))
		manifest := version.Content.Manifest
		equalBytes(manifest.License, packLicenseFromFormat(manifest.License))
		equalBytes(manifest.Artifacts, sourceArtifactsFromFormat(manifest.Artifacts))
	}
	set, err := packformat.BuildSet(members)
	if err != nil {
		t.Fatal(err)
	}
	equalBytes(set, packSetFromFormat(set))
	equalBytes(set, packSetFromFormat(set).format())
	refs := packformat.PortableReferences{SchemaID: packformat.PortableReferencesSchemaID, Sets: []packformat.Set{set}, Versions: []packformat.PortableVersion{}}
	for _, member := range members {
		refs.Versions = append(refs.Versions, packformat.PortableVersion{SetMember: member, DistributionKind: "packaged_builtin", VerificationMethod: "packaged_release_manifest_v1", SourceProfileID: "source", SourceProfileSHA256: member.ManifestSHA256})
	}
	equalBytes(refs, portableReferencesFromFormat(refs))
	equalBytes(refs, portableReferencesFromFormat(refs).format())
	for _, policy := range packformat.IndicatorPolicies() {
		value, err := packformat.Evaluate(policy, policy.Kinds[0], "invalid value")
		if err != nil {
			t.Fatal(err)
		}
		equalBytes(value, Evaluation(value))
	}
	summary, err := packformat.CheckSummary(context.Background(), "container_bytes", nil, func(_ context.Context, emit packformat.FindingSink) error {
		return emit(packformat.LimitFinding("$", "max_container_bytes", 12, 13))
	})
	if err != nil {
		t.Fatal(err)
	}
	equalBytes(summary, summaryFromFormat(summary))
	for _, notices := range [][]string{nil, {}} {
		for _, bindings := range [][]packformat.LicenseBinding{nil, {}, {{Ref: "LicenseRef-Test", Path: "NOTICE"}}} {
			license := packformat.License{Notices: notices, Bindings: bindings}
			equalBytes(license, packLicenseFromFormat(license))
		}
	}
	for _, sets := range [][]packformat.Set{nil, {}} {
		refs := packformat.PortableReferences{Sets: sets}
		equalBytes(refs, portableReferencesFromFormat(refs))
	}
}
