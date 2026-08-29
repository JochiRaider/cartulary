package projectioncontract

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

const (
	hostViewSchemaID     = "cartulary.view.hosts.v1"
	identityViewSchemaID = "cartulary.view.identities.v1"
)

// HostProjectionInput is the source-owned materialization input for one host
// projection row. It contains semantic values and no storage-plan details.
type HostProjectionInput struct {
	RecordID          uuid.UUID
	IncidentID        uuid.UUID
	RowVersion        int64
	DisplayName       string
	Hostname          *string
	HostState         string
	LinkedEventCount  int
	EvidenceCount     int
	Location          *string
	OSPlatform        *string
	BusinessOwner     *string
	Criticality       *string
	ContainmentStatus *string
	EditedAt          time.Time
}

type HostProjectionPage struct {
	Inputs       []HostProjectionInput
	NextRecordID *uuid.UUID
}

type IdentityProjectionInput struct {
	RecordID         uuid.UUID
	IncidentID       uuid.UUID
	RowVersion       int64
	DisplayName      string
	UPN              *string
	Email            *string
	SamAccountName   *string
	IdentityState    string
	LinkedEventCount int
	EvidenceCount    int
	PrivilegeLevel   *string
	MFAState         *string
	ResetStatus      *string
	EditedAt         time.Time
}

type IdentityProjectionPage struct {
	Inputs       []IdentityProjectionInput
	NextRecordID *uuid.UUID
}

type SourceReader interface {
	LoadHostProjectionInputTx(context.Context, pgx.Tx, uuid.UUID) (HostProjectionInput, bool, error)
	ListHostProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (HostProjectionPage, error)
	LoadIdentityProjectionInputTx(context.Context, pgx.Tx, uuid.UUID) (IdentityProjectionInput, bool, error)
	ListIdentityProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (IdentityProjectionPage, error)
}

type Contribution struct {
	contract providercontract.Contribution
	source   SourceReader
}

func NewContribution(source SourceReader) (Contribution, error) {
	if source == nil {
		return Contribution{}, fmt.Errorf("entities projection source is required")
	}
	contract, err := providercontract.NewContribution(
		"entities",
		Descriptors(),
		[]providercontract.SurfaceIntent{hostSurfaceIntent(), identitySurfaceIntent()},
	)
	if err != nil {
		return Contribution{}, err
	}
	return Contribution{contract: contract, source: source}, nil
}

func (contribution Contribution) ProjectionContribution() providercontract.Contribution {
	return contribution.contract
}

func (contribution Contribution) Source() SourceReader {
	return contribution.source
}

func Descriptors() []providercontract.ProviderDescriptor {
	host := descriptor(
		"host",
		hostViewSchemaID,
		"host",
		"host_grid_projection",
		[]string{"timeline"},
	)
	host.Capabilities.Query = true
	identity := descriptor(
		"identity",
		identityViewSchemaID,
		"identity",
		"identity_grid_projection",
		[]string{"host"},
	)
	identity.Capabilities.Query = true
	return []providercontract.ProviderDescriptor{host, identity}
}

func identitySurfaceIntent() providercontract.SurfaceIntent {
	return providercontract.SurfaceIntent{
		ViewSchemaID: identityViewSchemaID,
		FieldKeys: []string{
			"identity.display_name",
			"identity.aad_object_id",
			"identity.sid",
			"identity.upn",
			"identity.email",
			"identity.sam_account_name",
			"identity.aliases",
			"identity.reusable_identifiers",
			"identity.identity_state",
			"identity.linked_event_count",
			"identity.evidence_count",
			"identity.privilege_level",
			"identity.mfa_state",
			"identity.reset_status",
			"identity.edited_at",
		},
	}
}

func hostSurfaceIntent() providercontract.SurfaceIntent {
	return providercontract.SurfaceIntent{
		ViewSchemaID: hostViewSchemaID,
		FieldKeys: []string{
			"host.display_name",
			"host.hostname",
			"host.aad_device_id",
			"host.fqdn",
			"host.aliases",
			"host.reusable_identifiers",
			"host.host_state",
			"host.linked_event_count",
			"host.evidence_count",
			"host.location",
			"host.os_platform",
			"host.business_owner",
			"host.criticality",
			"host.containment_status",
			"host.edited_at",
		},
	}
}

func descriptor(
	providerID string,
	viewSchemaID string,
	sourceRecordType string,
	tableID string,
	rebuildAfter []string,
) providercontract.ProviderDescriptor {
	return providercontract.ProviderDescriptor{
		SchemaVersion:                providercontract.DescriptorSchemaVersion,
		Status:                       providercontract.ProviderStatusActive,
		ProviderID:                   providerID,
		SourceOwnerModule:            "entities",
		ViewSchemaIDs:                []string{viewSchemaID},
		SourceRecordTypes:            []string{sourceRecordType},
		SourceAuthorityModules:       []string{"entities", "evidence", "links", "records"},
		ProjectionTableIDs:           []string{tableID},
		ProjectionStorageOwnerModule: "projections",
		Capabilities: providercontract.ProviderCapabilities{
			RefreshRow:      true,
			RestoreRebuild:  true,
			IncidentRebuild: true,
		},
		RestoreRebuild: providercontract.RestoreRebuildRequired,
		FacadePackages: []string{"internal/modules/entities/projectioncontract"},
		RebuildAfter:   rebuildAfter,
		CharacterizationRefs: []string{
			"internal/modules/entities/origin_upsert_integration_test.go",
			"internal/modules/entities/resolution_route_integration_test.go",
		},
	}
}
