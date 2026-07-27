package workbook

import (
	"context"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestWorkbookContributionCatalogValidatesEveryActiveQuerySurface_Unit(t *testing.T) {
	descriptors, queryContributions, createContributions, patchContributions, conflictContributions := validCatalogInputs(t)
	catalog, err := NewWorkbookContributionCatalog(
		descriptors,
		queryContributions,
		createContributions,
		patchContributions,
		conflictContributions,
	)
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
		edit func([]projections.ProviderDescriptor, []QueryContribution) ([]projections.ProviderDescriptor, []QueryContribution)
		want string
	}{
		{
			name: "duplicate contribution",
			edit: func(descriptors []projections.ProviderDescriptor, contributions []QueryContribution) ([]projections.ProviderDescriptor, []QueryContribution) {
				return descriptors, append(contributions, contributions[0])
			},
			want: "duplicate workbook query contribution",
		},
		{
			name: "missing contribution",
			edit: func(descriptors []projections.ProviderDescriptor, contributions []QueryContribution) ([]projections.ProviderDescriptor, []QueryContribution) {
				return descriptors, contributions[1:]
			},
			want: "missing active surface",
		},
		{
			name: "unknown contribution",
			edit: func(descriptors []projections.ProviderDescriptor, contributions []QueryContribution) ([]projections.ProviderDescriptor, []QueryContribution) {
				contributions[0].ViewSchemaID = "cartulary.view.unknown.v1"
				return descriptors, contributions
			},
			want: "unknown active surface",
		},
		{
			name: "owner mismatch",
			edit: func(descriptors []projections.ProviderDescriptor, contributions []QueryContribution) ([]projections.ProviderDescriptor, []QueryContribution) {
				contributions[0].SourceOwnerKey = "wrong-owner"
				return descriptors, contributions
			},
			want: "does not match descriptor owner",
		},
		{
			name: "record type mismatch",
			edit: func(descriptors []projections.ProviderDescriptor, contributions []QueryContribution) ([]projections.ProviderDescriptor, []QueryContribution) {
				contributions[0].SourceRecordTypes = []string{"wrong_record"}
				return descriptors, contributions
			},
			want: "do not match active schema",
		},
		{
			name: "backend capability mismatch",
			edit: func(descriptors []projections.ProviderDescriptor, contributions []QueryContribution) ([]projections.ProviderDescriptor, []QueryContribution) {
				contributions[0].BackendKind = QueryBackendSourceOwner
				return descriptors, contributions
			},
			want: "does not match descriptor capability backend",
		},
		{
			name: "nil provider",
			edit: func(descriptors []projections.ProviderDescriptor, contributions []QueryContribution) ([]projections.ProviderDescriptor, []QueryContribution) {
				contributions[0].Provider = nil
				return descriptors, contributions
			},
			want: "has nil provider",
		},
		{
			name: "descriptor record type mismatch",
			edit: func(descriptors []projections.ProviderDescriptor, contributions []QueryContribution) ([]projections.ProviderDescriptor, []QueryContribution) {
				descriptors[0].SourceRecordTypes = []string{"wrong_record"}
				return descriptors, contributions
			},
			want: "do not match projection descriptor",
		},
		{
			name: "duplicate descriptor view",
			edit: func(descriptors []projections.ProviderDescriptor, contributions []QueryContribution) ([]projections.ProviderDescriptor, []QueryContribution) {
				duplicate := descriptors[0]
				duplicate.ProviderKey += "-duplicate"
				return append(descriptors, duplicate), contributions
			},
			want: "duplicate active projection descriptor ownership",
		},
		{
			name: "missing descriptor",
			edit: func(descriptors []projections.ProviderDescriptor, contributions []QueryContribution) ([]projections.ProviderDescriptor, []QueryContribution) {
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
			_, err := NewWorkbookContributionCatalog(
				testDescriptors,
				testContributions,
				createContributions,
				patchContributions,
				conflictContributions,
			)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}
}

func TestWorkbookContributionCatalogValidatesCreateAndPatchRequirements_Unit(t *testing.T) {
	descriptors, queries, creates, patches, conflicts := validCatalogInputs(t)
	catalog, err := NewWorkbookContributionCatalog(
		descriptors,
		queries,
		creates,
		patches,
		conflicts,
	)
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
	}
	for _, test := range createTests {
		t.Run("create/"+test.name, func(t *testing.T) {
			_, err := NewWorkbookContributionCatalog(
				descriptors,
				queries,
				test.edit(cloneCreateContributions(creates)),
				patches,
				conflicts,
			)
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
	}
	for _, test := range patchTests {
		t.Run("patch/"+test.name, func(t *testing.T) {
			_, err := NewWorkbookContributionCatalog(
				descriptors,
				queries,
				creates,
				test.edit(clonePatchContributions(patches)),
				conflicts,
			)
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
	}
	for _, test := range conflictTests {
		t.Run("conflict/"+test.name, func(t *testing.T) {
			_, err := NewWorkbookContributionCatalog(
				descriptors,
				queries,
				creates,
				patches,
				test.edit(cloneConflictContributions(conflicts)),
			)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}
}

func validCatalogInputs(t testing.TB) (
	[]projections.ProviderDescriptor,
	[]QueryContribution,
	[]CreateContribution,
	[]PatchContribution,
	[]ConflictContribution,
) {
	t.Helper()
	resources := viewschema.ListPublicResources()
	descriptors := make([]projections.ProviderDescriptor, 0, len(resources))
	contributions := make([]QueryContribution, 0, len(resources))
	creates := make([]CreateContribution, 0, len(resources))
	createProvider := createProvider{
		decode: func(io.Reader) (CreateAdmission, *httpapi.APIError) {
			return CreateAdmission{}, nil
		},
		create: func(context.Context, CreateCommand) (MutationResult, error) {
			return MutationResult{}, nil
		},
	}
	for _, resource := range resources {
		ownerKey := "owner:" + resource.ViewSchemaID
		descriptors = append(descriptors, projections.ProviderDescriptor{
			Status:            projections.ProviderStatusActive,
			ProviderKey:       "provider:" + resource.ViewSchemaID,
			SourceOwnerKey:    ownerKey,
			ViewSchemaIDs:     []string{resource.ViewSchemaID},
			SourceRecordTypes: append([]string(nil), resource.SourceRecordTypes...),
			Capabilities:      projections.ProviderCapabilities{Query: true},
		})
		contributions = append(contributions, QueryContribution{
			ViewSchemaID:      resource.ViewSchemaID,
			SourceOwnerKey:    ownerKey,
			SourceRecordTypes: append([]string(nil), resource.SourceRecordTypes...),
			BackendKind:       QueryBackendProjection,
			Provider: QueryProviderFunc(
				func(context.Context, QueryCommand) (querypage.Result, error) {
					return querypage.Result{}, nil
				},
			),
		})
		schema, _ := viewschema.Lookup(resource.ViewSchemaID)
		if schema.CreateCapable {
			creates = append(creates, CreateContribution{
				ViewSchemaID:      resource.ViewSchemaID,
				SourceOwnerKey:    ownerKey,
				SourceRecordTypes: append([]string(nil), resource.SourceRecordTypes...),
				Provider:          createProvider,
			})
		}
	}
	patchProvider := patchProvider{
		decode: func(io.Reader) (PatchAdmission, *httpapi.APIError) {
			return PatchAdmission{}, nil
		},
		patch: func(context.Context, PatchCommand) (MutationResult, error) {
			return MutationResult{}, nil
		},
	}
	conflictProvider := conflictProvider{
		decode: func(
			io.Reader,
			string,
			conflicttokens.ConflictTokenClaims,
		) (ConflictAdmission, *httpapi.APIError) {
			return ConflictAdmission{}, nil
		},
		resolve: func(context.Context, ConflictCommand) (MutationResult, error) {
			return MutationResult{}, nil
		},
	}
	patches := make([]PatchContribution, 0, len(expectedPatchSurfaces()))
	for recordType, viewSchemaIDs := range expectedPatchSurfaces() {
		patches = append(patches, PatchContribution{
			RecordType:    recordType,
			ViewSchemaIDs: append([]string(nil), viewSchemaIDs...),
			Provider:      patchProvider,
		})
	}
	conflicts := make([]ConflictContribution, 0, len(expectedConflictSurfaces()))
	for recordType, viewSchemaIDs := range expectedConflictSurfaces() {
		conflicts = append(conflicts, ConflictContribution{
			RecordType:    recordType,
			ViewSchemaIDs: append([]string(nil), viewSchemaIDs...),
			Provider:      conflictProvider,
		})
	}
	return descriptors, contributions, creates, patches, conflicts
}

func cloneProviderDescriptors(input []projections.ProviderDescriptor) []projections.ProviderDescriptor {
	cloned := append([]projections.ProviderDescriptor(nil), input...)
	for index := range cloned {
		cloned[index].ViewSchemaIDs = append([]string(nil), cloned[index].ViewSchemaIDs...)
		cloned[index].SourceRecordTypes = append([]string(nil), cloned[index].SourceRecordTypes...)
	}
	return cloned
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
