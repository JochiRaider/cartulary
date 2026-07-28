package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestImportTargetRegistryDerivationIsDeterministicAndComplete(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	first, err := materializeImportTargetArtifacts(root)
	if err != nil {
		t.Fatalf("materialize import targets: %v", err)
	}
	second, err := materializeImportTargetArtifacts(root)
	if err != nil {
		t.Fatalf("materialize import targets again: %v", err)
	}
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("encode first materialization: %v", err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("encode second materialization: %v", err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatal("import-target materialization is not byte-deterministic")
	}
	if first.Registry.SchemaID != "cartulary.import_target_registry.v1" ||
		len(first.Registry.Rows) != 18 ||
		len(first.Adapters.Adapters) != 18 ||
		len(first.Verification.Targets) != 18 {
		t.Fatalf(
			"unexpected projection identity or counts: registry=%s rows=%d adapters=%d verification=%d",
			first.Registry.SchemaID,
			len(first.Registry.Rows),
			len(first.Adapters.Adapters),
			len(first.Verification.Targets),
		)
	}
	if first.Registry.SourceSHA256 == "" ||
		first.Registry.SourceSHA256 != first.Adapters.SourceSHA256 ||
		first.Registry.SourceSHA256 != first.Verification.SourceSHA256 {
		t.Fatal("generated projections do not share one source digest")
	}

	dispositions := map[string]int{}
	seenRows := map[string]struct{}{}
	seenVerification := map[string]struct{}{}
	for index, row := range first.Registry.Rows {
		if row.RegistryOrder != index || row.RowSHA256 == "" {
			t.Fatalf("row %d has unstable order or missing digest: %#v", index, row)
		}
		if _, duplicate := seenRows[row.TargetID]; duplicate {
			t.Fatalf("duplicate generated target %s", row.TargetID)
		}
		seenRows[row.TargetID] = struct{}{}
		dispositions[row.PublicProjectionDisposition] += 1
		verification := first.Verification.Targets[index]
		if verification.TargetID != row.TargetID || verification.RegistryOrder != index {
			t.Fatalf("verification row %d does not match registry row", index)
		}
		if _, duplicate := seenVerification[verification.VerificationID]; duplicate {
			t.Fatalf("duplicate verification identity %s", verification.VerificationID)
		}
		seenVerification[verification.VerificationID] = struct{}{}
	}
	if dispositions["selectable"] != 14 ||
		dispositions["hidden_reserved"] != 3 ||
		dispositions["extension_claim_gated"] != 1 {
		t.Fatalf("unexpected public dispositions: %#v", dispositions)
	}

	analytical := first.Registry.Rows[17]
	if analytical.TargetID != "network_flow_table:network_flow_activity" ||
		analytical.FacadeBindingID == nil ||
		*analytical.FacadeBindingID != "network_flow_activity.import_facade.v1" ||
		analytical.ErrorSchemaID != "cartulary.network_flow.import_owner_error.v1" {
		t.Fatalf("unexpected analytical registry row: %#v", analytical)
	}
}

func TestDeriveImportTargetArtifactsProducesIntegrityBoundProjections(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	first, err := deriveImportTargetArtifacts(root)
	if err != nil {
		t.Fatalf("derive import-target artifacts: %v", err)
	}
	second, err := deriveImportTargetArtifacts(root)
	if err != nil {
		t.Fatalf("derive import-target artifacts again: %v", err)
	}
	if len(first) != 5 || len(second) != len(first) {
		t.Fatalf("unexpected generated artifact counts %d and %d", len(first), len(second))
	}
	byPath := map[string]artifact{}
	for index := range first {
		if first[index] != second[index] {
			t.Fatalf("generated artifact %d is not deterministic", index)
		}
		byPath[first[index].Path] = first[index]
	}
	for _, path := range []string{
		importTargetSourceManifestPath,
		importTargetRegistryPath,
		importTargetAdapterDescriptorPath,
		importTargetVerificationPath,
		importTargetIntegrityPath,
	} {
		if _, exists := byPath[path]; !exists {
			t.Fatalf("missing generated import-target artifact %s", path)
		}
	}
	var integrity map[string]any
	if err := json.Unmarshal([]byte(byPath[importTargetIntegrityPath].JSON), &integrity); err != nil {
		t.Fatalf("decode integrity projection: %v", err)
	}
	if integrity["registry_sha256"] != byPath[importTargetRegistryPath].SHA256 ||
		integrity["adapter_descriptors_sha256"] != byPath[importTargetAdapterDescriptorPath].SHA256 ||
		integrity["verification_targets_sha256"] != byPath[importTargetVerificationPath].SHA256 {
		t.Fatalf("integrity projection does not bind every generated semantic projection: %#v", integrity)
	}
}

func TestImportTargetCatalogValidationRejectsAmbiguity(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	input, err := readDeclaredImportTargetInput[importTargetInputCatalog](
		root,
		"contracts/imports/index.json",
	)
	if err != nil {
		t.Fatalf("read catalog: %v", err)
	}

	t.Run("duplicate target selector", func(t *testing.T) {
		catalog := input.Value
		catalog.TargetOrder = append([]string(nil), catalog.TargetOrder...)
		catalog.TargetOrder[1] = catalog.TargetOrder[0]
		requireErrorContains(
			t,
			validateImportTargetCatalog(catalog),
			"target_order contains duplicate",
		)
	})

	t.Run("unknown facade owner", func(t *testing.T) {
		catalog := input.Value
		catalog.FacadeContracts = append(
			[]importTargetFacadeContract(nil),
			catalog.FacadeContracts...,
		)
		catalog.FacadeContracts[0].OwnerContractRef = "module.unknown@1"
		requireErrorContains(
			t,
			validateImportTargetCatalog(catalog),
			"references unknown owner",
		)
	})
}

func TestImportTargetAvailabilityValidationFailsClosed(t *testing.T) {
	t.Run("enabled target missing binding", func(t *testing.T) {
		requireErrorContains(t, validateImportViewAvailability(importViewTargetInput{
			TargetViewSchemaID:          "cartulary.view.test.v1",
			OwnerContractRef:            "module.test@1",
			SourceResourceFamily:        "test",
			AvailabilityKind:            "enabled",
			ActivationPolicy:            "always",
			MappingContractSchemaID:     "cartulary.imports.approved_view_mapping.v1",
			DefaultUnknownColumnPolicy:  "reject_if_unmapped",
			EntityBearingDefault:        "none",
			PublicProjectionDisposition: "selectable",
		}), "enabled target requires")
	})

	t.Run("reserved target projected selectable", func(t *testing.T) {
		requireErrorContains(t, validateImportViewAvailability(importViewTargetInput{
			TargetViewSchemaID:          "cartulary.view.test.v1",
			OwnerContractRef:            "module.test@1",
			SourceResourceFamily:        "test",
			AvailabilityKind:            "reserved",
			ActivationPolicy:            "unavailable_reserved",
			MappingContractSchemaID:     "cartulary.imports.approved_view_mapping.v1",
			DefaultUnknownColumnPolicy:  "reject_if_unmapped",
			EntityBearingDefault:        "none",
			PublicProjectionDisposition: "selectable",
		}), "reserved target must be")
	})

	t.Run("claim target without claim policy", func(t *testing.T) {
		requireErrorContains(t, validateAnalyticalAvailability(importTargetAnalyticalInput{
			AvailabilityKind:            "claim_gated",
			ActivationPolicy:            "always",
			DefaultUnknownColumnPolicy:  "target_owned",
			EntityBearingDefault:        "target_owned",
			PublicProjectionDisposition: "extension_claim_gated",
		}), "must require a claim")
	})
}

func TestResolveImportTargetContributionRejectsMissingAndDuplicateBindings(t *testing.T) {
	base := map[string]any{
		"facts": []any{
			map[string]any{
				"fact_kind":  "contribution",
				"profile_id": "network_flow_activity",
				"contribution": map[string]any{
					"kind":              "import_target",
					"target_kind":       "network_flow_table",
					"facade_binding_id": "network_flow_activity.import_facade.v1",
				},
			},
		},
	}
	bindingID, err := resolveImportTargetContribution(
		base,
		"network_flow_table",
		"network_flow_activity",
	)
	if err != nil || bindingID != "network_flow_activity.import_facade.v1" {
		t.Fatalf("resolve valid contribution: id=%s err=%v", bindingID, err)
	}

	missing := cloneJSONMap(t, base)
	delete(
		missing["facts"].([]any)[0].(map[string]any)["contribution"].(map[string]any),
		"facade_binding_id",
	)
	requireErrorContains(
		t,
		func() error {
			_, err := resolveImportTargetContribution(
				missing,
				"network_flow_table",
				"network_flow_activity",
			)
			return err
		}(),
		"invalid facade_binding_id",
	)

	duplicate := cloneJSONMap(t, base)
	facts := duplicate["facts"].([]any)
	duplicate["facts"] = append(facts, cloneJSONMap(t, facts[0].(map[string]any)))
	requireErrorContains(
		t,
		func() error {
			_, err := resolveImportTargetContribution(
				duplicate,
				"network_flow_table",
				"network_flow_activity",
			)
			return err
		}(),
		"expected exactly one",
	)
}

func TestImportTargetRowDigestExcludesOnlyDigestMember(t *testing.T) {
	targetID := "view_schema:cartulary.view.test.v1"
	viewID := "cartulary.view.test.v1"
	row := importTargetRegistryRow{
		TargetID:                    targetID,
		TargetKind:                  "view_schema",
		TargetViewSchemaID:          &viewID,
		OwnerContractRef:            "module.test@1",
		SourceResourceFamily:        "test",
		FacadeKind:                  "owner_create",
		AvailabilityKind:            "reserved",
		ActivationPolicy:            "unavailable_reserved",
		MappingContractSchemaID:     "cartulary.imports.approved_view_mapping.v1",
		ErrorSchemaID:               "cartulary.imports.owner_create_error.v1",
		CommitProtocolID:            "cartulary.imports.unit_commit.v1",
		DefaultUnknownColumnPolicy:  "reject_if_unmapped",
		EntityBearingDefault:        "none",
		PublicProjectionDisposition: "hidden_reserved",
	}
	first, err := importTargetRowDigest(row)
	if err != nil {
		t.Fatalf("digest row: %v", err)
	}
	row.RowSHA256 = strings.Repeat("f", 64)
	replay, err := importTargetRowDigest(row)
	if err != nil {
		t.Fatalf("digest replay row: %v", err)
	}
	if replay != first {
		t.Fatal("row digest included its own digest member")
	}
	row.OwnerContractRef = "module.changed@1"
	changed, err := importTargetRowDigest(row)
	if err != nil {
		t.Fatalf("digest changed row: %v", err)
	}
	if changed == first {
		t.Fatal("row digest did not bind a semantic member")
	}
}
