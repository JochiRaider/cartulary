package workbookprojection

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	hostViewSchemaID     = "cartulary.view.hosts.v1"
	identityViewSchemaID = "cartulary.view.identities.v1"
)

// Writer is the complete typed Entities mutation boundary for derived workbook
// rows. Transactions remain owned by the caller.
type Writer interface {
	RefreshHostTx(context.Context, pgx.Tx, uuid.UUID) error
	RefreshIdentityTx(context.Context, pgx.Tx, uuid.UUID) error
	DeleteHostTx(context.Context, pgx.Tx, uuid.UUID) error
	DeleteIdentityTx(context.Context, pgx.Tx, uuid.UUID) error
	RebuildHostsTx(context.Context, pgx.Tx, uuid.UUID) error
	RebuildIdentitiesTx(context.Context, pgx.Tx, uuid.UUID) error
}

type Rebuilder interface {
	RebuildHosts(context.Context, uuid.UUID) error
	RebuildIdentities(context.Context, uuid.UUID) error
}

// HostProjectionInput is the source-owned, typed materialization input for one
// host projection row. It contains semantic values and no physical table or
// query-plan details.
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

// HostQueryProjection is the bounded derived-state selection returned before
// the Entities owner hydrates authoritative host and record facts by exact ID.
type HostQueryProjection struct {
	RecordID          uuid.UUID
	HostState         string
	LinkedEventCount  int
	EvidenceCount     int
	Location          *string
	OSPlatform        *string
	BusinessOwner     *string
	Criticality       *string
	ContainmentStatus *string
}

type IdentityQueryProjection struct {
	RecordID         uuid.UUID
	IdentityState    string
	LinkedEventCount int
	EvidenceCount    int
	PrivilegeLevel   *string
	MFAState         *string
	ResetStatus      *string
}

type DerivedFact struct {
	RecordID     uuid.UUID
	RecordType   string
	ContentClass string
	Value        map[string]any
}

type Reader interface {
	SelectHostQueryProjections(context.Context, uuid.UUID, viewschema.QueryMeta, querypage.Window) ([]HostQueryProjection, error)
	SelectIdentityQueryProjections(context.Context, uuid.UUID, viewschema.QueryMeta, querypage.Window) ([]IdentityQueryProjection, error)
	CollectHostDerivedFactsTx(context.Context, pgx.Tx, uuid.UUID) ([]DerivedFact, error)
	CollectIdentityDerivedFactsTx(context.Context, pgx.Tx, uuid.UUID) ([]DerivedFact, error)
}

type Ports struct {
	Writer    Writer
	Rebuilder Rebuilder
	Reader    Reader
}

func (ports Ports) Ready() bool {
	return ports.Writer != nil && ports.Rebuilder != nil && ports.Reader != nil
}

type Contribution struct {
	contract providercontract.Contribution
	source   SourceReader
}

func NewContribution(
	descriptors []providercontract.ProviderDescriptor,
	intents []providercontract.SurfaceIntent,
	sources ...SourceReader,
) (Contribution, error) {
	if len(sources) > 1 {
		return Contribution{}, fmt.Errorf("entities projection contribution accepts at most one source")
	}
	contract, err := providercontract.NewContribution("entities", descriptors, intents)
	if err != nil {
		return Contribution{}, err
	}
	var source SourceReader
	if len(sources) == 1 {
		source = sources[0]
	}
	return Contribution{contract: contract, source: source}, nil
}

func NewRuntimeContribution(source SourceReader) (Contribution, error) {
	if source == nil {
		return Contribution{}, fmt.Errorf("entities projection source is required")
	}
	return NewContribution(
		Descriptors(),
		[]providercontract.SurfaceIntent{HostSurfaceIntent(), IdentitySurfaceIntent()},
		source,
	)
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
	return []providercontract.ProviderDescriptor{
		host,
		identity,
	}
}

func IdentitySurfaceIntent() providercontract.SurfaceIntent {
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

func HostSurfaceIntent() providercontract.SurfaceIntent {
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
		RestoreRebuild:       providercontract.RestoreRebuildRequired,
		FacadePackages:       []string{"internal/modules/entities/workbookprojection"},
		RebuildAfter:         rebuildAfter,
		CharacterizationRefs: []string{"internal/modules/entities/resolution_integration_test.go"},
	}
}
