package workbook

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestWorkbookContributionCatalogValidatesEveryActiveQuerySurface_Unit(t *testing.T) {
	if _, err := NewQueryProvider(nil); err == nil {
		t.Fatal("query provider accepted a nil query function")
	}
	descriptors, queryContributions, createContributions, patchContributions, conflictContributions := validCatalogInputs(t)
	actions := validActionContributions()
	catalog, err := NewWorkbookContributionCatalog(catalogInput(
		mustDescriptorSet(t, descriptors),
		queryContributions,
		createContributions,
		patchContributions,
		conflictContributions,
		validActionRequirements(),
		actions,
	))
	if err != nil {
		t.Fatalf("construct valid contribution catalog: %v", err)
	}

	resources := viewschema.ListPublicResources()
	wantIDs := make([]string, 0, len(resources))
	for _, resource := range resources {
		wantIDs = append(wantIDs, resource.ViewSchemaID)
		if provider, ok := catalog.QueryFor(resource.ViewSchemaID); !ok || provider == nil {
			t.Fatalf("active surface %s did not resolve exactly once", resource.ViewSchemaID)
		}
	}
	if got := catalog.QuerySurfaceIDs(); !reflect.DeepEqual(got, wantIDs) {
		t.Fatalf("catalog surface IDs:\ngot  %v\nwant %v", got, wantIDs)
	}

	tests := []struct {
		name string
		edit func([]providercontract.ProviderDescriptor, []QueryContribution) ([]providercontract.ProviderDescriptor, []QueryContribution)
		want string
	}{
		{
			name: "duplicate contribution",
			edit: func(descriptors []providercontract.ProviderDescriptor, contributions []QueryContribution) ([]providercontract.ProviderDescriptor, []QueryContribution) {
				return descriptors, append(contributions, contributions[0])
			},
			want: "duplicate workbook query contribution",
		},
		{
			name: "missing contribution",
			edit: func(descriptors []providercontract.ProviderDescriptor, contributions []QueryContribution) ([]providercontract.ProviderDescriptor, []QueryContribution) {
				return descriptors, contributions[1:]
			},
			want: "missing active surface",
		},
		{
			name: "unknown contribution",
			edit: func(descriptors []providercontract.ProviderDescriptor, contributions []QueryContribution) ([]providercontract.ProviderDescriptor, []QueryContribution) {
				contributions[0].ViewSchemaID = "cartulary.view.unknown.v1"
				return descriptors, contributions
			},
			want: "unknown active surface",
		},
		{
			name: "owner mismatch",
			edit: func(descriptors []providercontract.ProviderDescriptor, contributions []QueryContribution) ([]providercontract.ProviderDescriptor, []QueryContribution) {
				contributions[0].SourceOwnerKey = "wrong-owner"
				return descriptors, contributions
			},
			want: "does not match descriptor owner",
		},
		{
			name: "record type mismatch",
			edit: func(descriptors []providercontract.ProviderDescriptor, contributions []QueryContribution) ([]providercontract.ProviderDescriptor, []QueryContribution) {
				contributions[0].SourceRecordTypes = []string{"wrong_record"}
				return descriptors, contributions
			},
			want: "do not match active schema",
		},
		{
			name: "backend capability mismatch",
			edit: func(descriptors []providercontract.ProviderDescriptor, contributions []QueryContribution) ([]providercontract.ProviderDescriptor, []QueryContribution) {
				contributions[0].BackendKind = QueryBackendSourceOwner
				return descriptors, contributions
			},
			want: "does not match descriptor capability backend",
		},
		{
			name: "nil provider",
			edit: func(descriptors []providercontract.ProviderDescriptor, contributions []QueryContribution) ([]providercontract.ProviderDescriptor, []QueryContribution) {
				contributions[0].Provider = nil
				return descriptors, contributions
			},
			want: "has nil provider",
		},
		{
			name: "typed nil provider",
			edit: func(descriptors []providercontract.ProviderDescriptor, contributions []QueryContribution) ([]providercontract.ProviderDescriptor, []QueryContribution) {
				var provider *queryProvider
				contributions[0].Provider = provider
				return descriptors, contributions
			},
			want: "has nil provider",
		},
		{
			name: "descriptor record type mismatch",
			edit: func(descriptors []providercontract.ProviderDescriptor, contributions []QueryContribution) ([]providercontract.ProviderDescriptor, []QueryContribution) {
				descriptors[0].SourceRecordTypes = []string{"wrong_record"}
				return descriptors, contributions
			},
			want: "do not match projection descriptor",
		},
		{
			name: "duplicate descriptor view",
			edit: func(descriptors []providercontract.ProviderDescriptor, contributions []QueryContribution) ([]providercontract.ProviderDescriptor, []QueryContribution) {
				duplicate := descriptors[0]
				duplicate.ProviderID += "-duplicate"
				return append(descriptors, duplicate), contributions
			},
			want: "duplicate active projection descriptor ownership",
		},
		{
			name: "missing descriptor",
			edit: func(descriptors []providercontract.ProviderDescriptor, contributions []QueryContribution) ([]providercontract.ProviderDescriptor, []QueryContribution) {
				return descriptors[1:], contributions
			},
			want: "has no active projection descriptor",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testDescriptors := cloneProviderDescriptors(descriptors)
			testContributions := cloneQueryContributions(queryContributions)
			testDescriptors, testContributions = test.edit(testDescriptors, testContributions)
			_, err := NewWorkbookContributionCatalog(catalogInput(
				mustDescriptorSet(t, testDescriptors),
				testContributions,
				createContributions,
				patchContributions,
				conflictContributions,
				validActionRequirements(),
				actions,
			))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}
}

func TestWorkbookContributionCatalogValidatesCreateAndPatchRequirements_Unit(t *testing.T) {
	descriptors, queries, creates, patches, conflicts := validCatalogInputs(t)
	actions := validActionContributions()
	catalog, err := NewWorkbookContributionCatalog(catalogInput(
		mustDescriptorSet(t, descriptors),
		queries,
		creates,
		patches,
		conflicts,
		validActionRequirements(),
		actions,
	))
	if err != nil {
		t.Fatalf("construct valid contribution catalog: %v", err)
	}
	for _, resource := range viewschema.ListPublicResources() {
		schema, _ := viewschema.Lookup(resource.ViewSchemaID)
		provider, registered := catalog.CreateFor(resource.ViewSchemaID)
		if registered != schema.CreateCapable || registered && provider == nil {
			t.Fatalf(
				"create lookup %s: registered=%t provider_nil=%t create_capable=%t",
				resource.ViewSchemaID,
				registered,
				provider == nil,
				schema.CreateCapable,
			)
		}
	}
	for recordType := range expectedPatchSurfaces() {
		if provider, registered := catalog.PatchFor(recordType); !registered || provider == nil {
			t.Fatalf("writable record type %s has no patch provider", recordType)
		}
	}
	for recordType := range expectedConflictSurfaces() {
		if provider, registered := catalog.ConflictFor(recordType); !registered || provider == nil {
			t.Fatalf("conflict-capable record type %s has no conflict provider", recordType)
		}
	}
	for _, viewSchemaID := range []string{
		"cartulary.view.hosts.v1",
		"cartulary.view.identities.v1",
		"cartulary.view.timeline.v2",
	} {
		if provider, registered := catalog.ClipboardFor(viewSchemaID); !registered || provider == nil {
			t.Fatalf("clipboard surface %s has no provider", viewSchemaID)
		}
	}
	if provider, registered := catalog.BulkFor("cartulary.view.timeline.v2"); !registered || provider == nil {
		t.Fatal("Timeline bulk surface has no provider")
	}
	for _, recordType := range []string{"evidence", "host", "identity", "timeline_event"} {
		if provider, registered := catalog.LinkedNoteFor(recordType); !registered || provider == nil {
			t.Fatalf("linked-note record type %s has no provider", recordType)
		}
	}
	for _, recordType := range []string{"decision", "timeline_event"} {
		if provider, registered := catalog.SupersedeFor(recordType); !registered || provider == nil {
			t.Fatalf("supersede record type %s has no provider", recordType)
		}
	}

	createTests := []struct {
		name string
		edit func([]CreateContribution) []CreateContribution
		want string
	}{
		{
			name: "duplicate",
			edit: func(input []CreateContribution) []CreateContribution {
				return append(input, input[0])
			},
			want: "duplicate workbook create contribution",
		},
		{
			name: "missing",
			edit: func(input []CreateContribution) []CreateContribution {
				return input[1:]
			},
			want: "missing create-capable surface",
		},
		{
			name: "unknown",
			edit: func(input []CreateContribution) []CreateContribution {
				input[0].ViewSchemaID = "cartulary.view.unknown.v1"
				return input
			},
			want: "unknown active surface",
		},
		{
			name: "owner mismatch",
			edit: func(input []CreateContribution) []CreateContribution {
				input[0].SourceOwnerKey = "wrong-owner"
				return input
			},
			want: "does not match descriptor owner",
		},
		{
			name: "record type mismatch",
			edit: func(input []CreateContribution) []CreateContribution {
				input[0].SourceRecordTypes = []string{"wrong_record"}
				return input
			},
			want: "do not match active schema",
		},
		{
			name: "nil provider",
			edit: func(input []CreateContribution) []CreateContribution {
				input[0].Provider = nil
				return input
			},
			want: "has nil provider",
		},
		{
			name: "typed nil provider",
			edit: func(input []CreateContribution) []CreateContribution {
				var provider *neutralCreateProvider
				input[0].Provider = provider
				return input
			},
			want: "has nil provider",
		},
	}
	for _, test := range createTests {
		t.Run("create/"+test.name, func(t *testing.T) {
			_, err := NewWorkbookContributionCatalog(catalogInput(
				mustDescriptorSet(t, descriptors),
				queries,
				test.edit(cloneCreateContributions(creates)),
				patches,
				conflicts,
				validActionRequirements(),
				actions,
			))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}

	patchTests := []struct {
		name string
		edit func([]PatchContribution) []PatchContribution
		want string
	}{
		{
			name: "duplicate",
			edit: func(input []PatchContribution) []PatchContribution {
				return append(input, input[0])
			},
			want: "duplicate workbook patch contribution",
		},
		{
			name: "missing",
			edit: func(input []PatchContribution) []PatchContribution {
				return input[1:]
			},
			want: "missing writable record type",
		},
		{
			name: "unknown",
			edit: func(input []PatchContribution) []PatchContribution {
				input[0].RecordType = "unknown_record"
				return input
			},
			want: "non-writable record type",
		},
		{
			name: "surface mismatch",
			edit: func(input []PatchContribution) []PatchContribution {
				input[0].ViewSchemaIDs = []string{"cartulary.view.unknown.v1"}
				return input
			},
			want: "do not match active writable surfaces",
		},
		{
			name: "nil provider",
			edit: func(input []PatchContribution) []PatchContribution {
				input[0].Provider = nil
				return input
			},
			want: "has nil provider",
		},
		{
			name: "typed nil provider",
			edit: func(input []PatchContribution) []PatchContribution {
				var provider *neutralPatchProvider
				input[0].Provider = provider
				return input
			},
			want: "has nil provider",
		},
	}
	for _, test := range patchTests {
		t.Run("patch/"+test.name, func(t *testing.T) {
			_, err := NewWorkbookContributionCatalog(catalogInput(
				mustDescriptorSet(t, descriptors),
				queries,
				creates,
				test.edit(clonePatchContributions(patches)),
				conflicts,
				validActionRequirements(),
				actions,
			))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}

	conflictTests := []struct {
		name string
		edit func([]ConflictContribution) []ConflictContribution
		want string
	}{
		{
			name: "duplicate",
			edit: func(input []ConflictContribution) []ConflictContribution {
				return append(input, input[0])
			},
			want: "duplicate workbook conflict contribution",
		},
		{
			name: "missing",
			edit: func(input []ConflictContribution) []ConflictContribution {
				return input[1:]
			},
			want: "missing conflict-capable record type",
		},
		{
			name: "unknown",
			edit: func(input []ConflictContribution) []ConflictContribution {
				input[0].RecordType = "unknown_record"
				return input
			},
			want: "non-conflict-capable record type",
		},
		{
			name: "surface mismatch",
			edit: func(input []ConflictContribution) []ConflictContribution {
				input[0].ViewSchemaIDs = []string{"cartulary.view.unknown.v1"}
				return input
			},
			want: "do not match active conflict-capable surfaces",
		},
		{
			name: "nil provider",
			edit: func(input []ConflictContribution) []ConflictContribution {
				input[0].Provider = nil
				return input
			},
			want: "has nil provider",
		},
		{
			name: "typed nil provider",
			edit: func(input []ConflictContribution) []ConflictContribution {
				var provider *neutralConflictProvider
				input[0].Provider = provider
				return input
			},
			want: "has nil provider",
		},
	}
	for _, test := range conflictTests {
		t.Run("conflict/"+test.name, func(t *testing.T) {
			_, err := NewWorkbookContributionCatalog(catalogInput(
				mustDescriptorSet(t, descriptors),
				queries,
				creates,
				patches,
				test.edit(cloneConflictContributions(conflicts)),
				validActionRequirements(),
				actions,
			))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}

	constructActions := func(actionContributions MutationActionContributions) error {
		_, err := NewWorkbookContributionCatalog(catalogInput(
			mustDescriptorSet(t, descriptors),
			queries,
			creates,
			patches,
			conflicts,
			validActionRequirements(),
			actionContributions,
		))
		return err
	}
	actionTests := []struct {
		name string
		edit func(*MutationActionContributions)
		want string
	}{
		{
			name: "clipboard missing",
			edit: func(input *MutationActionContributions) { input.Clipboard = input.Clipboard[1:] },
			want: "clipboard contribution missing required surface",
		},
		{
			name: "clipboard duplicate",
			edit: func(input *MutationActionContributions) {
				input.Clipboard = append(input.Clipboard, input.Clipboard[0])
			},
			want: "duplicate workbook clipboard contribution",
		},
		{
			name: "clipboard cross-surface",
			edit: func(input *MutationActionContributions) {
				input.Clipboard[0].ViewSchemaID = "cartulary.view.evidence.v1"
			},
			want: "clipboard contribution references unsupported surface",
		},
		{
			name: "clipboard typed nil",
			edit: func(input *MutationActionContributions) {
				var provider *clipboardProvider
				input.Clipboard[0].Provider = provider
			},
			want: "clipboard contribution \"cartulary.view.hosts.v1\" has nil provider",
		},
		{
			name: "bulk missing",
			edit: func(input *MutationActionContributions) { input.Bulk = nil },
			want: "bulk contribution missing required surface",
		},
		{
			name: "bulk incomplete provider",
			edit: func(input *MutationActionContributions) { input.Bulk[0].Provider = &bulkProvider{} },
			want: "bulk contribution \"cartulary.view.timeline.v2\" is invalid",
		},
		{
			name: "linked-note missing",
			edit: func(input *MutationActionContributions) { input.LinkedNote = input.LinkedNote[1:] },
			want: "linked-note contribution missing required record type",
		},
		{
			name: "linked-note unknown record type",
			edit: func(input *MutationActionContributions) { input.LinkedNote[0].RecordType = "assessment" },
			want: "linked-note contribution references unsupported record type",
		},
		{
			name: "supersede missing",
			edit: func(input *MutationActionContributions) { input.Supersede = input.Supersede[1:] },
			want: "supersede contribution missing required record type",
		},
		{
			name: "supersede unknown record type",
			edit: func(input *MutationActionContributions) { input.Supersede[0].RecordType = "task_request" },
			want: "supersede contribution references unsupported record type",
		},
	}
	for _, test := range actionTests {
		t.Run("action/"+test.name, func(t *testing.T) {
			actions := validActionContributions()
			test.edit(&actions)
			err := constructActions(actions)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}

	requirementTests := []struct {
		name string
		edit func(*ActionCapabilityRequirements)
		want string
	}{
		{
			name: "duplicate clipboard requirement",
			edit: func(input *ActionCapabilityRequirements) {
				input.ClipboardViewSchemaIDs = append(input.ClipboardViewSchemaIDs, input.ClipboardViewSchemaIDs[0])
			},
			want: "duplicate workbook clipboard capability requirement",
		},
		{
			name: "unknown bulk surface",
			edit: func(input *ActionCapabilityRequirements) {
				input.BulkViewSchemaIDs = []string{"cartulary.view.unknown.v1"}
			},
			want: "bulk capability requirement references unknown surface",
		},
		{
			name: "empty linked-note key",
			edit: func(input *ActionCapabilityRequirements) {
				input.LinkedNoteRecordTypes = append(input.LinkedNoteRecordTypes, "")
			},
			want: "linked-note capability requirement has empty key",
		},
	}
	for _, test := range requirementTests {
		t.Run("requirements/"+test.name, func(t *testing.T) {
			requirements := validActionRequirements()
			test.edit(&requirements)
			_, err := NewWorkbookContributionCatalog(catalogInput(
				mustDescriptorSet(t, descriptors), queries, creates, patches, conflicts,
				requirements, validActionContributions(),
			))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}

	t.Run("catalog input collections are copied", func(t *testing.T) {
		requirements := validActionRequirements()
		actions := validActionContributions()
		input := catalogInput(
			mustDescriptorSet(t, descriptors),
			cloneQueryContributions(queries),
			cloneCreateContributions(creates),
			clonePatchContributions(patches),
			cloneConflictContributions(conflicts),
			requirements, actions,
		)
		queryID := input.Queries[0].ViewSchemaID
		createID := input.Creates[0].ViewSchemaID
		patchType := input.Patches[0].RecordType
		conflictType := input.Conflicts[0].RecordType
		clipboardID := input.Actions.Clipboard[0].ViewSchemaID
		catalog, err := NewWorkbookContributionCatalog(input)
		if err != nil {
			t.Fatalf("construct catalog: %v", err)
		}
		input.Queries[0].ViewSchemaID = "cartulary.view.unknown.v1"
		input.Queries[0].SourceRecordTypes[0] = "unknown_record"
		input.Creates[0].ViewSchemaID = "cartulary.view.unknown.v1"
		input.Creates[0].SourceRecordTypes[0] = "unknown_record"
		input.Patches[0].RecordType = "unknown_record"
		input.Patches[0].ViewSchemaIDs[0] = "cartulary.view.unknown.v1"
		input.Conflicts[0].RecordType = "unknown_record"
		input.Conflicts[0].ViewSchemaIDs[0] = "cartulary.view.unknown.v1"
		input.ActionRequirements.ClipboardViewSchemaIDs[0] = "cartulary.view.unknown.v1"
		input.Actions.Clipboard[0].ViewSchemaID = "cartulary.view.unknown.v1"
		if _, ok := catalog.QueryFor(queryID); !ok {
			t.Fatal("catalog retained caller-owned query input")
		}
		if _, ok := catalog.CreateFor(createID); !ok {
			t.Fatal("catalog retained caller-owned create input")
		}
		if _, ok := catalog.PatchFor(patchType); !ok {
			t.Fatal("catalog retained caller-owned patch input")
		}
		if _, ok := catalog.ConflictFor(conflictType); !ok {
			t.Fatal("catalog retained caller-owned conflict input")
		}
		if _, ok := catalog.ClipboardFor(clipboardID); !ok {
			t.Fatal("catalog retained caller-owned action input")
		}
	})

	t.Run("neutral mutation providers reject incomplete construction", func(t *testing.T) {
		createDecode := func(io.Reader) (CreateAdmission, *MutationFailure, error) { return nil, nil, nil }
		createExecute := func(context.Context, CreateCommand) (MutationOutcome, error) {
			return SuccessfulRowMutation(MutationResult{StatusCode: 200, Payload: map[string]any{"row": map[string]any{}}}), nil
		}
		if _, err := NewCreateProvider(nil, createExecute); err == nil {
			t.Fatal("create provider accepted a nil decoder")
		}
		if _, err := NewCreateProvider(createDecode, nil); err == nil {
			t.Fatal("create provider accepted a nil executor")
		}

		patchDecode := func(io.Reader) (PatchAdmission, *MutationFailure, error) { return nil, nil, nil }
		patchExecute := func(context.Context, PatchCommand) (MutationOutcome, error) {
			return SuccessfulRowMutation(MutationResult{StatusCode: 200, Payload: map[string]any{"row": map[string]any{}}}), nil
		}
		if _, err := NewPatchProvider(nil, patchExecute); err == nil {
			t.Fatal("patch provider accepted a nil decoder")
		}
		if _, err := NewPatchProvider(patchDecode, nil); err == nil {
			t.Fatal("patch provider accepted a nil executor")
		}

		conflictDecode := func(io.Reader, string, ConflictClaims) (ConflictAdmission, *MutationFailure, error) {
			return nil, nil, nil
		}
		conflictExecute := func(context.Context, ConflictCommand) (MutationOutcome, error) {
			return SuccessfulRowMutation(MutationResult{StatusCode: 200, Payload: map[string]any{"row": map[string]any{}}}), nil
		}
		if _, err := NewConflictProvider(nil, conflictExecute); err == nil {
			t.Fatal("conflict provider accepted a nil decoder")
		}
		if _, err := NewConflictProvider(conflictDecode, nil); err == nil {
			t.Fatal("conflict provider accepted a nil executor")
		}

		clipboardDecode := func(io.Reader) (ClipboardAdmission, *MutationFailure, error) { return nil, nil, nil }
		clipboardExecute := func(context.Context, ClipboardCommand) (MutationOutcome, error) {
			return SuccessfulRowMutation(MutationResult{StatusCode: 200, Payload: map[string]any{"row": map[string]any{}}}), nil
		}
		if _, err := NewClipboardProvider(nil, clipboardExecute); err == nil {
			t.Fatal("clipboard provider accepted a nil decoder")
		}
		if _, err := NewClipboardProvider(clipboardDecode, nil); err == nil {
			t.Fatal("clipboard provider accepted a nil executor")
		}

		bulkDecode := func(io.Reader) (BulkAdmission, *MutationFailure, error) { return nil, nil, nil }
		bulkExecute := func(context.Context, BulkCommand) (MutationOutcome, error) {
			return SuccessfulRowMutation(MutationResult{StatusCode: 200, Payload: map[string]any{"row": map[string]any{}}}), nil
		}
		if _, err := NewBulkProvider(nil, bulkExecute); err == nil {
			t.Fatal("bulk provider accepted a nil decoder")
		}
		if _, err := NewBulkProvider(bulkDecode, nil); err == nil {
			t.Fatal("bulk provider accepted a nil executor")
		}

		linkedNoteDecode := func(io.Reader) (LinkedNoteAdmission, *MutationFailure, error) { return nil, nil, nil }
		linkedNoteExecute := func(context.Context, LinkedNoteCommand) (MutationOutcome, error) {
			return SuccessfulRowMutation(MutationResult{StatusCode: 200, Payload: map[string]any{"row": map[string]any{}}}), nil
		}
		if _, err := NewLinkedNoteProvider(nil, linkedNoteExecute); err == nil {
			t.Fatal("linked-note provider accepted a nil decoder")
		}
		if _, err := NewLinkedNoteProvider(linkedNoteDecode, nil); err == nil {
			t.Fatal("linked-note provider accepted a nil executor")
		}

		supersedeDecode := func(io.Reader) (SupersedeAdmission, *MutationFailure, error) { return nil, nil, nil }
		supersedeExecute := func(context.Context, SupersedeCommand) (MutationOutcome, error) {
			return SuccessfulRowMutation(MutationResult{StatusCode: 200, Payload: map[string]any{"row": map[string]any{}}}), nil
		}
		if _, err := NewSupersedeProvider(nil, supersedeExecute); err == nil {
			t.Fatal("supersede provider accepted a nil decoder")
		}
		if _, err := NewSupersedeProvider(supersedeDecode, nil); err == nil {
			t.Fatal("supersede provider accepted a nil executor")
		}

		invalidCreates := cloneCreateContributions(creates)
		invalidCreates[0].Provider = &neutralCreateProvider{}
		if _, err := NewWorkbookContributionCatalog(catalogInput(
			mustDescriptorSet(t, descriptors), queries, invalidCreates, patches, conflicts,
			validActionRequirements(), validActionContributions(),
		)); err == nil || !strings.Contains(err.Error(), "create contribution") {
			t.Fatalf("catalog accepted incomplete create provider: %v", err)
		}
		invalidPatches := clonePatchContributions(patches)
		invalidPatches[0].Provider = &neutralPatchProvider{}
		if _, err := NewWorkbookContributionCatalog(catalogInput(
			mustDescriptorSet(t, descriptors), queries, creates, invalidPatches, conflicts,
			validActionRequirements(), validActionContributions(),
		)); err == nil || !strings.Contains(err.Error(), "patch contribution") {
			t.Fatalf("catalog accepted incomplete patch provider: %v", err)
		}
		invalidConflicts := cloneConflictContributions(conflicts)
		invalidConflicts[0].Provider = &neutralConflictProvider{}
		if _, err := NewWorkbookContributionCatalog(catalogInput(
			mustDescriptorSet(t, descriptors), queries, creates, patches, invalidConflicts,
			validActionRequirements(), validActionContributions(),
		)); err == nil || !strings.Contains(err.Error(), "conflict contribution") {
			t.Fatalf("catalog accepted incomplete conflict provider: %v", err)
		}
	})

	t.Run("mutation outcome is exclusive and details are copied", func(t *testing.T) {
		if err := (MutationOutcome{}).Validate(); err == nil {
			t.Fatal("empty mutation outcome was accepted")
		}
		if err := (MutationOutcome{result: &MutationResult{}, failure: IncidentClosedFailure()}).Validate(); err == nil {
			t.Fatal("mutation outcome containing both result and failure was accepted")
		}
		if err := SuccessfulRowMutation(MutationResult{StatusCode: 200, Payload: map[string]any{"row": map[string]any{}}}).Validate(); err != nil {
			t.Fatalf("successful mutation outcome: %v", err)
		}
		if err := SuccessfulRowMutation(MutationResult{
			StatusCode: 200,
			Payload:    map[string]any{"row": map[string]any{}, "unsafe_owner_payload": "secret"},
		}).Validate(); err == nil {
			t.Fatal("mutation outcome accepted an arbitrary outer response field")
		}
		guards := []string{"first"}
		failure := IllegalTransitionFailure("open", "closed", "guard_failed", guards)
		guards[0] = "mutated"
		apiErr := mutationFailureAPIError(failure)
		if got := apiErr.Details["violated_guards"]; !reflect.DeepEqual(got, []string{"first"}) {
			t.Fatalf("failure retained caller-owned guards: %v", got)
		}
		apiErr = internalAPIError(errors.New("secret upstream detail"))
		if apiErr.Code != "internal_error" || apiErr.Message != "internal_error" || strings.Contains(apiErr.Message, "secret") {
			t.Fatalf("unexpected failure was not content-safe: %#v", apiErr)
		}
	})

	t.Run("safe failure vocabulary has stable public mappings", func(t *testing.T) {
		recordID := mustUUIDForCatalogTest(t, "00000000-0000-0000-0000-000000000123")
		sameFieldFailure, err := SameFieldConflictFailure(SameFieldConflictInput{
			ConflictToken: "opaque-token", RecordID: recordID, FieldKey: "timeline.summary",
			ConflictResolutionClass: SameFieldConflictTextCompareMerge,
			BaseRowVersion:          1, CurrentRowVersion: 2,
			ClientValue: "client", ServerValue: "server",
			BaseValue:       OptionalConflictValue{Present: true, Value: "base"},
			ServerUpdatedBy: recordID, ServerUpdatedAt: time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC),
		})
		if err != nil {
			t.Fatalf("same-field failure: %v", err)
		}
		conflictAPIError := mutationFailureAPIError(sameFieldFailure)
		conflictObject, ok := conflictAPIError.Conflict.(map[string]any)
		if !ok {
			t.Fatalf("same-field conflict has unexpected type: %#v", conflictAPIError.Conflict)
		}
		for _, key := range []string{
			"conflict_token", "record_id", "field_key", "conflict_resolution_class",
			"base_row_version", "current_row_version", "client_value", "server_value",
			"base_value", "server_updated_by", "server_updated_at",
		} {
			if _, ok := conflictObject[key]; !ok {
				t.Fatalf("same-field conflict omitted %q: %#v", key, conflictAPIError.Conflict)
			}
		}
		if _, ok := conflictObject["view_schema_id"]; ok {
			t.Fatalf("same-field conflict exposed non-contractual view_schema_id: %#v", conflictAPIError.Conflict)
		}
		tests := []struct {
			name      string
			failure   *MutationFailure
			status    int
			code      string
			retryable bool
		}{
			{name: "invalid payload", failure: InvalidPayloadFailure("field", "invalid_value"), status: 400, code: "invalid_mutation_payload"},
			{name: "client transaction", failure: ClientTxnConflictFailure("txn"), status: 409, code: "client_txn_conflict"},
			{name: "incident closed", failure: IncidentClosedFailure(), status: 409, code: "incident_closed"},
			{name: "concealed target", failure: TargetNotFoundFailure(), status: 404, code: "incident_not_found"},
			{name: "deleted", failure: RecordDeletedFailure(), status: 409, code: "record_deleted_use_restore"},
			{name: "row version", failure: RowVersionConflictFailure(recordID, 1, 2), status: 409, code: "row_version_conflict"},
			{name: "same field", failure: sameFieldFailure, status: 409, code: "same_field_conflict"},
			{name: "no change", failure: NoEffectiveChangeFailure("changes"), status: 400, code: "invalid_mutation_payload"},
			{name: "transition", failure: IllegalTransitionFailure("open", "closed", "guard", []string{"required"}), status: 409, code: "illegal_transition"},
			{name: "entity match", failure: EntityMatchConflictFailure("host", "hostname", []uuid.UUID{recordID}), status: 409, code: "entity_match_conflict"},
			{name: "Evidence attach", failure: EvidenceAttachFailure("blob_not_visible"), status: 409, code: "evidence_attach_rejected"},
			{name: "object invalid", failure: ObjectStoreInvalidFailure("invalid_request"), status: 500, code: "object_store_invalid_request"},
			{name: "object rejected", failure: ObjectStoreAccessRejectedFailure("credential_denied"), status: 503, code: "object_store_access_rejected"},
			{name: "object unavailable", failure: ObjectStoreUnavailableFailure("endpoint_unreachable"), status: 503, code: "object_store_unavailable", retryable: true},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				apiErr := mutationFailureAPIError(test.failure)
				if apiErr.Status != test.status || apiErr.Code != test.code || apiErr.Retryable != test.retryable {
					t.Fatalf("mapping = %#v, want status=%d code=%s retryable=%t", apiErr, test.status, test.code, test.retryable)
				}
			})
		}
	})

	t.Run("same-field conflict details are closed immutable and class-aware", func(t *testing.T) {
		recordID := mustUUIDForCatalogTest(t, "00000000-0000-0000-0000-000000000123")
		actorID := mustUUIDForCatalogTest(t, "00000000-0000-0000-0000-000000000456")
		updatedAt := time.Date(2026, 8, 20, 12, 0, 0, 123, time.FixedZone("source", -4*60*60))
		item := map[string]any{"item_ref": "tag:one", "item_kind": "tag", "display_text": "one"}
		collection := map[string]any{
			"kind": "collection_value_v1", "ordered": false, "items": []any{item},
		}
		failure, err := SameFieldConflictFailure(SameFieldConflictInput{
			ConflictToken: "opaque-token", RecordID: recordID, FieldKey: "timeline.tags",
			ConflictResolutionClass: SameFieldConflictCollectionReview,
			BaseRowVersion:          1, CurrentRowVersion: 2,
			ClientValue: collection, ServerValue: collection,
			BaseValue:       OptionalConflictValue{Present: true, Value: collection},
			ServerUpdatedBy: actorID, ServerUpdatedAt: updatedAt,
		})
		if err != nil {
			t.Fatalf("collection same-field failure: %v", err)
		}
		item["display_text"] = "mutated"
		collection["ordered"] = true
		apiErr := mutationFailureAPIError(failure)
		conflictObject := apiErr.Conflict.(map[string]any)
		clientValue := conflictObject["client_value"].(map[string]any)
		clientItems := clientValue["items"].([]any)
		if clientValue["ordered"] != false || clientItems[0].(map[string]any)["display_text"] != "one" {
			t.Fatalf("same-field failure retained caller-owned values: %#v", clientValue)
		}
		if conflictObject["server_updated_at"] != updatedAt.UTC().Format(time.RFC3339Nano) {
			t.Fatalf("same-field timestamp was not canonical UTC: %#v", apiErr.Conflict)
		}

		atomic, err := SameFieldConflictFailure(SameFieldConflictInput{
			ConflictToken: "opaque-token", RecordID: recordID, FieldKey: "artifact.status",
			ConflictResolutionClass: SameFieldConflictAtomicReplace,
			BaseRowVersion:          2, CurrentRowVersion: 3,
			ClientValue: "client", ServerValue: "server",
			ServerUpdatedBy: actorID, ServerUpdatedAt: updatedAt,
		})
		if err != nil {
			t.Fatalf("atomic same-field failure: %v", err)
		}
		atomicConflict := mutationFailureAPIError(atomic).Conflict.(map[string]any)
		if _, ok := atomicConflict["base_value"]; ok {
			t.Fatal("atomic same-field conflict invented an omitted base value")
		}

		validText := SameFieldConflictInput{
			ConflictToken: "opaque-token", RecordID: recordID, FieldKey: "timeline.summary",
			ConflictResolutionClass: SameFieldConflictTextCompareMerge,
			BaseRowVersion:          1, CurrentRowVersion: 2,
			ClientValue: "client", ServerValue: "server",
			BaseValue:       OptionalConflictValue{Present: true, Value: "base"},
			ServerUpdatedBy: actorID, ServerUpdatedAt: updatedAt,
			SuggestedMergedValue: OptionalConflictValue{Present: true, Value: nil},
		}
		withNullSuggestion, err := SameFieldConflictFailure(validText)
		if err != nil {
			t.Fatalf("text same-field failure: %v", err)
		}
		textConflict := mutationFailureAPIError(withNullSuggestion).Conflict.(map[string]any)
		if suggestion, ok := textConflict["suggested_merged_value"]; !ok || suggestion != nil {
			t.Fatalf("explicit null suggestion was not preserved: %#v", suggestion)
		}

		invalidInputs := []struct {
			name  string
			input SameFieldConflictInput
		}{
			{name: "empty token", input: func() SameFieldConflictInput { value := validText; value.ConflictToken = ""; return value }()},
			{name: "nil record", input: func() SameFieldConflictInput { value := validText; value.RecordID = uuid.Nil; return value }()},
			{name: "invalid window", input: func() SameFieldConflictInput { value := validText; value.CurrentRowVersion = 1; return value }()},
			{name: "nil actor", input: func() SameFieldConflictInput { value := validText; value.ServerUpdatedBy = uuid.Nil; return value }()},
			{name: "zero time", input: func() SameFieldConflictInput { value := validText; value.ServerUpdatedAt = time.Time{}; return value }()},
			{name: "unknown class", input: func() SameFieldConflictInput {
				value := validText
				value.ConflictResolutionClass = "future"
				return value
			}()},
			{name: "missing merge base", input: func() SameFieldConflictInput {
				value := validText
				value.BaseValue = OptionalConflictValue{}
				return value
			}()},
			{name: "non-text value", input: func() SameFieldConflictInput { value := validText; value.ClientValue = 7; return value }()},
			{name: "invalid JSON value", input: func() SameFieldConflictInput { value := validText; value.ClientValue = make(chan int); return value }()},
			{name: "suggestion on atomic", input: func() SameFieldConflictInput {
				value := validText
				value.ConflictResolutionClass = SameFieldConflictAtomicReplace
				return value
			}()},
			{name: "invalid collection", input: func() SameFieldConflictInput {
				value := validText
				value.ConflictResolutionClass = SameFieldConflictCollectionReview
				value.ClientValue = map[string]any{"kind": "collection_value_v1"}
				value.ServerValue = value.ClientValue
				value.BaseValue = OptionalConflictValue{Present: true, Value: value.ClientValue}
				value.SuggestedMergedValue = OptionalConflictValue{}
				return value
			}()},
		}
		for _, test := range invalidInputs {
			t.Run(test.name, func(t *testing.T) {
				if failure, err := SameFieldConflictFailure(test.input); err == nil || failure != nil {
					t.Fatalf("invalid same-field input accepted: failure=%#v err=%v", failure, err)
				}
			})
		}

		_, internalErr := resolveMutationOutcome(MutationOutcome{}, errors.New("secret malformed owner conflict"))
		if internalErr == nil || internalErr.Code != "internal_error" || internalErr.Message != "internal_error" {
			t.Fatalf("malformed owner conflict was not content-safe: %#v", internalErr)
		}
	})

	t.Run("decoder admissions and outcomes fail closed", func(t *testing.T) {
		var typedNil *catalogTestClipboardAdmission
		if apiErr := decodeMutationAPIError(typedNil, nil, nil); apiErr == nil || apiErr.Code != "internal_error" {
			t.Fatalf("typed-nil admission was not rejected: %#v", apiErr)
		}
		admission := &catalogTestClipboardAdmission{}
		if apiErr := decodeMutationAPIError(admission, InvalidPayloadFailure("field", "invalid_value"), nil); apiErr == nil || apiErr.Code != "internal_error" {
			t.Fatalf("admission plus failure was not rejected: %#v", apiErr)
		}
		if _, apiErr := resolveMutationOutcome(MutationOutcome{}, errors.New("secret provider failure")); apiErr == nil || apiErr.Message != "internal_error" {
			t.Fatalf("unexpected provider error was not content-safe: %#v", apiErr)
		}
	})
}

type catalogTestClipboardAdmission struct{}

func (*catalogTestClipboardAdmission) ClientTransactionID() string { return "txn" }
func (*catalogTestClipboardAdmission) AdmittedViewSchemaID() string {
	return "cartulary.view.timeline.v2"
}

func mustUUIDForCatalogTest(t testing.TB, raw string) uuid.UUID {
	t.Helper()
	value, err := uuid.Parse(raw)
	if err != nil {
		t.Fatalf("parse UUID: %v", err)
	}
	return value
}

func catalogInput(
	descriptors providercontract.DescriptorSet,
	queries []QueryContribution,
	creates []CreateContribution,
	patches []PatchContribution,
	conflicts []ConflictContribution,
	actionRequirements ActionCapabilityRequirements,
	actions MutationActionContributions,
) ContributionCatalogInput {
	return ContributionCatalogInput{
		ProjectionDescriptors: descriptors,
		Queries:               queries,
		Creates:               creates,
		Patches:               patches,
		Conflicts:             conflicts,
		ActionRequirements:    actionRequirements,
		Actions:               actions,
	}
}

func validCatalogInputs(t testing.TB) (
	[]providercontract.ProviderDescriptor,
	[]QueryContribution,
	[]CreateContribution,
	[]PatchContribution,
	[]ConflictContribution,
) {
	t.Helper()
	resources := viewschema.ListPublicResources()
	descriptors := make([]providercontract.ProviderDescriptor, 0, len(resources))
	contributions := make([]QueryContribution, 0, len(resources))
	creates := make([]CreateContribution, 0, len(resources))
	createProvider := neutralCreateProvider{
		decode: func(io.Reader) (CreateAdmission, *MutationFailure, error) {
			return nil, nil, nil
		},
		execute: func(context.Context, CreateCommand) (MutationOutcome, error) {
			return SuccessfulRowMutation(MutationResult{StatusCode: 200, Payload: map[string]any{"row": map[string]any{}}}), nil
		},
	}
	for _, resource := range resources {
		ownerKey := "owner:" + resource.ViewSchemaID
		descriptors = append(descriptors, providercontract.ProviderDescriptor{
			Status:            providercontract.ProviderStatusActive,
			ProviderID:        "provider:" + resource.ViewSchemaID,
			SourceOwnerModule: ownerKey,
			ViewSchemaIDs:     []string{resource.ViewSchemaID},
			SourceRecordTypes: append([]string(nil), resource.SourceRecordTypes...),
			Capabilities:      providercontract.ProviderCapabilities{Query: true},
		})
		queryProvider, err := NewQueryProvider(
			func(context.Context, QueryCommand) (querypage.Result, error) {
				return querypage.Result{}, nil
			},
		)
		if err != nil {
			t.Fatalf("construct query provider for %s: %v", resource.ViewSchemaID, err)
		}
		contributions = append(contributions, QueryContribution{
			ViewSchemaID:      resource.ViewSchemaID,
			SourceOwnerKey:    ownerKey,
			SourceRecordTypes: append([]string(nil), resource.SourceRecordTypes...),
			BackendKind:       QueryBackendProjection,
			Provider:          queryProvider,
		})
		schema, _ := viewschema.Lookup(resource.ViewSchemaID)
		if schema.CreateCapable {
			creates = append(creates, CreateContribution{
				ViewSchemaID:      resource.ViewSchemaID,
				SourceOwnerKey:    ownerKey,
				SourceRecordTypes: append([]string(nil), resource.SourceRecordTypes...),
				Provider:          &createProvider,
			})
		}
	}
	patchProvider := neutralPatchProvider{
		decode: func(io.Reader) (PatchAdmission, *MutationFailure, error) {
			return nil, nil, nil
		},
		execute: func(context.Context, PatchCommand) (MutationOutcome, error) {
			return SuccessfulRowMutation(MutationResult{StatusCode: 200, Payload: map[string]any{"row": map[string]any{}}}), nil
		},
	}
	conflictProvider := neutralConflictProvider{
		decode: func(
			io.Reader,
			string,
			ConflictClaims,
		) (ConflictAdmission, *MutationFailure, error) {
			return nil, nil, nil
		},
		execute: func(context.Context, ConflictCommand) (MutationOutcome, error) {
			return SuccessfulRowMutation(MutationResult{StatusCode: 200, Payload: map[string]any{"row": map[string]any{}}}), nil
		},
	}
	patches := make([]PatchContribution, 0, len(expectedPatchSurfaces()))
	for recordType, viewSchemaIDs := range expectedPatchSurfaces() {
		patches = append(patches, PatchContribution{
			RecordType:    recordType,
			ViewSchemaIDs: append([]string(nil), viewSchemaIDs...),
			Provider:      &patchProvider,
		})
	}
	conflicts := make([]ConflictContribution, 0, len(expectedConflictSurfaces()))
	for recordType, viewSchemaIDs := range expectedConflictSurfaces() {
		conflicts = append(conflicts, ConflictContribution{
			RecordType:    recordType,
			ViewSchemaIDs: append([]string(nil), viewSchemaIDs...),
			Provider:      &conflictProvider,
		})
	}
	return descriptors, contributions, creates, patches, conflicts
}

func validActionContributions() MutationActionContributions {
	batchResult := MutationResult{StatusCode: 200, Payload: map[string]any{
		"view_schema_id": "cartulary.view.timeline.v2", "rows": []any{},
	}}
	linkedNoteResult := MutationResult{StatusCode: 201, Payload: map[string]any{
		"row": map[string]any{}, "source_record_id": "00000000-0000-0000-0000-000000000001", "link_type": "contextual_note",
	}}
	supersedeResult := MutationResult{StatusCode: 200, Payload: map[string]any{"row": map[string]any{}}}
	clipboardProvider, err := NewClipboardProvider(
		func(io.Reader) (ClipboardAdmission, *MutationFailure, error) { return nil, nil, nil },
		func(context.Context, ClipboardCommand) (MutationOutcome, error) {
			return SuccessfulBatchMutation(batchResult), nil
		},
	)
	if err != nil {
		panic(err)
	}
	bulkProvider, err := NewBulkProvider(
		func(io.Reader) (BulkAdmission, *MutationFailure, error) { return nil, nil, nil },
		func(context.Context, BulkCommand) (MutationOutcome, error) {
			return SuccessfulBatchMutation(batchResult), nil
		},
	)
	if err != nil {
		panic(err)
	}
	linkedNoteProvider, err := NewLinkedNoteProvider(
		func(io.Reader) (LinkedNoteAdmission, *MutationFailure, error) { return nil, nil, nil },
		func(context.Context, LinkedNoteCommand) (MutationOutcome, error) {
			return SuccessfulLinkedNoteMutation(linkedNoteResult), nil
		},
	)
	if err != nil {
		panic(err)
	}
	supersedeProvider, err := NewSupersedeProvider(
		func(io.Reader) (SupersedeAdmission, *MutationFailure, error) { return nil, nil, nil },
		func(context.Context, SupersedeCommand) (MutationOutcome, error) {
			return SuccessfulSupersedeMutation(supersedeResult), nil
		},
	)
	if err != nil {
		panic(err)
	}
	return MutationActionContributions{
		Clipboard: []ClipboardContribution{
			{ViewSchemaID: "cartulary.view.hosts.v1", Provider: clipboardProvider},
			{ViewSchemaID: "cartulary.view.identities.v1", Provider: clipboardProvider},
			{ViewSchemaID: "cartulary.view.timeline.v2", Provider: clipboardProvider},
		},
		Bulk: []BulkContribution{
			{ViewSchemaID: "cartulary.view.timeline.v2", Provider: bulkProvider},
		},
		LinkedNote: []LinkedNoteContribution{
			{RecordType: "evidence", Provider: linkedNoteProvider},
			{RecordType: "host", Provider: linkedNoteProvider},
			{RecordType: "identity", Provider: linkedNoteProvider},
			{RecordType: "timeline_event", Provider: linkedNoteProvider},
		},
		Supersede: []SupersedeContribution{
			{RecordType: "decision", Provider: supersedeProvider},
			{RecordType: "timeline_event", Provider: supersedeProvider},
		},
	}
}

func validActionRequirements() ActionCapabilityRequirements {
	return ActionCapabilityRequirements{
		ClipboardViewSchemaIDs: []string{
			"cartulary.view.hosts.v1",
			"cartulary.view.identities.v1",
			"cartulary.view.timeline.v2",
		},
		BulkViewSchemaIDs:     []string{"cartulary.view.timeline.v2"},
		LinkedNoteRecordTypes: []string{"evidence", "host", "identity", "timeline_event"},
		SupersedeRecordTypes:  []string{"decision", "timeline_event"},
	}
}

func cloneProviderDescriptors(input []providercontract.ProviderDescriptor) []providercontract.ProviderDescriptor {
	cloned := append([]providercontract.ProviderDescriptor(nil), input...)
	for index := range cloned {
		cloned[index].ViewSchemaIDs = append([]string(nil), cloned[index].ViewSchemaIDs...)
		cloned[index].SourceRecordTypes = append([]string(nil), cloned[index].SourceRecordTypes...)
	}
	return cloned
}

func mustDescriptorSet(t testing.TB, descriptors []providercontract.ProviderDescriptor) providercontract.DescriptorSet {
	t.Helper()
	set, err := providercontract.NewDescriptorSet(descriptors)
	if err != nil {
		t.Fatalf("construct projection descriptor set: %v", err)
	}
	return set
}

func cloneQueryContributions(input []QueryContribution) []QueryContribution {
	cloned := append([]QueryContribution(nil), input...)
	for index := range cloned {
		cloned[index].SourceRecordTypes = append([]string(nil), cloned[index].SourceRecordTypes...)
	}
	return cloned
}

func cloneCreateContributions(input []CreateContribution) []CreateContribution {
	cloned := append([]CreateContribution(nil), input...)
	for index := range cloned {
		cloned[index].SourceRecordTypes = append([]string(nil), cloned[index].SourceRecordTypes...)
	}
	return cloned
}

func clonePatchContributions(input []PatchContribution) []PatchContribution {
	cloned := append([]PatchContribution(nil), input...)
	for index := range cloned {
		cloned[index].ViewSchemaIDs = append([]string(nil), cloned[index].ViewSchemaIDs...)
	}
	return cloned
}

func cloneConflictContributions(input []ConflictContribution) []ConflictContribution {
	cloned := append([]ConflictContribution(nil), input...)
	for index := range cloned {
		cloned[index].ViewSchemaIDs = append([]string(nil), cloned[index].ViewSchemaIDs...)
	}
	return cloned
}
