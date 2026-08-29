package server

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func TestBuildCollaborationPublicationCatalogJoinsIndependentAuthorities(t *testing.T) {
	descriptor := collaborationPublicationTestDescriptor()
	contribution := collaborationPublicationTestContribution()

	tests := []struct {
		name         string
		descriptor   providercontract.ProviderDescriptor
		contribution revisions.ProviderContribution
		want         string
	}{
		{name: "complete", descriptor: descriptor, contribution: contribution},
		{name: "descriptor resource record mismatch", descriptor: mutateCollaborationPublicationDescriptor(descriptor, func(value *providercontract.ProviderDescriptor) {
			value.SourceRecordTypes = []string{"private"}
		}), contribution: contribution, want: "record types disagree"},
		{name: "route record mismatch", descriptor: descriptor, contribution: mutateCollaborationPublicationContribution(contribution, func(value *revisions.ProviderContribution) {
			value.Records[0].RecordType = "note"
		}), want: "record types do not match"},
		{name: "route owner mismatch", descriptor: descriptor, contribution: mutateCollaborationPublicationContribution(contribution, func(value *revisions.ProviderContribution) {
			value.SourceOwnerModule = revisions.SourceOwnerParties
		}), want: "belongs to source owner \"artifacts\""},
		{name: "missing route", descriptor: descriptor, contribution: mutateCollaborationPublicationContribution(contribution, func(value *revisions.ProviderContribution) {
			value.Records[0].RecordViewRoutes = nil
		}), want: "is incomplete"},
		{name: "unknown route view", descriptor: descriptor, contribution: mutateCollaborationPublicationContribution(contribution, func(value *revisions.ProviderContribution) {
			value.Records[0].RecordViewRoutes[0].ViewSchemaIDs[0] = "cartulary.view.unknown.v1"
		}), want: "unknown view schema"},
		{name: "unknown descriptor view", descriptor: mutateCollaborationPublicationDescriptor(descriptor, func(value *providercontract.ProviderDescriptor) {
			value.ViewSchemaIDs = []string{"cartulary.view.unknown.v1"}
		}), contribution: contribution, want: "has unknown view schema"},
		{name: "inactive descriptor", descriptor: mutateCollaborationPublicationDescriptor(descriptor, func(value *providercontract.ProviderDescriptor) {
			value.Status = providercontract.ProviderStatusExperimental
			value.RestoreRebuild = providercontract.RestoreRebuildUnsupported
			value.Capabilities.RestoreRebuild = false
		}), contribution: contribution, want: "active projection descriptor set is empty"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			descriptors, err := providercontract.NewDescriptorSet([]providercontract.ProviderDescriptor{test.descriptor})
			if err != nil {
				t.Fatalf("construct descriptor set: %v", err)
			}
			catalog, err := buildCollaborationPublicationCatalog(
				[]revisions.ProviderContribution{test.contribution},
				descriptors,
			)
			if test.want == "" {
				if err != nil || catalog == nil {
					t.Fatalf("catalog = %#v, error = %v", catalog, err)
				}
				return
			}
			if err == nil || catalog != nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("catalog = %#v, error = %v; want error containing %q", catalog, err, test.want)
			}
		})
	}
}

func collaborationPublicationTestDescriptor() providercontract.ProviderDescriptor {
	return providercontract.ProviderDescriptor{
		SchemaVersion:                providercontract.DescriptorSchemaVersion,
		Status:                       providercontract.ProviderStatusActive,
		ProviderID:                   "artifact",
		SourceOwnerModule:            "artifacts",
		ViewSchemaIDs:                []string{"cartulary.view.notes.v1"},
		SourceRecordTypes:            []string{"artifact"},
		SourceAuthorityModules:       []string{"artifacts"},
		ProjectionTableIDs:           []string{"artifact_grid_projection"},
		ProjectionStorageOwnerModule: "projections",
		Capabilities: providercontract.ProviderCapabilities{
			Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true,
		},
		RestoreRebuild: providercontract.RestoreRebuildRequired,
		FacadePackages: []string{"internal/modules/artifacts"},
	}
}

func collaborationPublicationTestContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerArtifacts,
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule: revisions.SourceOwnerArtifacts,
			RecordType:        "artifact",
			RecordViewRoutes: []revisions.RecordViewRouteContribution{{
				ContributionID: "artifacts.views.v1",
				ViewSchemaIDs:  []string{"cartulary.view.notes.v1"},
			}},
		}},
	}
}

func mutateCollaborationPublicationDescriptor(
	descriptor providercontract.ProviderDescriptor,
	mutate func(*providercontract.ProviderDescriptor),
) providercontract.ProviderDescriptor {
	descriptor = descriptor.Clone()
	mutate(&descriptor)
	return descriptor
}

func mutateCollaborationPublicationContribution(
	contribution revisions.ProviderContribution,
	mutate func(*revisions.ProviderContribution),
) revisions.ProviderContribution {
	contribution.Records = append([]revisions.RecordProviderContribution(nil), contribution.Records...)
	for index := range contribution.Records {
		contribution.Records[index].RecordViewRoutes = append(
			[]revisions.RecordViewRouteContribution(nil),
			contribution.Records[index].RecordViewRoutes...,
		)
		for routeIndex := range contribution.Records[index].RecordViewRoutes {
			contribution.Records[index].RecordViewRoutes[routeIndex].ViewSchemaIDs = append(
				[]string(nil),
				contribution.Records[index].RecordViewRoutes[routeIndex].ViewSchemaIDs...,
			)
		}
	}
	mutate(&contribution)
	return contribution
}
