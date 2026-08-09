package workbookprojection

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

const timelineViewSchemaID = "cartulary.view.timeline.v2"

type SourceReader interface {
	BuildProjectionMutationTx(context.Context, pgx.Tx, uuid.UUID) (ProjectionMutation, error)
	ListProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (ProjectionInputPage, error)
}

type Writer interface {
	ApplyTimelineMutationTx(context.Context, pgx.Tx, ProjectionMutation) error
}

type Rebuilder interface {
	RebuildTimeline(context.Context, uuid.UUID) error
}

type Ports struct {
	Writer    Writer
	Rebuilder Rebuilder
}

type Contribution struct {
	contract providercontract.Contribution
	source   SourceReader
}

func NewContribution(source SourceReader) (Contribution, error) {
	if source == nil {
		return Contribution{}, fmt.Errorf("timeline projection source is required")
	}
	contract, err := providercontract.NewContribution(
		"timeline",
		[]providercontract.ProviderDescriptor{Descriptor()},
		[]providercontract.SurfaceIntent{SurfaceIntent()},
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

func Descriptor() providercontract.ProviderDescriptor {
	return providercontract.ProviderDescriptor{
		SchemaVersion:                providercontract.DescriptorSchemaVersion,
		Status:                       providercontract.ProviderStatusActive,
		ProviderID:                   "timeline",
		SourceOwnerModule:            "timeline",
		ViewSchemaIDs:                []string{timelineViewSchemaID},
		SourceRecordTypes:            []string{"timeline_event"},
		SourceAuthorityModules:       []string{"entities", "evidence", "links", "records", "timeline"},
		ProjectionTableIDs:           []string{"timeline_grid_projection"},
		ProjectionStorageOwnerModule: "projections",
		Capabilities: providercontract.ProviderCapabilities{
			Query:           true,
			RefreshRow:      true,
			RestoreRebuild:  true,
			IncidentRebuild: true,
		},
		RestoreRebuild:       providercontract.RestoreRebuildRequired,
		FacadePackages:       []string{"internal/modules/timeline/workbookprojection"},
		CharacterizationRefs: []string{"internal/modules/timeline/projection_contract_test.go"},
	}
}

func SurfaceIntent() providercontract.SurfaceIntent {
	return providercontract.SurfaceIntent{
		ViewSchemaID: timelineViewSchemaID,
		FieldKeys: []string{
			"timeline.date_entered_text",
			"timeline.analyst_text",
			"timeline.mitre_stage_text",
			"timeline.device_object_text",
			"timeline.ip_address_text",
			"timeline.activity_utc_text",
			"timeline.activity_local_text",
			"timeline.raw_activity_text",
			"timeline.activity_synopsis_text",
			"timeline.data_source_text",
			"timeline.host_refs",
			"timeline.identity_refs",
			"timeline.tags",
			"timeline.attached_evidence_ids",
			"timeline.evidence_count",
			"timeline.recorded_at",
			"timeline.edited_at",
			"timeline.activity_sort_ts",
			"timeline.date_entered_sort_day",
			"timeline.activity_time_pair_state",
			"timeline.capture_state",
			"timeline.replacement_record_id",
			"timeline.has_evidence",
			"timeline.has_unresolved_mentions",
		},
	}
}
